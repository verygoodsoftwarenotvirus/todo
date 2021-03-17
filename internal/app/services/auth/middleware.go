package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/o1egl/paseto"
)

const (
	// userLoginInputMiddlewareCtxKey is the context key for login input.
	userLoginInputMiddlewareCtxKey types.ContextKey = "user_login_input"
	// pasetoCreationInputMiddlewareCtxKey is the context key for login input.
	pasetoCreationInputMiddlewareCtxKey types.ContextKey = "paseto_creation_input"
	// changeActiveAccountMiddlewareCtxKey is the context key for login input.
	changeActiveAccountMiddlewareCtxKey types.ContextKey = "change_active_account"

	signatureHeaderKey     = "Signature"
	pasetoAuthorizationKey = "Authorization"

	// usernameFormKey is the string we look for in request forms for username information.
	usernameFormKey = "username"
	// passwordFormKey is the string we look for in request forms for authentication information.
	passwordFormKey = "password"
	// totpTokenFormKey is the string we look for in request forms for TOTP token information.
	totpTokenFormKey = "totpToken"
)

// parseLoginInputFromForm checks a request for a login form, and returns the parsed login data if relevant.
func parseLoginInputFromForm(req *http.Request) *types.UserLoginInput {
	if err := req.ParseForm(); err == nil {
		input := &types.UserLoginInput{
			Username:  req.FormValue(usernameFormKey),
			Password:  req.FormValue(passwordFormKey),
			TOTPToken: req.FormValue(totpTokenFormKey),
		}

		if input.Username != "" && input.Password != "" && input.TOTPToken != "" {
			return input
		}
	}

	return nil
}

func (s *service) checkRequestForToken(ctx context.Context, req *http.Request) (*types.RequestContext, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	if rawToken := req.Header.Get(pasetoAuthorizationKey); rawToken != "" {
		var token paseto.JSONToken

		if decryptErr := paseto.NewV2().Decrypt(rawToken, s.config.PASETO.LocalModeKey, &token, nil); decryptErr != nil {
			logger.Error(decryptErr, "error decrypting PASETO")
			return nil, decryptErr
		}

		if time.Now().UTC().After(token.Expiration) {
			return nil, errors.New("token expired")
		}

		gobEncoded, base64DecodeErr := base64.RawURLEncoding.DecodeString(token.Get(pasetoDataKey))
		if base64DecodeErr != nil {
			logger.Error(base64DecodeErr, "error decoding base64 encoded GOB payload")
			return nil, base64DecodeErr
		}

		var reqContext *types.RequestContext

		gobDecodeErr := gob.NewDecoder(bytes.NewReader(gobEncoded)).Decode(&reqContext)
		if gobDecodeErr != nil {
			logger.Error(gobDecodeErr, "error decoding GOB encoded session info payload")
			return nil, gobDecodeErr
		}

		return reqContext, nil
	}

	return nil, errors.New("no token data found")
}

// CookieAuthenticationMiddleware checks every request for a user cookie.
func (s *service) CookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// fetch the user from the request.
		user, err := s.fetchUserFromCookie(ctx, req)
		if err != nil {
			// we deliberately aren't logging here because it's done in fetchUserFromCookie
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "cookie required", http.StatusUnauthorized)
			return
		}

		if user != nil {
			defaultAccount, userPermissions, membershipRetrievalErr := s.accountMembershipManager.GetMembershipsForUser(ctx, user.ID)
			if membershipRetrievalErr != nil {
				logger.Error(membershipRetrievalErr, "retrieving userPermissions for cookie authentication")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			reqCtx, requestContextErr := types.RequestContextFromUser(user, defaultAccount, userPermissions)
			if requestContextErr != nil {
				logger.Error(membershipRetrievalErr, "forming request context")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.UserIDContextKey, reqCtx)))
			return
		}

		// if no error was attached to the request, tell them to login first.
		http.Redirect(res, req, "/auth/login", http.StatusUnauthorized)
	})
}

// UserAttributionMiddleware is concerned with figuring out who a user is, but not worried about kicking out users who are not known.
func (s *service) UserAttributionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// handle cookies if relevant.
		if cookieContext, userID, err := s.getUserIDFromCookie(ctx, req); err == nil && userID != 0 {
			tracing.AttachUserIDToSpan(span, userID)
			ctx = cookieContext

			logger = logger.WithValue(keys.RequesterKey, userID)

			reqCtx, userIsBannedErr := s.userDataManager.GetRequestContextForUser(ctx, userID)
			if userIsBannedErr != nil {
				logger.Error(userIsBannedErr, "fetching user info for cookie")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			if activeAccount, ok := s.sessionManager.Get(ctx, string(types.AccountIDContextKey)).(uint64); ok {
				reqCtx.User.ActiveAccountID = activeAccount
			}

			ctx = context.WithValue(ctx, types.RequestContextKey, reqCtx)

			next.ServeHTTP(res, req.WithContext(ctx))
			return
		}

		tokenRequestContext, tokenErr := s.checkRequestForToken(ctx, req)
		if tokenErr != nil {
			logger.Error(tokenErr, "error extracting token from request")
			s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
			return
		}

		if tokenRequestContext != nil {
			// no need to fetch info since tokens are so short-lived.
			next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.RequestContextKey, tokenRequestContext)))
			return
		}

		next.ServeHTTP(res, req)
	})
}

// AuthorizationMiddleware checks to see if a user is associated with the request, and then determines whether said request can proceed.
func (s *service) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// UserAttributionMiddleware should be called before this middleware.
		if reqCtx, ok := ctx.Value(types.RequestContextKey).(*types.RequestContext); ok {
			// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere.
			if reqCtx.User.Status == types.BannedAccountStatus {
				logger.Debug("banned user attempted to make request")
				http.Redirect(res, req, "/", http.StatusForbidden)
				return
			}

			if _, authorizedForAccount := reqCtx.User.AccountPermissionsMap[reqCtx.User.ActiveAccountID]; !authorizedForAccount {
				logger.Debug("user trying to access account they are not authorized for")
				http.Redirect(res, req, "/", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(res, req)
			return
		}

		logger.Debug("no user attached to request")
		http.Redirect(res, req, "/auth/login", http.StatusUnauthorized)
	})
}

// PermissionRestrictionMiddleware is concerned with figuring otu who a user is, but not worried about kicking out users who are not known.
func (s *service) PermissionRestrictionMiddleware(perms ...permissions.ServiceUserPermissions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx, span := s.tracer.StartSpan(req.Context())
			defer span.End()

			logger := s.logger.WithRequest(req)

			// check for a cookie first if we can.
			requestContext, ok := ctx.Value(types.RequestContextKey).(*types.RequestContext)
			if !ok {
				logger.Debug("no request context attached!")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			if requestContext.User.ServiceAdminPermissions != 0 {
				logger.Debug("allowing admin user!")
				next.ServeHTTP(res, req)
				return
			}

			accountPermissions, allowed := requestContext.User.AccountPermissionsMap[requestContext.User.ActiveAccountID]
			if !allowed {
				logger.Debug("not authorized for account!")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			logger = logger.WithValue(keys.RequesterKey, requestContext.User.ID).
				WithValue(keys.AccountIDKey, requestContext.User.ActiveAccountID).
				WithValue(keys.PermissionsKey, requestContext.User.AccountPermissionsMap)

			for _, p := range perms {
				if !accountPermissions.HasPermission(p) {
					logger.WithValue("requested_permission", p).Debug("inadequate permissions")
					s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
					return
				}
			}

			next.ServeHTTP(res, req)
		})
	}
}

// AdminMiddleware restricts requests to admin users only.
func (s *service) AdminMiddleware(next http.Handler) http.Handler {
	const staticError = "admin status required"

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		reqCtx, ok := ctx.Value(types.RequestContextKey).(*types.RequestContext)

		if !ok || reqCtx == nil {
			logger.Debug("AdminMiddleware called without user attached to context")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

		if !reqCtx.User.ServiceAdminPermissions.IsServiceAdmin() {
			logger.Debug("AdminMiddleware called by non-admin user")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(res, req)
	})
}

// ChangeActiveAccountInputMiddleware fetches user login input from requests.
func (s *service) ChangeActiveAccountInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.ChangeActiveAccountInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.Error(err, "provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, changeActiveAccountMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// UserLoginInputMiddleware fetches user login input from requests.
func (s *service) UserLoginInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		logger.Debug("UserLoginInputMiddleware called")

		x := new(types.UserLoginInput)
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			if x = parseLoginInputFromForm(req); x == nil {
				logger.Error(err, "error encountered decoding request body")
				s.encoderDecoder.EncodeErrorResponse(ctx, res, "attached input is invalid", http.StatusBadRequest)
				return
			}
		}

		if err := x.Validate(ctx, s.config.MinimumUsernameLength, s.config.MinimumPasswordLength); err != nil {
			logger.Error(err, "provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PASETOCreationInputMiddleware fetches user login input from requests.
func (s *service) PASETOCreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		logger.Debug("PASETOCreationInputMiddleware called")

		x := new(types.PASETOCreationInput)
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "attached input is invalid", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.Error(err, "provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, pasetoCreationInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

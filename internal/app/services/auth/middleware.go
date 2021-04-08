package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
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

var (
	errTokenExpired  = errors.New("token expired")
	errTokenNotFound = errors.New("no token data found")
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

func (s *service) fetchRequestContextFromPASETO(ctx context.Context, req *http.Request) (*types.RequestContext, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	if rawToken := req.Header.Get(pasetoAuthorizationKey); rawToken != "" {
		var token paseto.JSONToken

		if err := paseto.NewV2().Decrypt(rawToken, s.config.PASETO.LocalModeKey, &token, nil); err != nil {
			return nil, observability.PrepareError(err, logger, span, "decrypting PASETO")
		}

		if time.Now().UTC().After(token.Expiration) {
			return nil, errTokenExpired
		}

		gobEncoded, err := base64.RawURLEncoding.DecodeString(token.Get(pasetoDataKey))
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "decoding base64 encoded GOB payload")
		}

		var reqContext *types.RequestContext

		if err = gob.NewDecoder(bytes.NewReader(gobEncoded)).Decode(&reqContext); err != nil {
			return nil, observability.PrepareError(err, logger, span, "decoding GOB encoded session info payload")
		}

		logger.WithValue("active_account_id", reqContext.ActiveAccountID).Debug("returning request context")

		return reqContext, nil
	}

	return nil, errTokenNotFound
}

// CookieAuthenticationMiddleware requires every request have a valid cookie.
func (s *service) CookieRequirementMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		if cookie, cookieErr := req.Cookie(s.config.Cookies.Name); !errors.Is(cookieErr, http.ErrNoCookie) && cookie != nil {
			var token string
			if err := s.cookieManager.Decode(s.config.Cookies.Name, cookie.Value, &token); err == nil {
				next.ServeHTTP(res, req)
			}
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
			ctx = cookieContext

			tracing.AttachRequestingUserIDToSpan(span, userID)
			logger = logger.WithValue(keys.RequesterIDKey, userID)

			reqCtx, reqCtxErr := s.accountMembershipManager.BuildRequestContextForUser(ctx, userID)
			if reqCtxErr != nil {
				observability.AcknowledgeError(reqCtxErr, logger, span, "fetching user info for cookie")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			s.overrideRequestContextValuesWithSessionData(ctx, reqCtx)

			next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.RequestContextKey, reqCtx)))
			return
		}

		tokenRequestContext, err := s.fetchRequestContextFromPASETO(ctx, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "extracting token from request")
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
		_, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// UserAttributionMiddleware should be called before this middleware.
		if reqCtx, err := s.requestContextFetcher(req); err == nil && reqCtx != nil {
			logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID)

			// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere.
			if reqCtx.Requester.Reputation == types.BannedAccountStatus {
				logger.Debug("banned user attempted to make request")
				http.Redirect(res, req, "/", http.StatusForbidden)
				return
			}

			var authorizedAccounts []uint64
			for k := range reqCtx.AccountPermissionsMap {
				authorizedAccounts = append(authorizedAccounts, k)
			}

			logger = logger.WithValue("requested_account_id", reqCtx.ActiveAccountID).
				WithValue("authorized_accounts", authorizedAccounts)

			if _, authorizedForAccount := reqCtx.AccountPermissionsMap[reqCtx.ActiveAccountID]; !authorizedForAccount {
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

			// check for a request context first.
			requestContext, err := s.requestContextFetcher(req)
			if err != nil {
				observability.AcknowledgeError(err, logger, span, "retrieving request context")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			if requestContext.Requester.ServiceAdminPermissions != 0 {
				logger.Debug("allowing admin user!")
				next.ServeHTTP(res, req)
				return
			}

			accountMemberships, allowed := requestContext.AccountPermissionsMap[requestContext.ActiveAccountID]
			if !allowed {
				logger.Debug("not authorized for account!")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			logger = logger.WithValue(keys.RequesterIDKey, requestContext.Requester.ID).
				WithValue(keys.AccountIDKey, requestContext.ActiveAccountID).
				WithValue(keys.PermissionsKey, requestContext.AccountPermissionsMap)

			for _, p := range perms {
				if !accountMemberships.Permissions.HasPermission(p) {
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

		reqCtx, err := s.requestContextFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "retrieving request context")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		logger = logger.WithValue(keys.RequesterIDKey, reqCtx.Requester.ID)

		if !reqCtx.Requester.ServiceAdminPermissions.IsServiceAdmin() {
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
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.WithValue("validation_error", err).Debug("invalid input attached to request")
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

		x := new(types.UserLoginInput)
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			if x = parseLoginInputFromForm(req); x == nil {
				observability.AcknowledgeError(err, logger, span, "decoding request body")
				s.encoderDecoder.EncodeErrorResponse(ctx, res, "attached input is invalid", http.StatusBadRequest)
				return
			}
		}

		if err := x.Validate(ctx, s.config.MinimumUsernameLength, s.config.MinimumPasswordLength); err != nil {
			observability.AcknowledgeError(err, logger, span, "validating input")
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

		x := new(types.PASETOCreationInput)
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "attached input is invalid", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.WithValue("validation_error", err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, pasetoCreationInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

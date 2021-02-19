package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/o1egl/paseto"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// userLoginInputMiddlewareCtxKey is the context key for login input.
	userLoginInputMiddlewareCtxKey types.ContextKey = "user_login_input"
	// pasetoCreationInputMiddlewareCtxKey is the context key for login input.
	pasetoCreationInputMiddlewareCtxKey types.ContextKey = "paseto_creation_input"

	signatureHeaderKey     = "Signature"
	pasetoAuthorizationKey = "Authorization"

	// usernameFormKey is the string we look for in request forms for username information.
	usernameFormKey = "username"
	// passwordFormKey is the string we look for in request forms for authentication information.
	passwordFormKey = "authentication"
	// totpTokenFormKey is the string we look for in request forms for TOTP token information.
	totpTokenFormKey = "totpToken"
)

// CookieAuthenticationMiddleware checks every request for a user cookie.
func (s *service) CookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		// fetch the user from the request.
		user, err := s.fetchUserFromCookie(ctx, req)
		if err != nil {
			// we deliberately aren't logging here because it's done in fetchUserFromCookie
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "cookie required", http.StatusUnauthorized)
			return
		}

		if user != nil {
			req = req.WithContext(
				context.WithValue(
					ctx,
					types.SessionInfoKey,
					types.SessionInfoFromUser(user),
				),
			)
			next.ServeHTTP(res, req)
			return
		}

		// if no error was attached to the request, tell them to login first.
		http.Redirect(res, req, "/auth/login", http.StatusUnauthorized)
	})
}

// UserAttributionMiddleware is concerned with figuring otu who a user is, but not worried about kicking out users who are not known.
func (s *service) UserAttributionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// check for a cookie first if we can.
		if cookieAuth, err := s.DecodeCookieFromRequest(ctx, req); err == nil && cookieAuth != nil {
			user, userRetrievalErr := s.userDB.GetUser(ctx, cookieAuth.UserID)
			if userRetrievalErr != nil {
				logger.Error(userRetrievalErr, "error authenticating request")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			logger.WithValue("user_found", user != nil).Debug("serving request")

			if user != nil {
				tracing.AttachUserIDToSpan(span, user.ID)
				next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.SessionInfoKey, types.SessionInfoFromUser(user))))
				return
			}
		}

		rawToken := req.Header.Get(pasetoAuthorizationKey)

		if rawToken != "" {
			var token paseto.JSONToken

			if decryptErr := paseto.NewV2().Decrypt(rawToken, s.config.PASETO.LocalModeKey, &token, nil); decryptErr != nil {
				logger.Error(decryptErr, "error decrypting PASETO")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			if time.Now().UTC().After(token.Expiration) {
				logger.WithValues(map[string]interface{}{
					"current_time": time.Now().UTC().String(),
					"token_expiry": token.Expiration.String(),
				}).Debug("PASETO expired")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			payload := token.Get(pasetoDataKey)

			gobEncoded, base64DecodeErr := base64.RawURLEncoding.DecodeString(payload)
			if base64DecodeErr != nil {
				logger.Error(base64DecodeErr, "error decoding base64 encoded GOB payload")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			var si *types.SessionInfo
			gobDecodeErr := gob.NewDecoder(bytes.NewReader(gobEncoded)).Decode(&si)
			if gobDecodeErr != nil {
				logger.Error(gobDecodeErr, "error decoding GOB encoded session info payload")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.SessionInfoKey, si)))
			return
		}

		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
	})
}

// AuthorizationMiddleware checks to see if a user is associated with the request, and then determines whether said request can proceed.
func (s *service) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// UserAttributionMiddleware should be called before this middleware.
		if si, ok := ctx.Value(types.SessionInfoKey).(*types.SessionInfo); ok {
			// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere.
			if si.UserAccountStatus == types.BannedAccountStatus {
				logger.Debug("banned user attempted to make request")
				http.Redirect(res, req, "/", http.StatusForbidden)
				return
			}

			next.ServeHTTP(res, req)
			return
		}

		logger.Debug("no user attached to request request")
		http.Redirect(res, req, "/auth/login", http.StatusUnauthorized)
	})
}

// AdminMiddleware restricts requests to admin users only.
func (s *service) AdminMiddleware(next http.Handler) http.Handler {
	const staticError = "admin status required"

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		si, ok := ctx.Value(types.SessionInfoKey).(*types.SessionInfo)

		if !ok || si == nil {
			logger.Debug("AdminMiddleware called without user attached to context")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		logger = logger.WithValue(keys.UserIDKey, si.UserID)

		if !si.ServiceAdminPermissions.IsServiceAdmin() {
			logger.Debug("AdminMiddleware called by non-admin user")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(res, req)
	})
}

// parseLoginInputFromForm checks a request for a login form, and returns the parsed login data if relevant.
func parseLoginInputFromForm(req *http.Request) *types.UserLoginInput {
	if err := req.ParseForm(); err == nil {
		uli := &types.UserLoginInput{
			Username:  req.FormValue(usernameFormKey),
			Password:  req.FormValue(passwordFormKey),
			TOTPToken: req.FormValue(totpTokenFormKey),
		}

		if uli.Username != "" && uli.Password != "" && uli.TOTPToken != "" {
			return uli
		}
	}

	return nil
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

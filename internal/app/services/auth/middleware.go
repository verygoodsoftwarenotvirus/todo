package auth

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// userLoginInputMiddlewareCtxKey is the context key for login input.
	userLoginInputMiddlewareCtxKey types.ContextKey = "user_login_input"

	// usernameFormKey is the string we look for in request forms for username information.
	usernameFormKey = "username"
	// passwordFormKey is the string we look for in request forms for password information.
	passwordFormKey = "password"
	// totpTokenFormKey is the string we look for in request forms for TOTP token information.
	totpTokenFormKey = "totpToken"
)

// CookieAuthenticationMiddleware checks every request for a user cookie.
func (s *service) CookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		// fetch the user from the request.
		user, err := s.fetchUserFromCookie(ctx, req)
		if err != nil {
			// we deliberately aren't logging here because it's done in fetchUserFromCookie
			s.encoderDecoder.EncodeErrorResponse(res, "cookie required", http.StatusUnauthorized)
			return
		}

		if user != nil {
			req = req.WithContext(
				context.WithValue(
					ctx,
					types.SessionInfoKey,
					user.ToSessionInfo(),
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
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		var user *types.User
		logger := s.logger.WithRequest(req)

		// check for a cookie first if we can.
		if cookieAuth, err := s.DecodeCookieFromRequest(ctx, req); err == nil && cookieAuth != nil {
			if user, err = s.userDB.GetUser(ctx, cookieAuth.UserID); err != nil {
				logger.Error(err, "error authenticating request")
				next.ServeHTTP(res, req)
				return
			}
		}

		logger.WithValue("user_found", user != nil).Debug("checked for cookie")

		// if the cookie wasn't present, or didn't indicate who the user is.
		if user == nil {
			// check to see if there is an OAuth2 token for a valid client attached to the request.
			// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
			if oauth2Client, err := s.oauth2ClientsService.ExtractOAuth2ClientFromRequest(ctx, req); err == nil && oauth2Client != nil {
				ctx = context.WithValue(ctx, types.OAuth2ClientKey, oauth2Client)
				tracing.AttachOAuth2ClientIDToSpan(span, oauth2Client.ClientID)

				logger.Debug("getting user")
				// attach the oauth2 client and user's info to the request.
				if user, err = s.userDB.GetUser(ctx, oauth2Client.BelongsToUser); err != nil {
					logger.Error(err, "error fetching user info for authentication")
				}
			}
		}

		if user != nil {
			tracing.AttachUserIDToSpan(span, user.ID)
			next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.SessionInfoKey, user.ToSessionInfo())))
			return
		}

		logger.Debug("serving request")
		next.ServeHTTP(res, req)
	})
}

// AuthorizationMiddleware checks to see if a user is associated with the request, and then determines whether said request can proceed.
func (s *service) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context())
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
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		si, ok := ctx.Value(types.SessionInfoKey).(*types.SessionInfo)

		if !ok || si == nil {
			logger.Debug("AdminMiddleware called without user attached to context")
			s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusUnauthorized)
			return
		}

		if !si.UserIsAdmin {
			logger.Debug("AdminMiddleware called by non-admin user")
			s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusUnauthorized)
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
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		logger.Debug("UserLoginInputMiddleware called")

		x := new(types.UserLoginInput)
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			if x = parseLoginInputFromForm(req); x == nil {
				logger.Error(err, "error encountered decoding request body")
				s.encoderDecoder.EncodeErrorResponse(res, "attached input is invalid", http.StatusBadRequest)
				return
			}
		}

		if err := x.Validate(ctx, s.config.MinimumUsernameLength, s.config.MinimumPasswordLength); err != nil {
			logger.Error(err, "provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

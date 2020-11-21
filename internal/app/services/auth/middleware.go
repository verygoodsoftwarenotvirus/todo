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
func (s *Service) CookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "auth.service.CookieAuthenticationMiddleware")
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

// AuthorizationMiddleware authenticates based on either an oauth2 token or a cookie.
func (s *Service) AuthorizationMiddleware(allowValidCookieInLieuOfAValidToken bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return s.authorizationMiddleware(allowValidCookieInLieuOfAValidToken, next)
	}
}

func (s *Service) authorizationMiddleware(allowCookies bool, next http.Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		const this = "auth.service.authorizationMiddleware"
		ctx, span := tracing.StartSpan(req.Context(), this)
		defer span.End()

		var (
			logger = s.logger.WithRequest(req).WithValue("source", this)
			user   *types.User
		)

		// check for a cookie first if we can.
		if allowCookies {
			if cookieAuth, err := s.DecodeCookieFromRequest(ctx, req); err == nil && cookieAuth != nil {
				user, err = s.userDB.GetUser(ctx, cookieAuth.UserID)
				if err != nil {
					logger.Error(err, "error authenticating request")
					http.Error(res, "fetching user", http.StatusInternalServerError)
					// if we get here, then we just don't have a valid cookie, and we need to move on.
					return
				}
			}

			logger.Debug("checked for cookie")
		}

		// if the cookie wasn't present, or didn't indicate who the user is.
		if user == nil {
			// check to see if there is an OAuth2 token for a valid client attached to the request.
			// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
			oauth2Client, err := s.oauth2ClientsService.ExtractOAuth2ClientFromRequest(ctx, req)
			if err != nil || oauth2Client == nil {
				logger.WithValue("oauth_client_is_nil", oauth2Client == nil).Error(err, "error authenticating request")
				http.Redirect(res, req, "/auth/login", http.StatusUnauthorized)
				return
			}

			// attach the oauth2 client and user's info to the request.
			ctx = context.WithValue(ctx, types.OAuth2ClientKey, oauth2Client)
			if user, err = s.userDB.GetUser(ctx, oauth2Client.BelongsToUser); err != nil {
				logger.Error(err, "error authenticating request")
				http.Error(res, "fetching user", http.StatusInternalServerError)
				return
			}
		}

		// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere.
		if user == nil {
			logger.Debug("no user attached to request request")
			http.Redirect(res, req, "/auth/login", http.StatusUnauthorized)
			return
		}

		if user.IsBanned() {
			logger.Debug("banned user attempted to make request")
			http.Redirect(res, req, "/", http.StatusForbidden)
			return
		}

		logger = logger.WithValue("user_is_admin", user.IsAdmin).
			WithValue("user_id", user.ID)

		logger.Debug("fetched user")
		ctx = context.WithValue(ctx, types.SessionInfoKey, user.ToSessionInfo())

		next.ServeHTTP(res, req.WithContext(ctx))
	}
}

// AuthenticationMiddleware is concerned with figuring otu who a user is, but not worried about kicking out users who are not known.
func (s *Service) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		const this = "auth.service.AuthenticationMiddleware"
		ctx, span := tracing.StartSpan(req.Context(), this)
		defer span.End()

		var user *types.User
		logger := s.logger.WithRequest(req).WithValue("source", this)

		// check for a cookie first if we can.
		if cookieAuth, err := s.DecodeCookieFromRequest(ctx, req); err == nil && cookieAuth != nil {
			user, err = s.userDB.GetUser(ctx, cookieAuth.UserID)
			if err != nil {
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
				user, err = s.userDB.GetUser(ctx, oauth2Client.BelongsToUser)
				if err != nil {
					logger.Error(err, "error fetching user info for authentication")
				}
				logger.Debug("got user")
			}
		}

		if user != nil {
			ctx = context.WithValue(ctx, types.SessionInfoKey, user.ToSessionInfo())
			logger.Debug("user determined")
			tracing.AttachUserIDToSpan(span, user.ID)
			next.ServeHTTP(res, req.WithContext(ctx))
			return
		}

		logger.Debug("serving request")
		next.ServeHTTP(res, req)
	})
}

// AdminMiddleware restricts requests to admin users only.
func (s *Service) AdminMiddleware(next http.Handler) http.Handler {
	const staticError = "admin status required"

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "auth.service.AdminMiddleware")
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
func (s *Service) UserLoginInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "auth.service.UserLoginInputMiddleware")
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

		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

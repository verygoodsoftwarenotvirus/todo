package auth

import (
	"context"
	"net/http"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

const (
	// UserLoginInputMiddlewareCtxKey is the context key for login input
	UserLoginInputMiddlewareCtxKey models.ContextKey = "user_login_input"

	// UsernameFormKey is the string we look for in request forms for username information
	UsernameFormKey = "username"
	// PasswordFormKey is the string we look for in request forms for password information
	PasswordFormKey = "password"
	// TOTPTokenFormKey is the string we look for in request forms for TOTP token information
	TOTPTokenFormKey = "totp_token"
)

// CookieAuthenticationMiddleware checks every request for a user cookie
func (s *Service) CookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "CookieAuthenticationMiddleware")
		defer span.End()

		// fetch the user from the request
		user, err := s.FetchUserFromRequest(ctx, req)
		if err != nil {
			s.logger.Error(err, "error encountered fetching user")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if user != nil {
			req = req.WithContext(
				context.WithValue(
					context.WithValue(ctx, models.UserKey, user),
					models.UserIDKey,
					user.ID,
				),
			)
			next.ServeHTTP(res, req)
			return
		}

		// if no error was attached to the request, tell them to login first
		http.Redirect(res, req, "/login", http.StatusUnauthorized)
	})
}

// AuthenticationMiddleware authenticates based on either an oauth2 token or a cookie
func (s *Service) AuthenticationMiddleware(allowValidCookieInLieuOfAValidToken bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx, span := trace.StartSpan(req.Context(), "AuthenticationMiddleware")
			defer span.End()

			// let's figure out who the user is
			var user *models.User

			// check for a cookie first if we can
			if allowValidCookieInLieuOfAValidToken {
				cookieAuth, err := s.DecodeCookieFromRequest(ctx, req)

				if err == nil && cookieAuth != nil {
					user, err = s.userDB.GetUser(ctx, cookieAuth.UserID)
					if err != nil {
						s.logger.Error(err, "error authenticating request")
						http.Error(res, "fetching user", http.StatusInternalServerError)
						// if we get here, then we just don't have a valid cookie, and we need to move on
						return
					}
				}
			}

			// if the cookie wasn't present, or didn't indicate who the user is
			if user == nil {
				// check to see if there is an OAuth2 token for a valid client attached to the request.
				// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
				oauth2Client, err := s.oauth2ClientsService.ExtractOAuth2ClientFromRequest(ctx, req)
				if err != nil || oauth2Client == nil {
					s.logger.Error(err, "fetching oauth2 client")
					http.Redirect(res, req, "/login", http.StatusUnauthorized)
					return
				}

				// attach the oauth2 client and user's info to the request
				ctx = context.WithValue(ctx, models.OAuth2ClientKey, oauth2Client)
				user, err = s.userDB.GetUser(ctx, oauth2Client.BelongsTo)
				if err != nil {
					s.logger.Error(err, "error authenticating request")
					http.Error(res, "fetching user", http.StatusInternalServerError)
					return
				}
			}

			// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere
			if user == nil {
				s.logger.Debug("no user attached to request request")
				http.Redirect(res, req, "/login", http.StatusUnauthorized)
				return
			}

			// elsewise, load the request with extra context
			ctx = context.WithValue(ctx, models.UserKey, user)
			ctx = context.WithValue(ctx, models.UserIDKey, user.ID)
			ctx = context.WithValue(ctx, models.UserIsAdminKey, user.IsAdmin)

			next.ServeHTTP(res, req.WithContext(ctx))
		})
	}
}

// AdminMiddleware restricts requests to admin users only
func (s *Service) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "AdminMiddleware")
		defer span.End()

		logger := s.logger.WithRequest(req)
		user, ok := ctx.Value(models.UserKey).(*models.User)

		if !ok || user == nil {
			logger.Debug("AdminMiddleware called without user attached to context")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin {
			logger.Debug("AdminMiddleware called by non-admin user")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(res, req)
	})
}

// parseLoginInputFromForm checks a request for a login form, and returns the parsed login data if relevant
func parseLoginInputFromForm(req *http.Request) *models.UserLoginInput {
	if err := req.ParseForm(); err == nil {
		uli := &models.UserLoginInput{
			Username:  req.FormValue(UsernameFormKey),
			Password:  req.FormValue(PasswordFormKey),
			TOTPToken: req.FormValue(TOTPTokenFormKey),
		}

		if uli.Username != "" && uli.Password != "" && uli.TOTPToken != "" {
			return uli
		}
	}
	return nil
}

// UserLoginInputMiddleware fetches user login input from requests
func (s *Service) UserLoginInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "UserLoginInputMiddleware")
		defer span.End()

		x := new(models.UserLoginInput)
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			if x = parseLoginInputFromForm(req); x == nil {
				s.logger.Error(err, "error encountered decoding request body")
				res.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ctx = context.WithValue(ctx, UserLoginInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

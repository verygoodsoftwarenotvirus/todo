package auth

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

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
		ctx, span := trace.StartSpan(req.Context(), "cookie-authentication-middleware")
		defer span.End()

		s.logger.Debug("userCookieAuthenticationMiddleware triggered")
		user, err := s.FetchUserFromRequest(req)
		if err != nil {
			s.logger.Error(err, "error encountered fetching user")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if user != nil {
			req = req.WithContext(context.WithValue(
				context.WithValue(ctx, models.UserKey, user),
				models.UserIDKey,
				user.ID,
			))

			next.ServeHTTP(res, req)
			return
		}
		http.Redirect(res, req, "/login", http.StatusUnauthorized)
	})
}

// AuthenticationMiddleware authenticates based on either an oauth2 token or a cookie
func (s *Service) AuthenticationMiddleware(allowValidCookieInLieuOfAValidToken bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx, span := trace.StartSpan(req.Context(), "authentication-middleware")
			defer span.End()
			s.logger.Debug("apiAuthenticationMiddleware called")

			// let's figure out who the user is
			var user *models.User

			// check for a cookie first:
			if allowValidCookieInLieuOfAValidToken {
				cookieAuth, err := s.DecodeCookieFromRequest(req)
				if err == nil && cookieAuth != nil {
					user, err = s.userDB.GetUser(ctx, cookieAuth.UserID)
					if err != nil {
						s.logger.Error(err, "error authenticating request")
						http.Error(res, "fetching user", http.StatusInternalServerError)
						return
					}
				}
			}

			if user == nil {
				// check to see if there is an OAuth2 token for a valid client attached to the request.
				// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
				oauth2Client, err := s.oauth2ClientsService.RequestIsAuthenticated(req)
				if err != nil || oauth2Client == nil {
					s.logger.Error(err, "fetching oauth2 client")
					http.Redirect(res, req, "/login", http.StatusUnauthorized)
					return
				}

				ctx = context.WithValue(ctx, models.OAuth2ClientKey, oauth2Client)
				user, err = s.userDB.GetUser(ctx, oauth2Client.BelongsTo)
				if err != nil {
					s.logger.Error(err, "error authenticating request")
					http.Error(res, "fetching user", http.StatusInternalServerError)
					return
				}
			}

			if user == nil {
				// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere
				s.logger.Debug("no user attached to request request")
				http.Redirect(res, req, "/login", http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, models.UserIDKey, user.ID)
			ctx = context.WithValue(ctx, models.UserIsAdminKey, user.IsAdmin)
			req = req.WithContext(ctx)

			next.ServeHTTP(res, req)

		})
	}
}

func parseLoginInputFromForm(req *http.Request) *models.UserLoginInput {
	err := req.ParseForm()
	if err == nil {
		uli := &models.UserLoginInput{
			Username:  req.FormValue(UsernameFormKey),
			Password:  req.FormValue(PasswordFormKey),
			TOTPToken: req.FormValue(TOTPTokenFormKey),
		}
		if uli.Username == "" && uli.Password == "" && uli.TOTPToken == "" {
			return nil
		}
		return uli
	}
	return nil
}

// UserLoginInputMiddleware fetches user login input from requests
func (s *Service) UserLoginInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.UserLoginInput)
		s.logger.WithRequest(req).Debug("UserLoginInputMiddleware called")

		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			if formInput := parseLoginInputFromForm(req); formInput != nil {
				ctx := context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, formInput)
				next.ServeHTTP(res, req.WithContext(ctx))
				return
			}

			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

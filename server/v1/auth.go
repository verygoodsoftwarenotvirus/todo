package server

import (
	"context"
	"database/sql"
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"net/http"
)

const (
	// CookieName is the name of the cookie we attach to requests
	CookieName = "todo"
)

type cookieAuth struct {
	Username      string
	Authenticated bool
}

// UserCookieAuthenticationMiddleware authenticates user cookies
func (s *Server) UserCookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debugln("UserAuthenticationMiddleware triggered")

		if cookie, err := req.Cookie(CookieName); err == nil && cookie != nil {
			var ca cookieAuth

			if err := s.cookieBuilder.Decode(CookieName, cookie.Value, &ca); err == nil {
				// // TODO: refresh cookie
				// cookie.Expires = time.Now().Add(s.config.MaxCookieLifetime)
				// http.SetCookie(res, cookie)
				var ctx = req.Context()

				if u := ctx.Value(models.UserKey); u == nil {
					user, err := s.db.GetUser(ctx, ca.Username)
					if err != nil {
						s.logger.WithError(err).Errorln("error encountered fetching user")
						s.internalServerError(res, req, err)
						return
					}

					req = req.WithContext(context.WithValue(
						context.WithValue(ctx, models.UserKey, user),
						models.UserIDKey,
						user.ID,
					))
				}
				next.ServeHTTP(res, req)
				return
			}
			s.logger.Errorf("problem decoding cookie: %v", err)
		}
		http.Redirect(res, req, "/login", http.StatusUnauthorized)
	})
}

// Login is our login route
func (s *Server) Login(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	s.logger.Debugln("Login called")

	loginInput, user, errNotifier, err := s.fetchLoginDataFromRequest(req)
	if err != nil {
		s.logger.WithError(err).Errorln("error encountered fetching login data from request")
		if errNotifier != nil {
			errNotifier(res, req, err)
		} else {
			s.internalServerError(res, req, err)
		}
		return
	}

	logger := s.logger.WithField("login_input", loginInput)

	loginValid, errNotifier, err := s.validateLogin(ctx, user, loginInput)
	if err != nil {
		logger.WithError(err).Errorln("error encountered validating login")
		if errNotifier != nil {
			errNotifier(res, req, err)
		} else {
			s.internalServerError(res, req, err)
		}
		return
	}

	logger = logger.WithField("valid", loginValid)

	if !loginValid {
		logger.Debugln("login was invalid")
		// s.loginMonitor.LogUnsuccessfulAttempt(loginInput.Username)
		s.invalidInput(res, req, nil)
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	logger.Debugln("login was valid, returning cookie")
	cookie, err := s.buildCookie(user, loginValid)
	if err != nil {
		logger.WithError(err).Errorln("error building cookie")
		s.internalServerError(res, req, err)
		return
	}

	http.SetCookie(res, cookie)
	res.WriteHeader(http.StatusOK)
}

// Logout is our logout route
func (s *Server) Logout(res http.ResponseWriter, req *http.Request) {
	if cookie, err := req.Cookie(CookieName); err == nil {
		s.logger.Debugln("logout was called, clearing cookie")
		cookie.MaxAge = -1
		http.SetCookie(res, cookie)
	} else {
		s.logger.WithError(err).Errorln("logout was called, no cookie was found")
	}

	res.WriteHeader(http.StatusOK)
}

func (s *Server) fetchLoginDataFromRequest(req *http.Request) (*models.UserLoginInput, *models.User, ErrorNotifier, error) {
	ctx := req.Context()
	loginInput, ok := ctx.Value(users.MiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		s.logger.Debugln("no UserLoginInput found for /login request")
		return nil, nil, s.notifyUnauthorized, nil
	}
	username := loginInput.Username
	logger := s.logger.WithField("username", username)

	// if err := s.loginMonitor.LoginAttemptsExhausted(username); err != nil {
	// 	s.logger.Debugln("user has exhausted their number of login attempts")
	// 	return nil, nil, func(res http.ResponseWriter, req *http.Request, err error) {
	// 		s.loginMonitor.NotifyExhaustedAttempts(res)
	// 	}, err
	// }

	// you could ensure there isn't an unsatisfied password reset token requested before allowing login here

	user, err := s.db.GetUser(ctx, username)
	if err == sql.ErrNoRows {
		logger.WithError(err).Debugln("no matching user")
		return nil, nil, s.invalidInput, err
	} else if err != nil {
		logger.WithError(err).Debugln("error fetching user")
		return nil, nil, s.internalServerError, err
	}

	return loginInput, user, nil, nil
}

func (s *Server) validateLogin(
	ctx context.Context,
	user *models.User,
	loginInput *models.UserLoginInput,
) (bool, ErrorNotifier, error) {
	loginValid, err := s.authenticator.ValidateLogin(
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
	)
	if err == auth.ErrPasswordHashTooWeak && loginValid {
		s.logger.Debugln("hashed password was deemed to weak, updating its hash")

		updated, e := s.authenticator.HashPassword(loginInput.Password)
		if e != nil {
			return false, s.internalServerError, e
		}

		user.HashedPassword = updated
		if err := s.db.UpdateUser(ctx, user); err != nil {
			return false, s.internalServerError, err
		}
	} else if err != nil {
		return false, s.internalServerError, err
	}

	return loginValid, nil, nil
}

func (s *Server) buildCookie(user *models.User, loginValid bool) (*http.Cookie, error) {
	s.logger.WithFields(map[string]interface{}{
		"user_id":     user.ID,
		"login_valid": loginValid,
	}).Debugln("buildCookie called")

	encoded, err := s.cookieBuilder.Encode(
		CookieName, cookieAuth{
			Username:      user.Username,
			Authenticated: loginValid,
		},
	)

	if err != nil {
		s.logger.WithError(err).Errorln("error encoding cookie")
		return nil, err
	}

	// https://www.calhoun.io/securing-cookies-in-go/
	return &http.Cookie{
		Name:  CookieName,
		Value: encoded,
		// Defaults to any path on your app, but you can use this
		// to limit to a specific subdirectory.
		Path: "/",
		// true means no scripts, http requests only. This has
		// nothing to do with https vs http
		HttpOnly: true,
		// https vs http
		Secure: true,
		// T // Defaults to host-only, which means exact subdomain
		// O // matching. Only change this to enable subdomains if you
		// D // need to! The below code would work on any subdomain for
		// O // yoursite.com
		///////
		// Domain: s.config.Hostname,
		// Expires: time.Now().Add(s.config.MaxCookieLifetime),
	}, nil
}

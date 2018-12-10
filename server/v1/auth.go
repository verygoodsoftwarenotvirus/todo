package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
)

const (
	CookieName                         = "todo"
	userKey          models.ContextKey = "user"
	sessionUserIDKey                   = "user_id"
	sessionAuthKey                     = "authenticated"
)

type cookieAuth struct {
	UserID        string
	Authenticated bool
}

func (s *Server) UserAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debugln("UserAuthenticationMiddleware triggered")
		if cookie, err := req.Cookie(CookieName); err == nil && cookie != nil {
			var ca cookieAuth
			if err := s.cookieBuilder.Decode(CookieName, cookie.Value, &ca); err == nil {
				s.logger.Debugln("no problem decoding cookie")
				// // TODO: refresh cookie
				// cookie.Expires = time.Now().Add(s.config.MaxCookieLifetime)
				// http.SetCookie(res, cookie)

				ctx := context.WithValue(req.Context(), userKey, ca.UserID)
				next.ServeHTTP(res, req.WithContext(ctx))
				return
			} else {
				s.logger.Debugf("problem decoding cookie: %v", err)
			}
		}
		http.Redirect(res, req, "/login", http.StatusUnauthorized)
	})
}

func (s *Server) userAuthorizationHandler(res http.ResponseWriter, req *http.Request) (string, error) {
	userID, ok := req.Context().Value(userKey).(uint64)
	if !ok {
		s.logger.Debugln("no userKey found for authorization request")
		res.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("")
	}
	return strconv.FormatUint(userID, 10), nil
}

func (s *Server) fetchLoginDataFromRequest(req *http.Request) (*models.UserLoginInput, *models.User, ErrorNotifier, error) {
	loginInput, ok := req.Context().Value(users.MiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		s.logger.Debugln("no UserLoginInput found for /login request")
		return nil, nil, s.notifyUnauthorized, nil
	}
	username := loginInput.Username

	if err := s.loginMonitor.LoginAttemptsExhausted(username); err != nil {
		s.logger.Debugln("user has exhausted their number of login attempts")
		return nil, nil, func(res http.ResponseWriter, req *http.Request, err error) {
			s.loginMonitor.NotifyExhaustedAttempts(res)
		}, err
	}

	// you could ensure there isn't an unsatisfied password reset token requested before allowing login here

	user, err := s.db.GetUser(loginInput.Username)
	if err == sql.ErrNoRows {
		s.logger.Debugf("no matching user: %q", loginInput.Username)
		return nil, nil, s.invalidInput, err
	} else if err != nil {
		s.logger.Debugf("error fetching user: %q", loginInput.Username)
		return nil, nil, s.internalServerError, err
	}
	return loginInput, user, nil, nil
}

func (s *Server) validateLogin(user *models.User, loginInput *models.UserLoginInput) (bool, ErrorNotifier, error) {
	loginValid, err := s.authenticator.ValidateLogin(
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
	)
	if err == auth.ErrPasswordHashTooWeak && loginValid {
		s.logger.Debugln("hashed password was deemed to weak, updating its hash")
		updatedPassword, err := s.authenticator.HashPassword(loginInput.Password)
		if err != nil {
			return false, s.internalServerError, err
		}

		user.HashedPassword = updatedPassword
		if err := s.db.UpdateUser(user); err != nil {
			return false, s.internalServerError, nil
		}
	} else if err != nil {
		return false, s.internalServerError, err
	}

	return loginValid, nil, nil

}

func (s *Server) Login(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("Login called")
	var statusToWrite = http.StatusUnauthorized

	loginInput, user, errNotifier, err := s.fetchLoginDataFromRequest(req)
	if errNotifier != nil {
		errNotifier(res, req, err)
		return
	} else if err != nil {
		s.internalServerError(res, req, err)
	}

	loginValid, errNotifier, err := s.validateLogin(user, loginInput)
	if errNotifier != nil {
		errNotifier(res, req, err)
		return
	} else if err != nil {
		s.internalServerError(res, req, err)
	}

	if !loginValid {
		s.logger.Debugln("login was invalid")
		s.loginMonitor.LogUnsuccessfulAttempt(loginInput.Username)
		s.invalidInput(res, req, nil)
		return
	} else {
		s.logger.Debugln("login was valid, returning cookie")
		encoded, err := s.cookieBuilder.Encode(CookieName, cookieAuth{UserID: user.ID, Authenticated: loginValid})
		if err != nil {
			s.internalServerError(res, req, err)
			return
		}

		// https://www.calhoun.io/securing-cookies-in-go/
		http.SetCookie(res, &http.Cookie{
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
		})
		statusToWrite = http.StatusOK
	}
	res.WriteHeader(statusToWrite)
}

func (s *Server) Logout(res http.ResponseWriter, req *http.Request) {
	if cookie, err := req.Cookie(CookieName); err == nil {
		s.logger.Debugln("logout was called, clearing cookie")
		cookie.MaxAge = -1
		http.SetCookie(res, cookie)
	} else {
		s.logger.Debugln("logout was called, no cookie was found")
	}

	res.WriteHeader(http.StatusOK)
}

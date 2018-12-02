package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

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
	UserID        uint64
	Authenticated bool
}

func (s *Server) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debugln("AuthorizationMiddleware triggered")
		if cookie, err := req.Cookie(CookieName); err == nil {
			var ca cookieAuth
			if err := s.cookieBuilder.Decode(CookieName, cookie.Value, &ca); err == nil {
				s.logger.Debugln("no problem decoding cookie")
				// // refresh cookie
				// cookie.Expires = time.Now().Add(s.config.MaxCookieLifetime)
				// http.SetCookie(res, cookie)

				ctx := context.WithValue(req.Context(), userKey, ca.UserID)
				next.ServeHTTP(res, req.WithContext(ctx))
				return
			} else {
				s.logger.Debugf("problem decoding cookie: %v", err)
			}
		}
		res.WriteHeader(http.StatusUnauthorized)
	})
}

func (s *Server) Login(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("Login called")
	var statusToWrite = http.StatusUnauthorized
	loginInput, ok := req.Context().Value(users.MiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		s.logger.Debugln("no UserLoginInput found for /login request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	username := loginInput.Username

	allUsers, err := s.db.GetUsers(nil)
	if err != nil {
		s.logger.Debugln("no UserLoginInput found for /login request")
	}
	s.logger.Debugf("allUsers: %v", allUsers)

	if err := s.loginMonitor.LoginAttemptsExhausted(username); err != nil {
		s.logger.Debugln("user has exhausted their number of login attempts")
		json.NewEncoder(res).Encode(struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}{
			Status:  http.StatusUnauthorized,
			Message: "exhausted login attempts",
		})
		return
	} else if err != nil {
		s.internalServerError(res, err)
		return
	}

	// you could ensure there isn't an unsatisfied password reset token requested before allowing login here

	user, err := s.db.GetUser(loginInput.Username)
	if err == sql.ErrNoRows {
		s.logger.Debugf("no matching user: %q", loginInput.Username)
		s.invalidInput(res, req)
		return
	}
	if err != nil {
		s.internalServerError(res, err)
		return
	}

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
			s.internalServerError(res, err)
			return
		}

		user.HashedPassword = updatedPassword
		if err := s.db.UpdateUser(user); err != nil {
			s.internalServerError(res, err)
			return
		}
	} else if err != nil {
		s.internalServerError(res, err)
		return
	}

	if !loginValid {
		s.logger.Debugln("login was invalid")
		s.loginMonitor.LogUnsuccessfulAttempt(loginInput.Username)
		s.invalidInput(res, req)
		return
	} else {
		s.logger.Debugln("login was valid, returning cookie")
		encoded, err := s.cookieBuilder.Encode(CookieName, cookieAuth{UserID: user.ID, Authenticated: loginValid})
		if err != nil {
			s.internalServerError(res, err)
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

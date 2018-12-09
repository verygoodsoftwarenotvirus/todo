package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2server "gopkg.in/oauth2.v3/server"
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

func (s *Server) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debugln("AuthorizationMiddleware triggered")
		if cookie, err := req.Cookie(CookieName); err == nil {
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
		res.WriteHeader(http.StatusUnauthorized)
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

// gopkg.in/oauth2.v3/server specific implementations

func (s *Server) setOauth2Defaults() {
	s.oauth2Handler.SetAccessTokenExpHandler(s.AccessTokenExpirationHandler)
	s.oauth2Handler.SetClientAuthorizedHandler(s.ClientAuthorizedHandler)
	s.oauth2Handler.SetClientScopeHandler(s.ClientScopeHandler)
	s.oauth2Handler.SetClientInfoHandler(s.ClientInfoHandler)

	s.oauth2Handler.SetInternalErrorHandler(func(err error) (re *oauth2errors.Response) {
		s.logger.Errorln("Internal Error:", err.Error())
		return
	})

	s.oauth2Handler.SetResponseErrorHandler(func(re *oauth2errors.Response) {
		s.logger.Errorln("Response Error:", re.Error.Error())
	})
}

var _ oauth2server.AccessTokenExpHandler = (*Server)(nil).AccessTokenExpirationHandler

func (s *Server) AccessTokenExpirationHandler(w http.ResponseWriter, r *http.Request) (time.Duration, error) {
	return 10 * time.Minute, nil
}

var _ oauth2server.ClientAuthorizedHandler = (*Server)(nil).ClientAuthorizedHandler

func (s *Server) ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	// AuthorizationCode   GrantType = "authorization_code"
	// ClientCredentials   GrantType = "client_credentials"
	// Refreshing          GrantType = "refresh_token"
	// Implicit            GrantType = "__implicit"
	// PasswordCredentials GrantType = "password"

	if grant == oauth2.Implicit {
		// validate that the client ID is allowed to have implicits somehow?
	}

	return true, nil
}

var _ oauth2server.ClientScopeHandler = (*Server)(nil).ClientScopeHandler

func (s *Server) ClientScopeHandler(clientID, scope string) (allowed bool, err error) {
	if c, err := s.db.GetOauthClient(clientID); err != nil {
		return false, err
	} else {
		for _, s := range c.Scopes {
			if s == scope {
				return true, nil
			}
		}
	}
	return
}

var _ oauth2server.ClientInfoHandler = (*Server)(nil).ClientInfoHandler

func (s *Server) ClientInfoHandler(req *http.Request) (clientID, clientSecret string, err error) {
	c, err := s.db.GetOauthClient(req.Header.Get("X-TODO-CLIENT-ID"))
	if err != nil {
		return
	}
	return c.ClientID, c.ClientSecret, nil
}

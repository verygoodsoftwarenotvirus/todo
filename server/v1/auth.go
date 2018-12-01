package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
)

const (
	CookieName     = "todo"
	sessionUserKey = "user_id"
	sessionAuthKey = "authenticated"
)

func (s *Server) Login(res http.ResponseWriter, req *http.Request) {
	loginInput, ok := req.Context().Value(users.MiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	username := loginInput.Username

	if err := s.loginMonitor.LoginAttemptsExhausted(username); err != nil {
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

	// you could ensure there isn't an unsatisfied password reset token requested before allowing login

	user, err := s.db.GetUser(loginInput.Username)
	if err == sql.ErrNoRows {
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
		s.loginMonitor.LogUnsuccessfulAttempt(loginInput.Username)
		s.invalidInput(res, req)
		return
	}

	session, err := s.cookieStore.Get(req, CookieName)
	if err != nil {
		s.internalServerError(res, err)
		return
	}

	statusToWrite := http.StatusUnauthorized
	if loginValid {
		statusToWrite = http.StatusOK
		session.Values[sessionUserKey] = user.ID
		session.Values[sessionAuthKey] = true
		session.Save(req, res)
	}
	res.WriteHeader(statusToWrite)
}

func (s *Server) Logout(res http.ResponseWriter, req *http.Request) {
	session, err := s.cookieStore.Get(req, CookieName)
	if err != nil {
		s.notifyOfInvalidRequestCookie(res, req)
		return
	}
	session.Values[sessionAuthKey] = false
	session.Save(req, res)
	res.WriteHeader(http.StatusOK)
}

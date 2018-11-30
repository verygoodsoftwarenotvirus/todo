package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	URIParamKey = "userID"
)

func (us *UsersService) UserInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.UserLoginInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			us.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (us *UsersService) UserContextMiddleware(next http.Handler) http.Handler {
	x := new(models.UserLoginInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			us.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

type BruteForceLoginDetector interface {
	LogSuccessfulAttempt(userID string)
	LogUnsuccessfulAttempt(userID string)
	LoginAttemptsExhausted(userID string) error
}

type LoginAttemptLogger interface {
}

func (us *UsersService) Login(res http.ResponseWriter, req *http.Request) {
	loginInput, ok := req.Context().Value(MiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	username := loginInput.Username

	if err := us.loginMonitor.LoginAttemptsExhausted(username); err != nil {
		// notifyOfExaustedAuthenticationAttempts(res)
		return
	} else if err != nil {
		// notifyOfInternalIssue(res, err, "retrieve user")
		return
	}

	// TODO: we should ensure there isn't an unsatisfied password reset token requested before allowing login

	user, err := us.database.GetUser(loginInput.Username)
	if err == sql.ErrNoRows {
		// respondThatRowDoesNotExist(req, res, "user", username)
		return
	}
	if err != nil {
		// notifyOfInternalIssue(res, err, "retrieve user")
		return
	}

	loginValid, err := us.authenticator.ValidateLogin(
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
	)
	if err == auth.ErrPasswordHashTooWeak && loginValid {
		updatedPassword, err := us.authenticator.HashPassword(loginInput.Password)
		if err != nil {
			// notifyOfInvalidRequestBody(res, err)
			return
		}

		user.HashedPassword = updatedPassword
		if err := us.database.UpdateUser(user); err != nil {
			// notifyOfInternalIssue(res, err, "retrieve user")
			return
		}
	} else if err != nil {
		// notifyOfInternalIssue(res, err, "retrieve user")
		return
	}
	if !loginValid {
		us.loginMonitor.LogUnsuccessfulAttempt(loginInput.Username)
		// notifyOfInvalidAuthenticationAttempt(res)
		return
	}

	// // session, err := store.Get(req, cookieName)
	// if err != nil {
	// 	// notifyOfInternalIssue(res, err, "instantiate new cookie")
	// 	return
	// }

	statusToWrite := http.StatusUnauthorized
	if loginValid {
		statusToWrite = http.StatusOK
		// session.Values[sessionUserIDKeyName] = user.ID
		// session.Values[sessionAuthorizedKeyName] = true
		// session.Values[sessionAdminKeyName] = user.IsAdmin
		// session.Save(req, res)
	}
	res.WriteHeader(statusToWrite)
}

func (us *UsersService) Logout(res http.ResponseWriter, req *http.Request) {
	// session, err := store.Get(req, cookieName)
	// if err != nil {
	// 	notifyOfInvalidRequestCookie(res)
	// 	return
	// }
	// session.Values[sessionAuthorizedKeyName] = false
	// session.Save(req, res)
	res.WriteHeader(http.StatusOK)
}

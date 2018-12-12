package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/go-chi/chi"
)

const (
	URIParamKey = "userID"
)

func (s *UsersService) UserLoginInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.UserLoginInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (s *UsersService) UserInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.UserInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (s *UsersService) PasswordUpdateInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.PasswordUpdateInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (s *UsersService) TOTPSecretRefreshInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.TOTPSecretRefreshInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (s *UsersService) Read(res http.ResponseWriter, req *http.Request) {
	userID := chi.URLParam(req, URIParamKey)
	x, err := s.database.GetUser(userID)
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Errorf("error fetching user %q from database: %v", userID, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(x)
}

func (s *UsersService) Count(res http.ResponseWriter, req *http.Request) {
	qf := models.ParseQueryFilter(req)
	userCount, err := s.database.GetUserCount(qf)
	if err != nil {
		s.logger.Errorf("error fetching item count from database: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-type", "application/json")

	json.NewEncoder(res).Encode(struct {
		Count uint64 `json:"count"`
	}{userCount})
}

func (s *UsersService) List(res http.ResponseWriter, req *http.Request) {
	qf := models.ParseQueryFilter(req)
	users, err := s.database.GetUsers(qf)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(users)
}

func (s *UsersService) Delete(res http.ResponseWriter, req *http.Request) {
	userIDParam := chi.URLParam(req, URIParamKey)
	userID, _ := strconv.ParseUint(userIDParam, 10, 64)

	if err := s.database.DeleteUser(uint(userID)); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *UsersService) validateCredentialChangeRequest(
	usernameFetcher func(req *http.Request) string,
	req *http.Request,
	password string,
	totpToken string,
) (user *models.User, statusCode int) {
	var err error
	username := usernameFetcher(req)
	user, err = s.database.GetUser(username)
	if err != nil {
		s.logger.Errorf("error encountered fecthing user %q: %v", username, err)
		return nil, http.StatusInternalServerError
	}

	valid, err := s.authenticator.ValidateLogin(
		user.HashedPassword,
		password,
		user.TwoFactorSecret,
		totpToken,
	)
	if !valid {
		s.logger.Debugf("invalid attempt to cycle TOTP token by user %q: %v", username, err)
		return nil, http.StatusUnauthorized
	} else if err != nil {
		s.logger.Errorf("error encountered generating random TOTP string for user %q: %v", username, err)
		return nil, http.StatusInternalServerError
	}

	return user, 0
}

// NewTOTPSecret fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid
func (s *UsersService) NewTOTPSecret(usernameFetcher func(req *http.Request) string) http.HandlerFunc {
	if usernameFetcher == nil {
		panic("usernameFetcher must be provided")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		input, ok := req.Context().Value(MiddlewareCtxKey).(*models.TOTPSecretRefreshInput)
		if !ok {
			s.logger.Debugln("no input found on TOTP Secret refresh request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		user, sc := s.validateCredentialChangeRequest(
			usernameFetcher,
			req,
			input.CurrentPassword,
			input.TOTPToken,
		)
		if sc != 0 {
			res.WriteHeader(sc)
			return
		}

		user.TwoFactorSecret, err = auth.RandString(52) // I forgot how I know it needs to be this long
		if err != nil {
			s.logger.Errorf("error encountered generating random TOTP string for user %q: %v", user.Username, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := s.database.UpdateUser(user); err != nil {
			s.logger.Errorf("error encountered updating TOTP token for user %q: %v", user.Username, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-type", "application/json")
		json.NewEncoder(res).Encode(user)
	}
}

// UpdatePassword updates a user's password, after validating
// that information received from PasswordUpdateInputContextMiddleware is valid
func (s *UsersService) UpdatePassword(usernameFetcher func(req *http.Request) string) http.HandlerFunc {
	if usernameFetcher == nil {
		panic("usernameFetcher must be provided")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		input, ok := req.Context().Value(MiddlewareCtxKey).(*models.PasswordUpdateInput)
		if !ok {
			s.logger.Debugln("no input found on TOTP Secret refresh request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		user, sc := s.validateCredentialChangeRequest(
			usernameFetcher,
			req,
			input.CurrentPassword,
			input.TOTPToken,
		)
		if sc != 0 {
			res.WriteHeader(sc)
			return
		}

		user.HashedPassword, err = s.authenticator.HashPassword(input.NewPassword)
		if err != nil {
			s.logger.Errorf("error encountered generating random TOTP string for user %q: %v", user.Username, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := s.database.UpdateUser(user); err != nil {
			s.logger.Errorf("error encountered updating TOTP token for user %q: %v", user.Username, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-type", "application/json")
		json.NewEncoder(res).Encode(user)
	}
}

// Create is our user creation route
// note that Create is meant to be executed after UserInputContextMiddleware
func (s *UsersService) Create(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.UserInput)
	if !ok {
		s.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	x, err := s.database.CreateUser(input)
	if err != nil {
		s.logger.Errorf("error creating user: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(x)
}

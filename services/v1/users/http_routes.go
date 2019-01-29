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
	// URIParamKey is used to refer to user IDs in router params
	URIParamKey = "userID"
)

// UserLoginInputContextMiddleware fetches user login input from requests
func (s *Service) UserLoginInputContextMiddleware(next http.Handler) http.Handler {
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

// UserInputContextMiddleware fetches user input from requests
func (s *Service) UserInputContextMiddleware(next http.Handler) http.Handler {
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

// PasswordUpdateInputContextMiddleware fetches password update input from requests
func (s *Service) PasswordUpdateInputContextMiddleware(next http.Handler) http.Handler {
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

// TOTPSecretRefreshInputContextMiddleware fetches 2FA update input from requests
func (s *Service) TOTPSecretRefreshInputContextMiddleware(next http.Handler) http.Handler {
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

// Read is our read route
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	userID := s.usernameFetcher(req)
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

// Count is a handler for responding with a count of users
func (s *Service) Count(res http.ResponseWriter, req *http.Request) {
	qf := models.ExtractQueryFilter(req)
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

// List is a handler for responding with a list of users
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	qf := models.ExtractQueryFilter(req)
	users, err := s.database.GetUsers(qf)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(users)
}

// Delete is a handler for deleting a user
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	username := s.usernameFetcher(req)
	s.logger.Debugf("UsersService.Delete called for user %s", username)
	if err := s.database.DeleteUser(username); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type usernameFetcher func(req *http.Request) string

func (s *Service) validateCredentialChangeRequest(req *http.Request, password string, totpToken string) (user *models.User, statusCode int) {
	var err error
	username := s.usernameFetcher(req)
	user, err = s.database.GetUser(username)
	if err != nil {
		s.logger.Errorf("error encountered fecthing user %q: %v", username, err)
		return nil, http.StatusInternalServerError
	}

	if valid, err := s.authenticator.ValidateLogin(user.HashedPassword, password, user.TwoFactorSecret, totpToken); err != nil {
		s.logger.Errorf("error encountered generating random TOTP string for user %q: %v", username, err)
		return nil, http.StatusInternalServerError
	} else if !valid {
		s.logger.Debugf("invalid attempt to cycle TOTP token by user %q: %v", username, err)
		return nil, http.StatusUnauthorized
	}

	return user, 0
}

// NewTOTPSecret fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid
func (s *Service) NewTOTPSecret(res http.ResponseWriter, req *http.Request) {
	var err error
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.TOTPSecretRefreshInput)
	if !ok {
		s.logger.Debugln("no input found on TOTP Secret refresh request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, sc := s.validateCredentialChangeRequest(req, input.CurrentPassword, input.TOTPToken)
	if sc != 0 {
		res.WriteHeader(sc)
		return
	}

	tfc, err := auth.RandString(64)
	if err != nil {
		s.logger.Errorf("error encountered generating random TOTP string for user %q: %v", user.Username, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	user.TwoFactorSecret = tfc

	if err := s.database.UpdateUser(user); err != nil {
		s.logger.Errorf("error encountered updating TOTP token for user %q: %v", user.Username, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(user)
}

// UpdatePassword updates a user's password, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid
func (s *Service) UpdatePassword(res http.ResponseWriter, req *http.Request) {
	var err error
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.PasswordUpdateInput)
	if !ok {
		s.logger.Debugln("no input found on TOTP Secret refresh request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, sc := s.validateCredentialChangeRequest(req, input.CurrentPassword, input.TOTPToken)
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

// Create is our user creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.UserInput)
	if !ok {
		s.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	hp, err := s.authenticator.HashPassword(input.Password)
	if err != nil {
		s.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	input.Password = hp

	newSecret, err := auth.RandString(64)
	if err != nil {
		s.logger.Errorln("error generating TOTP secret: ", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	totpSecret := newSecret //[:52] // I forgot how I know it needs to be this long

	x, err := s.database.CreateUser(input, totpSecret)
	if err != nil {
		s.logger.Errorf("error creating user: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// UserCreationResponse is a struct we can use to notify the user of
	// their two factor secret, but ideally just this once and then never again.
	ucr := &models.UserCreationResponse{
		ID:                    x.ID,
		Username:              x.Username,
		TwoFactorSecret:       x.TwoFactorSecret,
		PasswordLastChangedOn: x.PasswordLastChangedOn,
		CreatedOn:             x.CreatedOn,
		UpdatedOn:             x.UpdatedOn,
		ArchivedOn:            x.ArchivedOn,
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(ucr)
}

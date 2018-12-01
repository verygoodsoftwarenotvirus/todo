package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/go-chi/chi"
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
	x := new(models.UserInput)
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

func (us *UsersService) Read(res http.ResponseWriter, req *http.Request) {
	userID := chi.URLParam(req, URIParamKey)
	i, err := us.database.GetUser(userID)
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		us.logger.Errorf("error fetching user %q from database: %v", userID, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

func (us *UsersService) List(res http.ResponseWriter, req *http.Request) {
	qf := models.ParseQueryFilter(req)
	users, err := us.database.GetUsers(qf)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(users)
}

func (us *UsersService) Delete(res http.ResponseWriter, req *http.Request) {
	userIDParam := chi.URLParam(req, URIParamKey)
	userID, _ := strconv.ParseUint(userIDParam, 10, 64)

	if err := us.database.DeleteUser(uint(userID)); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Update is the update route for
func (us *UsersService) Update(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.UserInput)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := chi.URLParam(req, URIParamKey)
	u, err := us.database.GetUser(userID)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var hashedPass = u.HashedPassword
	if !us.authenticator.PasswordMatches(hashedPass, input.Password) {
		hp, err := us.authenticator.HashPassword(input.Password)
		if err != nil {
		}
		hashedPass = hp
	}

	u.Update(models.User{
		Username:        input.Username,
		HashedPassword:  hashedPass,
		TwoFactorSecret: input.TOTPSecret,
	})
	if err := us.database.UpdateUser(u); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(u)
}

func (us *UsersService) Create(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.UserInput)
	if !ok {
		us.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	i, err := us.database.CreateUser(input)
	if err != nil {
		us.logger.Errorf("error creating user: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

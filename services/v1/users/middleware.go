package users

import (
	"context"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// UserLoginInputContextMiddleware fetches user login input from requests
func (s *Service) UserLoginInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.UserLoginInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithRequest(req).Debug("UserLoginInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
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
		s.logger.WithRequest(req).Debug("UserInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
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
		s.logger.WithRequest(req).Debug("PasswordUpdateInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
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
		s.logger.WithRequest(req).Debug("TOTPSecretRefreshInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

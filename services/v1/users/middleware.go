package users

import (
	"context"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// UserCreationMiddlewareCtxKey is the context key for creation input
	UserCreationMiddlewareCtxKey models.ContextKey = "user_creation_input"

	// PasswordChangeMiddlewareCtxKey is the context key for password changes
	PasswordChangeMiddlewareCtxKey models.ContextKey = "user_password_change"

	// TOTPSecretRefreshMiddlewareCtxKey is the context key for 2fa token refreshes
	TOTPSecretRefreshMiddlewareCtxKey models.ContextKey = "totp_refresh"
)

// UserInputMiddleware fetches user input from requests
func (s *Service) UserInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.UserInput)
		s.logger.WithRequest(req).Debug("UserInputMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PasswordUpdateInputMiddleware fetches password update input from requests
func (s *Service) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.PasswordUpdateInput)
		s.logger.WithRequest(req).Debug("PasswordUpdateInputMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), PasswordChangeMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretRefreshInputMiddleware fetches 2FA update input from requests
func (s *Service) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.TOTPSecretRefreshInput)
		s.logger.WithRequest(req).Debug("TOTPSecretRefreshInputMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), TOTPSecretRefreshMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

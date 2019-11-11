package users

import (
	"context"
	"net/http"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

const (
	// UserCreationMiddlewareCtxKey is the context key for creation input
	UserCreationMiddlewareCtxKey models.ContextKey = "user_creation_input"

	// PasswordChangeMiddlewareCtxKey is the context key for password changes
	PasswordChangeMiddlewareCtxKey models.ContextKey = "user_password_change"

	// TOTPSecretRefreshMiddlewareCtxKey is the context key for TOTP token refreshes
	TOTPSecretRefreshMiddlewareCtxKey models.ContextKey = "totp_refresh"
)

// UserInputMiddleware fetches user input from requests
func (s *Service) UserInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.UserInput)
		ctx, span := trace.StartSpan(req.Context(), "UserInputMiddleware")
		defer span.End()

		// decode the request
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// attach parsed value to request context
		ctx = context.WithValue(ctx, UserCreationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PasswordUpdateInputMiddleware fetches password update input from requests
func (s *Service) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.PasswordUpdateInput)
		ctx, span := trace.StartSpan(req.Context(), "PasswordUpdateInputMiddleware")
		defer span.End()

		// decode the request
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// attach parsed value to request context
		ctx = context.WithValue(ctx, PasswordChangeMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretRefreshInputMiddleware fetches 2FA update input from requests
func (s *Service) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.TOTPSecretRefreshInput)
		ctx, span := trace.StartSpan(req.Context(), "TOTPSecretRefreshInputMiddleware")
		defer span.End()

		// decode the request
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// attach parsed value to request context
		ctx = context.WithValue(ctx, TOTPSecretRefreshMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

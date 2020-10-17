package users

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// userCreationMiddlewareCtxKey is the context key for creation input.
	userCreationMiddlewareCtxKey models.ContextKey = "user_creation_input"

	// passwordChangeMiddlewareCtxKey is the context key for password changes.
	passwordChangeMiddlewareCtxKey models.ContextKey = "user_password_change"

	// totpSecretVerificationMiddlewareCtxKey is the context key for TOTP token refreshes.
	totpSecretVerificationMiddlewareCtxKey models.ContextKey = "totp_verify"

	// totpSecretRefreshMiddlewareCtxKey is the context key for TOTP token refreshes.
	totpSecretRefreshMiddlewareCtxKey models.ContextKey = "totp_refresh"
)

// UserInputMiddleware fetches user input from requests.
func (s *Service) UserInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.UserCreationInput)
		ctx, span := tracing.StartSpan(req.Context(), "UserInputMiddleware")
		defer span.End()

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, userCreationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PasswordUpdateInputMiddleware fetches password update input from requests.
func (s *Service) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.PasswordUpdateInput)
		ctx, span := tracing.StartSpan(req.Context(), "PasswordUpdateInputMiddleware")
		defer span.End()

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, passwordChangeMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretVerificationInputMiddleware fetches 2FA update input from requests.
func (s *Service) TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.TOTPSecretVerificationInput)
		ctx, span := tracing.StartSpan(req.Context(), "TOTPSecretVerificationInputMiddleware")
		defer span.End()

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, totpSecretVerificationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretRefreshInputMiddleware fetches 2FA update input from requests.
func (s *Service) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.TOTPSecretRefreshInput)
		ctx, span := tracing.StartSpan(req.Context(), "TOTPSecretRefreshInputMiddleware")
		defer span.End()

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, totpSecretRefreshMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

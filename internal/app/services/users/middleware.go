package users

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// userCreationMiddlewareCtxKey is the context key for creation input.
	userCreationMiddlewareCtxKey types.ContextKey = "user_creation_input"

	// passwordChangeMiddlewareCtxKey is the context key for password changes.
	passwordChangeMiddlewareCtxKey types.ContextKey = "user_password_change"

	// totpSecretVerificationMiddlewareCtxKey is the context key for TOTP token refreshes.
	totpSecretVerificationMiddlewareCtxKey types.ContextKey = "totp_verify"

	// totpSecretRefreshMiddlewareCtxKey is the context key for TOTP token refreshes.
	totpSecretRefreshMiddlewareCtxKey types.ContextKey = "totp_refresh"
)

// UserCreationInputMiddleware fetches user input from requests.
func (s *Service) UserCreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.UserCreationInput)
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		//if err := x.Validate(4, 6); err != nil {
		//	logger.Error(err, "provided input was invalid")
		//	s.encoderDecoder.EncodeErrorResponse(res, err.Error(), http.StatusBadRequest)
		//	return
		//}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, userCreationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PasswordUpdateInputMiddleware fetches password update input from requests.
func (s *Service) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.PasswordUpdateInput)
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		//if err := x.Validate(); err != nil {
		//	logger.Error(err, "provided input was invalid")
		//	s.encoderDecoder.EncodeErrorResponse(res, err.Error(), http.StatusBadRequest)
		//	return
		//}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, passwordChangeMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretVerificationInputMiddleware fetches 2FA update input from requests.
func (s *Service) TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.TOTPSecretVerificationInput)
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		//if err := x.Validate(); err != nil {
		//	logger.Error(err, "provided input was invalid")
		//	s.encoderDecoder.EncodeErrorResponse(res, err.Error(), http.StatusBadRequest)
		//	return
		//}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, totpSecretVerificationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretRefreshInputMiddleware fetches 2FA update input from requests.
func (s *Service) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.TOTPSecretRefreshInput)
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		//if err := x.Validate(); err != nil {
		//	logger.Error(err, "provided input was invalid")
		//	s.encoderDecoder.EncodeErrorResponse(res, err.Error(), http.StatusBadRequest)
		//	return
		//}

		// attach parsed value to request context.
		ctx = context.WithValue(ctx, totpSecretRefreshMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

package users

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// passwordChangeMiddlewareCtxKey is the context key for passwords changes.
	passwordChangeMiddlewareCtxKey types.ContextKey = "user_password_change"

	// totpSecretVerificationMiddlewareCtxKey is the context key for TOTP token refreshes.
	totpSecretVerificationMiddlewareCtxKey types.ContextKey = "totp_verify"

	// totpSecretRefreshMiddlewareCtxKey is the context key for TOTP token refreshes.
	totpSecretRefreshMiddlewareCtxKey types.ContextKey = "totp_refresh"

	urlEncodedFormHeaderKey = "application/x-www-form-urlencoded"
)

const (
	// usernameFormKey is the string we look for in request forms for username information.
	usernameFormKey = "username"
	// passwordFormKey is the string we look for in request forms for passwords information.
	passwordFormKey = "password"
)

// parseFormEncodedRegistrationRequest checks a request for a registration form, and returns the parsed login data if relevant.
func parseFormEncodedRegistrationRequest(req *http.Request) *types.UserRegistrationInput {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil
	}

	input := &types.UserRegistrationInput{
		Username: form.Get(usernameFormKey),
		Password: form.Get(passwordFormKey),
	}

	if input.Username != "" && input.Password != "" {
		return input
	}

	return nil
}

// UserRegistrationInputMiddleware fetches user input from requests.
func (s *service) UserRegistrationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.UserRegistrationInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if strings.HasPrefix(req.Header.Get("Content-Type"), urlEncodedFormHeaderKey) {
			if x = parseFormEncodedRegistrationRequest(req); x == nil {
				s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
				return
			}
		} else if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx, s.authSettings.MinimumUsernameLength, s.authSettings.MinimumPasswordLength); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		logger.WithValue("input", x).Info("WE MADE IT THROUGH THE MIDDLEWARE")

		// attach parsed value to session context data.
		ctx = context.WithValue(ctx, types.UserRegistrationInputContextKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PasswordUpdateInputMiddleware fetches passwords update input from requests.
func (s *service) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.PasswordUpdateInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx, s.authSettings.MinimumPasswordLength); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		// attach parsed value to session context data.
		ctx = context.WithValue(ctx, passwordChangeMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretVerificationInputMiddleware fetches 2FA update input from requests.
func (s *service) TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.TOTPSecretVerificationInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		// attach parsed value to session context data.
		ctx = context.WithValue(ctx, totpSecretVerificationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretRefreshInputMiddleware fetches 2FA update input from requests.
func (s *service) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.TOTPSecretRefreshInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// decode the request.
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		// attach parsed value to session context data.
		ctx = context.WithValue(ctx, totpSecretRefreshMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

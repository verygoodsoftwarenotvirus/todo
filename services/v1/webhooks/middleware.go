package webhooks

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// CreationInputMiddleware is a middleware for fetching, parsing, and attaching a parsed WebhookCreationInput struct from a request.
func (s *Service) CreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.WebhookCreationInput)
		ctx, span := tracing.StartSpan(req.Context(), "CreationInputMiddleware")
		defer span.End()

		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, createMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// UpdateInputMiddleware is a middleware for fetching, parsing, and attaching a parsed WebhookCreationInput struct from a request.
// This is the same as the creation one, but it won't always be.
func (s *Service) UpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(models.WebhookUpdateInput)
		ctx, span := tracing.StartSpan(req.Context(), "UpdateInputMiddleware")
		defer span.End()

		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, updateMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

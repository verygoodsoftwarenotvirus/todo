package webhooks

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// CreationInputMiddleware is a middleware for fetching, parsing, and attaching a parsed WebhookInput struct from a request
func (s *Service) CreationInputMiddleware(next http.Handler) http.Handler {
	x := new(models.WebhookInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := s.encoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		s.logger.
			WithRequest(req).
			WithValue("webhookInput", x).
			Debug("WebhookInputMiddleware called")
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)

		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// UpdateInputMiddleware is a middleware for fetching, parsing, and attaching a parsed WebhookInput struct from a request
// This is the same as the creation one, but it won't always be
func (s *Service) UpdateInputMiddleware(next http.Handler) http.Handler {
	x := new(models.WebhookInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := s.encoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		s.logger.
			WithRequest(req).
			WithValue("webhookInput", x).
			Debug("UpdateInputMiddleware called")
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)

		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

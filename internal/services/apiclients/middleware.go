package apiclients

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// CreationInputMiddleware is a middleware for attaching API client info to a request.
func (s *service) CreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()
		x := new(types.APIClientCreationInput)

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		// decode value from request.
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, creationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (s *service) fetchAPIClientFromRequest(req *http.Request) *types.APIClient {
	if client, ok := req.Context().Value(types.APIClientKey).(*types.APIClient); ok {
		return client
	}
	return nil
}

func (s *service) fetchAPIClientIDFromRequest(req *http.Request) string {
	if clientID, ok := req.Context().Value(clientIDKey).(string); ok {
		return clientID
	}
	return ""
}

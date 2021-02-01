package delegatedclients

import (
	"context"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	scopesSeparator = ","
)

var errClientUnauthorizedForScope = errors.New("client not authorized for scope")

// CreationInputMiddleware is a middleware for attaching Delegated client info to a request.
func (s *service) CreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()
		x := new(types.DelegatedClientCreationInput)

		// decode value from request.
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, creationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (s *service) fetchDelegatedClientFromRequest(req *http.Request) *types.DelegatedClient {
	if client, ok := req.Context().Value(types.DelegatedClientKey).(*types.DelegatedClient); ok {
		return client
	}
	return nil
}

func (s *service) fetchDelegatedClientIDFromRequest(req *http.Request) string {
	if clientID, ok := req.Context().Value(clientIDKey).(string); ok {
		return clientID
	}
	return ""
}

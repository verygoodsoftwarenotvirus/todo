package admin

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// accountStatusUpdateMiddlewareCtxKey is a string alias we can use for referring to item input data in contexts.
	accountStatusUpdateMiddlewareCtxKey types.ContextKey = "account_status_update_input"
)

// AccountStatusUpdateInputMiddleware is a middleware for fetching, parsing, and attaching a AccountStatusUpdateInput struct from a request.
func (s *Service) AccountStatusUpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.AccountStatusUpdateInput)
		ctx, span := tracing.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.Error(err, "provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, accountStatusUpdateMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

package admin

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// accountStatusUpdateMiddlewareCtxKey is a string alias we can use for referring to item input data in contexts.
	accountStatusUpdateMiddlewareCtxKey types.ContextKey = "account_status_update_input"
)

// AccountStatusUpdateInputMiddleware is a middleware for fetching, parsing, and attaching a UserReputationUpdateInput struct from a request.
func (s *service) AccountStatusUpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.UserReputationUpdateInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, accountStatusUpdateMiddlewareCtxKey, x)))
	})
}

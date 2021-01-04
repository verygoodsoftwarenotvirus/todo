package accounts

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// createMiddlewareCtxKey is a string alias we can use for referring to account input data in contexts.
	createMiddlewareCtxKey types.ContextKey = "account_create_input"
	// updateMiddlewareCtxKey is a string alias we can use for referring to account update data in contexts.
	updateMiddlewareCtxKey types.ContextKey = "account_update_input"
)

// CreationInputMiddleware is a middleware for fetching, parsing, and attaching an AccountInput struct from a request.
func (s *service) CreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.AccountCreationInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.Error(err, "provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, createMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// UpdateInputMiddleware is a middleware for fetching, parsing, and attaching an AccountInput struct from a request.
// This is the same as the creation one, but that won't always be the case.
func (s *service) UpdateInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.AccountUpdateInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.Validate(ctx); err != nil {
			logger.Error(err, "provided input was invalid")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, updateMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// AccountRestrictionMiddleware is a middleware for ensuring a given request only applies to a member of a given route.
func (s *service) AccountRestrictionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

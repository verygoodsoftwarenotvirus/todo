package accounts

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// addUserToAccountMiddlewareCtxKey is a string alias we can use for referring to account input data in contexts.
	addUserToAccountMiddlewareCtxKey types.ContextKey = "add_user_to_account"
	// transferAccountMiddlewareCtxKey is a string alias we can use for referring to account input data in contexts.
	transferAccountMiddlewareCtxKey types.ContextKey = "transfer_account"
	// modifyMembershipMiddlewareCtxKey is a string alias we can use for referring to account input data in contexts.
	modifyMembershipMiddlewareCtxKey types.ContextKey = "modify_membership"
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
		tracing.AttachRequestToSpan(span, req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("attached creation input for account to request")

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
		tracing.AttachRequestToSpan(span, req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("attached update input for account to request")

		ctx = context.WithValue(ctx, updateMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// AddMemberInputMiddleware does .
func (s *service) AddMemberInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.AddUserToAccountInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("attached account membership creation input to request")

		ctx = context.WithValue(ctx, addUserToAccountMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// ModifyMemberPermissionsInputMiddleware does .
func (s *service) ModifyMemberPermissionsInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.ModifyUserPermissionsInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("attached membership permission modification input to request")

		ctx = context.WithValue(ctx, modifyMembershipMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// AccountTransferInputMiddleware does .
func (s *service) AccountTransferInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		x := new(types.TransferAccountOwnershipInput)
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("attached account transfer input to request")

		ctx = context.WithValue(ctx, transferAccountMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

package webhooks

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

const (
	// WebhookIDURIParamKey is a standard string that we'll use to refer to webhook IDs with.
	WebhookIDURIParamKey = "webhookID"
)

// CreateHandler is our webhook creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	input := new(types.WebhookCreationInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	input.ID = ksuid.New().String()

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("provided input was invalid")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	input.BelongsToAccount = sessionCtxData.ActiveAccountID

	// create the webhook.
	wh, err := s.webhookDataManager.CreateWebhook(ctx, input)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify the relevant parties.
	tracing.AttachWebhookIDToSpan(span, wh.ID)
	s.webhookCounter.Increment(ctx)

	// let everybody know we're good.
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, wh, http.StatusCreated)
}

// ListHandler is our list route.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	filter := types.ExtractQueryFilter(req)
	logger := filter.AttachToLogger(s.logger)

	tracing.AttachRequestToSpan(span, req)
	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// find the webhooks.
	webhooks, err := s.webhookDataManager.GetWebhooks(ctx, sessionCtxData.ActiveAccountID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		webhooks = &types.WebhookList{
			Webhooks: []*types.Webhook{},
		}
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching webhooks")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode the response.
	s.encoderDecoder.RespondWithData(ctx, res, webhooks)
}

// ReadHandler returns a GET handler that returns an webhook.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)

	tracing.AttachAccountIDToSpan(span, sessionCtxData.ActiveAccountID)
	logger = logger.WithValue(keys.AccountIDKey, sessionCtxData.ActiveAccountID)

	// fetch the webhook from the database.
	webhook, err := s.webhookDataManager.GetWebhook(ctx, webhookID, sessionCtxData.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("No rows found in webhook database")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching webhook from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode the response.
	s.encoderDecoder.RespondWithData(ctx, res, webhook)
}

// UpdateHandler returns a handler that updates an webhook.
func (s *service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = sessionCtxData.AttachToLogger(logger)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)

	input := new(types.WebhookUpdateInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("provided input was invalid")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	// fetch the webhook in question.
	webhook, err := s.webhookDataManager.GetWebhook(ctx, webhookID, sessionCtxData.ActiveAccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("nonexistent webhook requested for update")
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		} else {
			logger.Error(err, "error encountered getting webhook")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		}

		return
	}

	// update it.
	webhook.Update(input)

	// save the update in the database.
	if err = s.webhookDataManager.UpdateWebhook(ctx, webhook); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("attempted to update nonexistent webhook")
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		} else {
			observability.AcknowledgeError(err, logger, span, "updating webhook")
			logger.Error(err, "error encountered updating webhook")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		}

		return
	}

	// let everybody know we're good.
	s.encoderDecoder.RespondWithData(ctx, res, webhook)
}

// ArchiveHandler returns a handler that archives an webhook.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine relevant user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	userID := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.UserIDKey, userID)

	accountID := sessionCtxData.ActiveAccountID
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)

	// do the deed.
	err = s.webhookDataManager.ArchiveWebhook(ctx, webhookID, sessionCtxData.ActiveAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no rows found for webhook")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// let the interested parties know.
	s.webhookCounter.Decrement(ctx)

	// let everybody go home.
	res.WriteHeader(http.StatusNoContent)
}

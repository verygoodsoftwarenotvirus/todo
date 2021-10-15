package webhooks

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/segmentio/ksuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
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

	providedInput := new(types.WebhookCreationRequestInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, providedInput); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = providedInput.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("provided input was invalid")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	input := types.WebhookDatabaseCreationInputFromWebhookCreationInput(providedInput)
	input.ID = ksuid.New().String()
	tracing.AttachWebhookIDToSpan(span, input.ID)
	input.BelongsToAccount = sessionCtxData.ActiveAccountID

	if s.async {
		preWrite := &types.PreWriteMessage{
			DataType:                types.WebhookDataType,
			Webhook:                 input,
			AttributableToUserID:    sessionCtxData.Requester.UserID,
			AttributableToAccountID: sessionCtxData.ActiveAccountID,
		}
		if err = s.preWritesPublisher.Publish(ctx, preWrite); err != nil {
			observability.AcknowledgeError(err, logger, span, "publishing webhook write message")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
			return
		}

		pwr := types.PreWriteResponse{ID: input.ID}

		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, pwr, http.StatusCreated)
	} else {
		// create the webhook.
		wh, webhookCreationErr := s.webhookDataManager.CreateWebhook(ctx, input)
		if webhookCreationErr != nil {
			observability.AcknowledgeError(webhookCreationErr, logger, span, "creating webhook")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
			return
		}

		// notify the relevant parties.
		tracing.AttachWebhookIDToSpan(span, wh.ID)

		// let everybody know we're good.
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, wh, http.StatusCreated)
	}
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

	if s.async {
		exists, webhookExistenceCheckErr := s.webhookDataManager.WebhookExists(ctx, webhookID, sessionCtxData.ActiveAccountID)
		if webhookExistenceCheckErr != nil && !errors.Is(webhookExistenceCheckErr, sql.ErrNoRows) {
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
			observability.AcknowledgeError(webhookExistenceCheckErr, logger, span, "checking item existence")
			return
		} else if !exists || errors.Is(webhookExistenceCheckErr, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
			return
		}

		pam := &types.PreArchiveMessage{
			DataType:                types.WebhookDataType,
			WebhookID:               webhookID,
			AttributableToUserID:    sessionCtxData.Requester.UserID,
			AttributableToAccountID: sessionCtxData.ActiveAccountID,
		}
		if err = s.preArchivesPublisher.Publish(ctx, pam); err != nil {
			observability.AcknowledgeError(err, logger, span, "publishing webhook archive message")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
			return
		}

		// let everybody go home.
		res.WriteHeader(http.StatusNoContent)
	} else {
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

		// let everybody go home.
		res.WriteHeader(http.StatusNoContent)
	}
}

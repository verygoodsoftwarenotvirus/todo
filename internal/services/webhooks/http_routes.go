package webhooks

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

const (
	// WebhookIDURIParamKey is a standard string that we'll use to refer to webhook IDs with.
	WebhookIDURIParamKey = "webhookID"
)

var errInvalidMethod = errors.New("invalid method provided")

// validateWebhook does some validation on a WebhookCreationInput and returns an error if anything runs foul.
func validateWebhook(input *models.WebhookCreationInput) error {
	_, err := url.Parse(input.URL)
	if err != nil {
		return fmt.Errorf("invalid URL provided: %w", err)
	}

	input.Method = strings.ToUpper(input.Method)
	switch input.Method {
	// allowed methods.
	case http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead:
		break
	default:
		return errInvalidMethod
	}

	return nil
}

// CreateHandler is our webhook creation route.
func (s *Service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "webhooks.service.CreateHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// try to pluck the parsed input from the request context.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*models.WebhookCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeNoInputResponse(res)

		return
	}

	input.BelongsToUser = si.UserID

	// ensure everything's on the up-and-up
	if err := validateWebhook(input); err != nil {
		logger.Info("invalid method provided")
		s.encoderDecoder.EncodeErrorResponse(res, err.Error(), http.StatusBadRequest)

		return
	}

	// create the webhook.
	wh, err := s.webhookDataManager.CreateWebhook(ctx, input)
	if err != nil {
		logger.Error(err, "error creating webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)

		return
	}

	// notify the relevant parties.
	tracing.AttachWebhookIDToSpan(span, wh.ID)
	s.webhookCounter.Increment(ctx)
	s.auditLog.LogWebhookCreationEvent(ctx, wh)

	// let everybody know we're good.
	s.encoderDecoder.EncodeResponseWithStatus(res, wh, http.StatusCreated)
}

// ListHandler is our list route.
func (s *Service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "webhooks.service.ListHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// figure out how specific we need to be.
	filter := models.ExtractQueryFilter(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// find the webhooks.
	webhooks, err := s.webhookDataManager.GetWebhooks(ctx, si.UserID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		webhooks = &models.WebhookList{
			Webhooks: []models.Webhook{},
		}
	} else if err != nil {
		logger.Error(err, "error encountered fetching webhooks")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)

		return
	}

	// encode the response.
	s.encoderDecoder.EncodeResponse(res, webhooks)
}

// ReadHandler returns a GET handler that returns an webhook.
func (s *Service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "webhooks.service.ReadHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue("webhook_id", webhookID)

	// fetch the webhook from the database.
	x, err := s.webhookDataManager.GetWebhook(ctx, webhookID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("No rows found in webhook database")
		s.encoderDecoder.EncodeNotFoundResponse(res)

		return
	} else if err != nil {
		logger.Error(err, "Error fetching webhook from webhook database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)

		return
	}

	// encode the response.
	s.encoderDecoder.EncodeResponse(res, x)
}

// UpdateHandler returns a handler that updates an webhook.
func (s *Service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "webhooks.service.UpdateHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue("webhook_id", webhookID)

	// fetch parsed creation input from request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*models.WebhookUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeNoInputResponse(res)

		return
	}

	// fetch the webhook in question.
	wh, err := s.webhookDataManager.GetWebhook(ctx, webhookID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no rows found for webhook")
		s.encoderDecoder.EncodeNotFoundResponse(res)

		return
	} else if err != nil {
		logger.Error(err, "error encountered getting webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)

		return
	}

	// update it.
	changes := wh.Update(input)

	// save the update in the database.
	if err = s.webhookDataManager.UpdateWebhook(ctx, wh); err != nil {
		logger.Error(err, "error encountered updating webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)

		return
	}

	s.auditLog.LogWebhookUpdateEvent(ctx, si.UserID, webhookID, changes)

	// let everybody know we're good.
	s.encoderDecoder.EncodeResponse(res, wh)
}

// ArchiveHandler returns a handler that archives an webhook.
func (s *Service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "webhooks.service.ArchiveHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine relevant user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue("webhook_id", webhookID)

	// do the deed.
	err := s.webhookDataManager.ArchiveWebhook(ctx, webhookID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no rows found for webhook")
		s.encoderDecoder.EncodeNotFoundResponse(res)

		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)

		return
	}

	// let the interested parties know.
	s.webhookCounter.Decrement(ctx)
	s.auditLog.LogWebhookArchiveEvent(ctx, si.UserID, webhookID)

	// let everybody go home.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an item.
func (s *Service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "webhooks.service.AuditEntryHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine item ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue("webhook_id", webhookID)

	x, err := s.auditLog.GetAuditLogEntriesForWebhook(ctx, webhookID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching audit log entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, x)
}

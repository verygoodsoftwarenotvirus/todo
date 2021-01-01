package webhooks

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// WebhookIDURIParamKey is a standard string that we'll use to refer to webhook IDs with.
	WebhookIDURIParamKey = "webhookID"
)

var errInvalidMethod = errors.New("invalid method provided")

// validateWebhook does some validation on a WebhookCreationInput and returns an error if anything runs foul.
func validateWebhook(input *types.WebhookCreationInput) error {
	_, err := url.Parse(input.URL)
	if err != nil {
		return fmt.Errorf("invalid url provided: %w", err)
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
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// try to pluck the parsed input from the request context.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*types.WebhookCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)

		return
	}

	input.BelongsToUser = si.UserID

	// ensure everything's on the up-and-up
	if err := validateWebhook(input); err != nil {
		logger.Info("invalid method provided")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)

		return
	}

	// create the webhook.
	wh, err := s.webhookDataManager.CreateWebhook(ctx, input)
	if err != nil {
		logger.Error(err, "error creating webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)

		return
	}

	// notify the relevant parties.
	tracing.AttachWebhookIDToSpan(span, wh.ID)
	s.webhookCounter.Increment(ctx)
	s.auditLog.LogWebhookCreationEvent(ctx, wh)

	// let everybody know we're good.
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, wh, http.StatusCreated)
}

// ListHandler is our list route.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// figure out how specific we need to be.
	filter := types.ExtractQueryFilter(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// find the webhooks.
	webhooks, err := s.webhookDataManager.GetWebhooks(ctx, si.UserID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		webhooks = &types.WebhookList{
			Webhooks: []types.Webhook{},
		}
	} else if err != nil {
		logger.Error(err, "error encountered fetching webhooks")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)

		return
	}

	// encode the response.
	s.encoderDecoder.EncodeResponse(ctx, res, webhooks)
}

// ReadHandler returns a GET handler that returns an webhook.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)

	// fetch the webhook from the database.
	x, err := s.webhookDataManager.GetWebhook(ctx, webhookID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("No rows found in webhook database")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)

		return
	} else if err != nil {
		logger.Error(err, "Error fetching webhook from webhook database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)

		return
	}

	// encode the response.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// UpdateHandler returns a handler that updates an webhook.
func (s *service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)

	// fetch parsed creation input from request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*types.WebhookUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)

		return
	}

	// fetch the webhook in question.
	wh, err := s.webhookDataManager.GetWebhook(ctx, webhookID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no rows found for webhook")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)

		return
	} else if err != nil {
		logger.Error(err, "error encountered getting webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)

		return
	}

	// update it.
	changes := wh.Update(input)

	// save the update in the database.
	if err = s.webhookDataManager.UpdateWebhook(ctx, wh); err != nil {
		logger.Error(err, "error encountered updating webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)

		return
	}

	s.auditLog.LogWebhookUpdateEvent(ctx, si.UserID, webhookID, changes)

	// let everybody know we're good.
	s.encoderDecoder.EncodeResponse(ctx, res, wh)
}

// ArchiveHandler returns a handler that archives an webhook.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine relevant user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)

	// do the deed.
	err := s.webhookDataManager.ArchiveWebhook(ctx, webhookID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no rows found for webhook")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)

		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting webhook")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)

		return
	}

	// let the interested parties know.
	s.webhookCounter.Decrement(ctx)
	s.auditLog.LogWebhookArchiveEvent(ctx, si.UserID, webhookID)

	// let everybody go home.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an item.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine item ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)

	x, err := s.auditLog.GetAuditLogEntriesForWebhook(ctx, webhookID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching audit log entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

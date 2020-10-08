package webhooks

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	// URIParamKey is a standard string that we'll use to refer to webhook IDs with.
	URIParamKey = "webhookID"
)

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
		return fmt.Errorf("invalid method provided: %q", input.Method)
	}

	return nil
}

// CreateHandler is our webhook creation route.
func (s *Service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "CreateHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// figure out who this is all for.
	userID := s.userIDFetcher(req)
	logger = logger.WithValue("user_id", userID)
	tracing.AttachUserIDToSpan(span, userID)

	// try to pluck the parsed input from the request context.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*models.WebhookCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	input.BelongsToUser = userID

	// ensure everything's on the up-and-up
	if err := validateWebhook(input); err != nil {
		logger.Info("invalid method provided")
		res.WriteHeader(http.StatusBadRequest)
		s.encoderDecoder.EncodeError(res, err.Error(), http.StatusBadRequest)
		return
	}

	// create the webhook.
	wh, err := s.webhookDataManager.CreateWebhook(ctx, input)
	if err != nil {
		logger.Error(err, "error creating webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// notify the relevant parties.
	tracing.AttachWebhookIDToSpan(span, wh.ID)
	s.webhookCounter.Increment(ctx)
	s.eventManager.Report(newsman.Event{
		EventType: string(models.Create),
		Data:      wh,
		Topics:    []string{topicName},
	})

	l := wh.ToListener(logger)
	s.eventManager.TuneIn(l)

	// let everybody know we're good.
	res.WriteHeader(http.StatusCreated)
	s.encoderDecoder.EncodeResponse(res, wh)
}

// ListHandler is our list route.
func (s *Service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ListHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// figure out how specific we need to be.
	filter := models.ExtractQueryFilter(req)

	// figure out who this is all for.
	userID := s.userIDFetcher(req)
	logger = logger.WithValue("user_id", userID)
	tracing.AttachUserIDToSpan(span, userID)

	// find the webhooks.
	webhooks, err := s.webhookDataManager.GetWebhooks(ctx, userID, filter)
	if err == sql.ErrNoRows {
		webhooks = &models.WebhookList{
			Webhooks: []models.Webhook{},
		}
	} else if err != nil {
		logger.Error(err, "error encountered fetching webhooks")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// encode the response.
	s.encoderDecoder.EncodeResponse(res, webhooks)
}

// ReadHandler returns a GET handler that returns an webhook.
func (s *Service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ReadHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine relevant user ID.
	userID := s.userIDFetcher(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue("user_id", userID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue("webhook_id", webhookID)

	// fetch the webhook from the database.
	x, err := s.webhookDataManager.GetWebhook(ctx, webhookID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("No rows found in webhook database")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "Error fetching webhook from webhook database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// encode the response.
	s.encoderDecoder.EncodeResponse(res, x)
}

// UpdateHandler returns a handler that updates an webhook.
func (s *Service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "UpdateHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine relevant user ID.
	userID := s.userIDFetcher(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue("user_id", userID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue("webhook_id", webhookID)

	// fetch parsed creation input from request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*models.WebhookUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch the webhook in question.
	wh, err := s.webhookDataManager.GetWebhook(ctx, webhookID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no rows found for webhook")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// update it.
	wh.Update(input)

	// save the update in the database.
	if err = s.webhookDataManager.UpdateWebhook(ctx, wh); err != nil {
		logger.Error(err, "error encountered updating webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// notify the relevant parties.
	s.eventManager.Report(newsman.Event{
		EventType: string(models.Update),
		Data:      wh,
		Topics:    []string{topicName},
	})

	// let everybody know we're good.
	s.encoderDecoder.EncodeResponse(res, wh)
}

// ArchiveHandler returns a handler that archives an webhook.
func (s *Service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "delete_route")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine relevant user ID.
	userID := s.userIDFetcher(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue("user_id", userID)

	// determine relevant webhook ID.
	webhookID := s.webhookIDFetcher(req)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	logger = logger.WithValue("webhook_id", webhookID)

	// do the deed.
	err := s.webhookDataManager.ArchiveWebhook(ctx, webhookID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no rows found for webhook")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// let the interested parties know.
	s.webhookCounter.Decrement(ctx)
	s.eventManager.Report(newsman.Event{
		EventType: string(models.Archive),
		Data:      models.Webhook{ID: webhookID},
		Topics:    []string{topicName},
	})

	// let everybody go home.
	res.WriteHeader(http.StatusNoContent)
}

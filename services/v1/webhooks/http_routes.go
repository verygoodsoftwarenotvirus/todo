package webhooks

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"go.opencensus.io/trace"
)

const (
	// URIParamKey is a standard string that we'll use to refer to webhook IDs with
	URIParamKey = "webhookID"
)

// attachWebhookIDToSpan provides a consistent way to attach a webhook ID to a given span
func attachWebhookIDToSpan(span *trace.Span, webhookID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("webhook_id", strconv.FormatUint(webhookID, 10)))
	}
}

// attachUserIDToSpan provides a consistent way to attach a user ID to a given span
func attachUserIDToSpan(span *trace.Span, userID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))
	}
}

// ListHandler is our list route
func (s *Service) ListHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ListHandler")
		defer span.End()

		// figure out how specific we need to be
		qf := models.ExtractQueryFilter(req)

		// figure out who this is all for
		userID := s.userIDFetcher(req)
		logger := s.logger.WithValue("user_id", userID)
		attachUserIDToSpan(span, userID)

		// find the webhooks
		webhooks, err := s.webhookDatabase.GetWebhooks(ctx, qf, userID)
		if err == sql.ErrNoRows {
			webhooks = &models.WebhookList{
				Webhooks: []models.Webhook{},
			}
		} else if err != nil {
			logger.Error(err, "error encountered fetching webhooks")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode the response
		if err = s.encoderDecoder.EncodeResponse(res, webhooks); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// validateWebhook does some validation on a WebhookCreationInput and returns an error if anything runs foul
func validateWebhook(input *models.WebhookCreationInput) error {
	_, err := url.Parse(input.URL)
	if err != nil {
		return fmt.Errorf("invalid URL provided: %w", err)
	}

	input.Method = strings.ToUpper(input.Method)
	switch input.Method {
	// allowed methods
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

// CreateHandler is our webhook creation route
func (s *Service) CreateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "CreateHandler")
		defer span.End()

		// figure out who this is all for
		userID := s.userIDFetcher(req)
		logger := s.logger.WithValue("user", userID)
		attachUserIDToSpan(span, userID)

		// try to pluck the parsed input from the request context
		input, ok := ctx.Value(CreateMiddlewareCtxKey).(*models.WebhookCreationInput)
		if !ok {
			logger.Info("valid input not attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		input.BelongsTo = userID

		// ensure everythings on the up-and-up
		if err := validateWebhook(input); err != nil {
			logger.Info("invalid method provided")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// create the webhook
		wh, err := s.webhookDatabase.CreateWebhook(ctx, input)
		if err != nil {
			logger.Error(err, "error creating webhook")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify the relevant parties
		attachWebhookIDToSpan(span, wh.ID)
		s.webhookCounter.Increment(ctx)
		s.eventManager.Report(newsman.Event{
			EventType: string(models.Create),
			Data:      wh,
			Topics:    []string{topicName},
		})

		l := wh.ToListener(s.logger)
		s.eventManager.TuneIn(l)

		// let everybody know we're good
		res.WriteHeader(http.StatusCreated)
		if err = s.encoderDecoder.EncodeResponse(res, wh); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// ReadHandler returns a GET handler that returns an webhook
func (s *Service) ReadHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ReadHandler")
		defer span.End()

		// figure out what this is for and who it belongs to
		userID := s.userIDFetcher(req)
		webhookID := s.webhookIDFetcher(req)

		// document it for posterity
		attachUserIDToSpan(span, userID)
		attachWebhookIDToSpan(span, webhookID)
		logger := s.logger.WithValues(map[string]interface{}{
			"user":    userID,
			"webhook": webhookID,
		})

		// fetch the webhook from the database
		x, err := s.webhookDatabase.GetWebhook(ctx, webhookID, userID)
		if err == sql.ErrNoRows {
			logger.Debug("No rows found in webhookDatabase")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "Error fetching webhook from webhookDatabase")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode the response
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// UpdateHandler returns a handler that updates an webhook
func (s *Service) UpdateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "UpdateHandler")
		defer span.End()

		// figure out what this is for and who it belongs to
		userID := s.userIDFetcher(req)
		webhookID := s.webhookIDFetcher(req)

		// document it for posterity
		attachUserIDToSpan(span, userID)
		attachWebhookIDToSpan(span, webhookID)
		logger := s.logger.WithValues(map[string]interface{}{
			"user_id":    userID,
			"webhook_id": webhookID,
		})

		// fetch parsed creation input from request context
		input, ok := ctx.Value(UpdateMiddlewareCtxKey).(*models.WebhookUpdateInput)
		if !ok {
			s.logger.Info("no input attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// fetch the webhook in question
		wh, err := s.webhookDatabase.GetWebhook(ctx, webhookID, userID)
		if err == sql.ErrNoRows {
			logger.Debug("no rows found for webhook")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error encountered getting webhook")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// update it
		wh.Update(input)

		// save the update in the database
		if err = s.webhookDatabase.UpdateWebhook(ctx, wh); err != nil {
			logger.Error(err, "error encountered updating webhook")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify the relevant parties
		s.eventManager.Report(newsman.Event{
			EventType: string(models.Update),
			Data:      wh,
			Topics:    []string{topicName},
		})

		// let everybody know we're good
		if err = s.encoderDecoder.EncodeResponse(res, wh); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// ArchiveHandler returns a handler that archives an webhook
func (s *Service) ArchiveHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "delete_route")
		defer span.End()

		// figure out what this is for and who it belongs to
		userID := s.userIDFetcher(req)
		webhookID := s.webhookIDFetcher(req)

		// document it for posterity
		attachUserIDToSpan(span, userID)
		attachWebhookIDToSpan(span, webhookID)
		logger := s.logger.WithValues(map[string]interface{}{
			"webhook_id": webhookID,
			"user_id":    userID,
		})

		// do the deed
		err := s.webhookDatabase.ArchiveWebhook(ctx, webhookID, userID)
		if err == sql.ErrNoRows {
			logger.Debug("no rows found for webhook")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error encountered deleting webhook")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// let the interested parties know
		s.webhookCounter.Decrement(ctx)
		s.eventManager.Report(newsman.Event{
			EventType: string(models.Archive),
			Data:      models.Webhook{ID: webhookID},
			Topics:    []string{topicName},
		})

		// let everybody go home
		res.WriteHeader(http.StatusNoContent)
	}
}

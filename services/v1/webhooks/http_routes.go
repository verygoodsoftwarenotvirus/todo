package webhooks

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	// URIParamKey is a standard string that we'll use to refer to webhook IDs with
	URIParamKey = "webhookID"
)

func attachWebhookIDToSpan(span *trace.Span, webhookID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("webhook_id", strconv.FormatUint(webhookID, 10)))
	}
}

func attachUserIDToSpan(span *trace.Span, userID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))
	}
}

// ListHandler is our list route
func (s *Service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "ListHandler")
	defer span.End()

	qf := models.ExtractQueryFilter(req)
	userID := s.userIDFetcher(req)
	logger := s.logger.WithValue("user_id", userID)
	attachUserIDToSpan(span, userID)

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

	if err = s.encoderDecoder.EncodeResponse(res, webhooks); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

func validateWebhook(input *models.WebhookInput) (statusCode int, err error) {
	_, err = url.Parse(input.URL)
	if err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "invalid URL provided")
	}

	input.Method = strings.ToUpper(input.Method)
	switch input.Method {
	case http.MethodGet, // allowed methods
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodHead:
		break
	default:
		return http.StatusBadRequest, errors.Wrap(nil, "invalid method provided")
	}

	return http.StatusOK, nil
}

// CreateHandler is our webhook creation route
func (s *Service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "CreateHandler")
	defer span.End()

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValue("user", userID)
	attachUserIDToSpan(span, userID)

	input, ok := ctx.Value(MiddlewareCtxKey).(*models.WebhookInput)
	if !ok {
		logger.Info("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	input.BelongsTo = userID

	code, err := validateWebhook(input)
	if code != http.StatusOK && err != nil {
		logger.Info("invalid method provided")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	wh, err := s.webhookDatabase.CreateWebhook(ctx, input)
	if err != nil {
		logger.Error(err, "error creating webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	attachWebhookIDToSpan(span, wh.ID)
	s.webhookCounter.Increment(ctx)

	s.eventManager.Report(newsman.Event{
		EventType: string(models.Create),
		Data:      wh,
		Topics:    []string{topicName},
	})

	l := wh.ToListener(s.logger)
	s.eventManager.TuneIn(l)

	res.WriteHeader(http.StatusCreated)
	if err = s.encoderDecoder.EncodeResponse(res, wh); err != nil {
		logger.Error(err, "encoding response")
	}
}

// ReadHandler returns a GET handler that returns an webhook
func (s *Service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "ReadHandler")
	defer span.End()

	userID := s.userIDFetcher(req)
	webhookID := s.webhookIDFetcher(req)

	attachUserIDToSpan(span, userID)
	attachWebhookIDToSpan(span, webhookID)

	logger := s.logger.WithValues(map[string]interface{}{
		"user":    userID,
		"webhook": webhookID,
	})

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

	if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
		logger.Error(err, "encoding response")
	}
}

// UpdateHandler returns a handler that updates an webhook
func (s *Service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "UpdateHandler")
	defer span.End()

	input, ok := ctx.Value(MiddlewareCtxKey).(*models.WebhookInput)
	if !ok {
		s.logger.Info("no input attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := s.userIDFetcher(req)
	webhookID := s.webhookIDFetcher(req)

	attachUserIDToSpan(span, userID)
	attachWebhookIDToSpan(span, webhookID)

	logger := s.logger.WithValues(map[string]interface{}{
		"user_id":    userID,
		"webhook_id": webhookID,
	})

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

	wh.Update(input)
	if err = s.webhookDatabase.UpdateWebhook(ctx, wh); err != nil {
		logger.Error(err, "error encountered updating webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.eventManager.Report(newsman.Event{
		EventType: string(models.Update),
		Data:      wh,
		Topics:    []string{topicName},
	})

	if err = s.encoderDecoder.EncodeResponse(res, wh); err != nil {
		logger.Error(err, "encoding response")
	}
}

// ArchiveHandler returns a handler that archives an webhook
func (s *Service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "delete_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	webhookID := s.webhookIDFetcher(req)

	attachUserIDToSpan(span, userID)
	attachWebhookIDToSpan(span, webhookID)

	logger := s.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	})

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
	s.webhookCounter.Decrement(ctx)

	s.eventManager.Report(newsman.Event{
		EventType: string(models.Archive),
		Data:      models.Webhook{ID: webhookID},
		Topics:    []string{topicName},
	})

	res.WriteHeader(http.StatusNoContent)
}

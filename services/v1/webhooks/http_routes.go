package webhooks

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	v1 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/events/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	// URIParamKey is a standard string that we'll use to refer to webhook IDs with
	URIParamKey = "webhookID"
)

// List is our list route
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "list_route")
	defer span.End()

	s.logger.Debug("WebhooksService.List called")
	qf := models.ExtractQueryFilter(req)

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"filter":  qf,
		"user_id": userID,
	})

	webhooks, err := s.webhookDatabase.GetWebhooks(ctx, qf, userID)
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching webhooks")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, webhooks); err != nil {
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
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead:
		break
	default:
		return http.StatusBadRequest, errors.Wrap(nil, "invalid method provided")
	}

	return http.StatusOK, nil
}

// Create is our webhook creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "create_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValue("user_id", userID)
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	logger.Debug("create route called")
	input, ok := ctx.Value(MiddlewareCtxKey).(*models.WebhookInput)
	if !ok {
		s.logger.Error(nil, "valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	input.BelongsTo = userID
	logger = logger.WithValue("input", input)

	code, err := validateWebhook(input)
	if code != http.StatusOK && err != nil {
		logger.Error(nil, "invalid method provided")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	x, err := s.webhookDatabase.CreateWebhook(ctx, input)
	if err != nil {
		s.logger.Error(err, "error creating webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.webhookCounter.Increment(ctx)

	s.newsman.Report(newsman.Event{
		EventType: string(v1.Create),
		Data:      x,
		Topics:    []string{topicName},
	})

	l := x.ToListener(s.logger)
	s.newsman.TuneIn(l)

	res.WriteHeader(http.StatusCreated)
	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Read returns a GET handler that returns an webhook
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "read_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	webhookID := s.webhookIDFetcher(req)

	logger := s.logger.WithValues(map[string]interface{}{
		"user_id":    userID,
		"webhook_id": webhookID,
	})
	logger.Debug("webhooksService.ReadHandler called")

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

	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Update returns a handler that updates an webhook
func (s *Service) Update(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "update_route")
	defer span.End()

	input, ok := ctx.Value(MiddlewareCtxKey).(*models.WebhookInput)
	if !ok {
		s.logger.Error(nil, "no input attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := s.userIDFetcher(req)
	webhookID := s.webhookIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"user_id":    userID,
		"webhook_id": webhookID,
		"input":      input,
	})

	x, err := s.webhookDatabase.GetWebhook(ctx, webhookID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no rows found for webhook")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	x.Update(input)
	if err = s.webhookDatabase.UpdateWebhook(ctx, x); err != nil {
		logger.Error(err, "error encountered updating webhook")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.newsman.Report(newsman.Event{
		EventType: string(v1.Update),
		Data:      x,
		Topics:    []string{topicName},
	})

	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Delete returns a handler that deletes an webhook
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "delete_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	webhookID := s.webhookIDFetcher(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	})
	logger.Debug("WebhooksService Deletion handler called")

	err := s.webhookDatabase.DeleteWebhook(ctx, webhookID, userID)
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

	s.newsman.Report(newsman.Event{
		EventType: string(v1.Delete),
		Data:      models.Webhook{ID: webhookID},
		Topics:    []string{topicName},
	})

	res.WriteHeader(http.StatusNoContent)
}

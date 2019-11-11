package models

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

type (
	// WebhookDataManager describes a structure capable of storing webhooks
	WebhookDataManager interface {
		GetWebhook(ctx context.Context, webhookID, userID uint64) (*Webhook, error)
		GetWebhookCount(ctx context.Context, filter *QueryFilter, userID uint64) (uint64, error)
		GetAllWebhooksCount(ctx context.Context) (uint64, error)
		GetWebhooks(ctx context.Context, filter *QueryFilter, userID uint64) (*WebhookList, error)
		GetAllWebhooks(ctx context.Context) (*WebhookList, error)
		GetAllWebhooksForUser(ctx context.Context, userID uint64) ([]Webhook, error)
		CreateWebhook(ctx context.Context, input *WebhookCreationInput) (*Webhook, error)
		UpdateWebhook(ctx context.Context, updated *Webhook) error
		ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error
	}

	// WebhookDataServer describes a structure capable of serving traffic related to webhooks
	WebhookDataServer interface {
		CreationInputMiddleware(next http.Handler) http.Handler
		UpdateInputMiddleware(next http.Handler) http.Handler

		ListHandler() http.HandlerFunc
		CreateHandler() http.HandlerFunc
		ReadHandler() http.HandlerFunc
		UpdateHandler() http.HandlerFunc
		ArchiveHandler() http.HandlerFunc
	}

	// Webhook represents a webhook listener, an endpoint to send an HTTP request to upon an event
	Webhook struct {
		ID          uint64   `json:"id"`
		Name        string   `json:"name"`
		ContentType string   `json:"content_type"`
		URL         string   `json:"url"`
		Method      string   `json:"method"`
		Events      []string `json:"events"`
		DataTypes   []string `json:"data_types"`
		Topics      []string `json:"topics"`
		CreatedOn   uint64   `json:"created_on"`
		UpdatedOn   *uint64  `json:"updated_on"`
		ArchivedOn  *uint64  `json:"archived_on"`
		BelongsTo   uint64   `json:"belongs_to"`
	}

	// WebhookCreationInput represents what a user could set as input for creating a webhook
	WebhookCreationInput struct {
		Name        string   `json:"name"`
		ContentType string   `json:"content_type"`
		URL         string   `json:"url"`
		Method      string   `json:"method"`
		Events      []string `json:"events"`
		DataTypes   []string `json:"data_types"`
		Topics      []string `json:"topics"`
		BelongsTo   uint64   `json:"-"`
	}

	// WebhookUpdateInput represents what a user could set as input for updating a webhook
	WebhookUpdateInput struct {
		Name        string   `json:"name"`
		ContentType string   `json:"content_type"`
		URL         string   `json:"url"`
		Method      string   `json:"method"`
		Events      []string `json:"events"`
		DataTypes   []string `json:"data_types"`
		Topics      []string `json:"topics"`
		BelongsTo   uint64   `json:"-"`
	}

	// WebhookList represents a list of webhooks
	WebhookList struct {
		Pagination
		Webhooks []Webhook `json:"webhooks"`
	}
)

// Update merges an WebhookCreationInput with an Webhook
func (w *Webhook) Update(input *WebhookUpdateInput) {
	if input.Name != "" {
		w.Name = input.Name
	}
	if input.ContentType != "" {
		w.ContentType = input.ContentType
	}
	if input.URL != "" {
		w.URL = input.URL
	}
	if input.Method != "" {
		w.Method = input.Method
	}

	if input.Events != nil && len(input.Events) > 0 {
		w.Events = input.Events
	}
	if input.DataTypes != nil && len(input.DataTypes) > 0 {
		w.DataTypes = input.DataTypes
	}
	if input.Topics != nil && len(input.Topics) > 0 {
		w.Topics = input.Topics
	}
}

func buildErrorLogFunc(w *Webhook, logger logging.Logger) func(error) {
	return func(err error) {
		logger.WithValues(map[string]interface{}{
			"url":          w.URL,
			"method":       w.Method,
			"content_type": w.ContentType,
		}).Error(err, "error executing webhook")
	}
}

// ToListener creates a newsman Listener from a Webhook
func (w *Webhook) ToListener(logger logging.Logger) newsman.Listener {
	return newsman.NewWebhookListener(
		buildErrorLogFunc(w, logger),
		&newsman.WebhookConfig{
			Method:      w.Method,
			URL:         w.URL,
			ContentType: w.ContentType,
		},
		&newsman.ListenerConfig{
			Events:    w.Events,
			DataTypes: w.DataTypes,
			Topics:    w.Topics,
		},
	)
}

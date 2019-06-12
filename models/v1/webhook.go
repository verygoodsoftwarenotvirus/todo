package models

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
)

// WebhookDataManager describes a structure capable of storing items permanently
type WebhookDataManager interface {
	GetWebhook(ctx context.Context, itemID, userID uint64) (*Webhook, error)
	GetWebhookCount(ctx context.Context, filter *QueryFilter, userID uint64) (uint64, error)
	GetAllWebhooksCount(ctx context.Context) (uint64, error)
	GetWebhooks(ctx context.Context, filter *QueryFilter, userID uint64) (*WebhookList, error)
	GetAllWebhooks(ctx context.Context) (*WebhookList, error)
	CreateWebhook(ctx context.Context, input *WebhookInput) (*Webhook, error)
	UpdateWebhook(ctx context.Context, updated *Webhook) error
	DeleteWebhook(ctx context.Context, id, userID uint64) error
}

// WebhookDataServer describes a structure capable of serving traffic related to items
type WebhookDataServer interface {
	CreationInputMiddleware(next http.Handler) http.Handler
	UpdateInputMiddleware(next http.Handler) http.Handler

	List(res http.ResponseWriter, req *http.Request)
	Create(res http.ResponseWriter, req *http.Request)
	Read(res http.ResponseWriter, req *http.Request)
	Update(res http.ResponseWriter, req *http.Request)
	Delete(res http.ResponseWriter, req *http.Request)
}

// Webhook represents an item
type Webhook struct {
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

// WebhookInput represents what a user could set as input for items
type WebhookInput struct {
	Name        string   `json:"name"`
	ContentType string   `json:"content_type"`
	URL         string   `json:"url"`
	Method      string   `json:"method"`
	Events      []string `json:"events"`
	DataTypes   []string `json:"data_types"`
	Topics      []string `json:"topics"`
	BelongsTo   uint64   `json:"-"`
}

// Update merges an WebhookInput with an Webhook
func (w *Webhook) Update(input *WebhookInput) {
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

// WebhookList represents a list of items
type WebhookList struct {
	Pagination
	Webhooks []Webhook `json:"items"`
}

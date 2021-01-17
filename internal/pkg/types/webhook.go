package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

type (
	// Webhook represents a webhook listener, an endpoint to send an HTTP request to upon an event.
	Webhook struct {
		ID            uint64   `json:"id"`
		Name          string   `json:"name"`
		ContentType   string   `json:"contentType"`
		URL           string   `json:"url"`
		Method        string   `json:"method"`
		Events        []string `json:"events"`
		DataTypes     []string `json:"dataTypes"`
		Topics        []string `json:"topics"`
		CreatedOn     uint64   `json:"createdOn"`
		LastUpdatedOn *uint64  `json:"lastUpdatedOn"`
		ArchivedOn    *uint64  `json:"archivedOn"`
		BelongsToUser uint64   `json:"belongsToUser"`
	}

	// WebhookCreationInput represents what a User could set as input for creating a webhook.
	WebhookCreationInput struct {
		Name          string   `json:"name"`
		ContentType   string   `json:"contentType"`
		URL           string   `json:"url"`
		Method        string   `json:"method"`
		Events        []string `json:"events"`
		DataTypes     []string `json:"dataTypes"`
		Topics        []string `json:"topics"`
		BelongsToUser uint64   `json:"-"`
	}

	// WebhookUpdateInput represents what a User could set as input for updating a webhook.
	WebhookUpdateInput struct {
		Name          string   `json:"name"`
		ContentType   string   `json:"contentType"`
		URL           string   `json:"url"`
		Method        string   `json:"method"`
		Events        []string `json:"events"`
		DataTypes     []string `json:"dataTypes"`
		Topics        []string `json:"topics"`
		BelongsToUser uint64   `json:"-"`
	}

	// WebhookList represents a list of webhooks.
	WebhookList struct {
		Pagination
		Webhooks []*Webhook `json:"webhooks"`
	}

	// WebhookSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	WebhookSQLQueryBuilder interface {
		BuildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{})
		BuildGetAllWebhooksCountQuery() string
		BuildGetBatchOfWebhooksQuery(beginID, endID uint64) (query string, args []interface{})
		BuildGetWebhooksQuery(userID uint64, filter *QueryFilter) (query string, args []interface{})
		BuildCreateWebhookQuery(x *Webhook) (query string, args []interface{})
		BuildUpdateWebhookQuery(input *Webhook) (query string, args []interface{})
		BuildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{})
	}

	// WebhookDataManager describes a structure capable of storing webhooks.
	WebhookDataManager interface {
		GetWebhook(ctx context.Context, webhookID, userID uint64) (*Webhook, error)
		GetAllWebhooksCount(ctx context.Context) (uint64, error)
		GetWebhooks(ctx context.Context, userID uint64, filter *QueryFilter) (*WebhookList, error)
		GetAllWebhooks(ctx context.Context, resultChannel chan []*Webhook, bucketSize uint16) error
		CreateWebhook(ctx context.Context, input *WebhookCreationInput) (*Webhook, error)
		UpdateWebhook(ctx context.Context, updated *Webhook) error
		ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error
	}

	// WebhookAuditManager describes a structure capable of .
	WebhookAuditManager interface {
		GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]*AuditLogEntry, error)
		LogWebhookCreationEvent(ctx context.Context, webhook *Webhook)
		LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []FieldChangeSummary)
		LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64)
	}

	// WebhookDataService describes a structure capable of serving traffic related to webhooks.
	WebhookDataService interface {
		CreationInputMiddleware(next http.Handler) http.Handler
		UpdateInputMiddleware(next http.Handler) http.Handler

		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an WebhookCreationInput with an Webhook.
func (w *Webhook) Update(input *WebhookUpdateInput) []FieldChangeSummary {
	changes := []FieldChangeSummary{}

	if input.Name != "" {
		changes = append(changes, FieldChangeSummary{
			FieldName: "Name",
			OldValue:  w.Name,
			NewValue:  input.Name,
		})
		w.Name = input.Name
	}

	if input.ContentType != "" {
		changes = append(changes, FieldChangeSummary{
			FieldName: "ContentType",
			OldValue:  w.ContentType,
			NewValue:  input.ContentType,
		})
		w.ContentType = input.ContentType
	}

	if input.URL != "" {
		changes = append(changes, FieldChangeSummary{
			FieldName: "url",
			OldValue:  w.URL,
			NewValue:  input.URL,
		})
		w.URL = input.URL
	}

	if input.Method != "" {
		changes = append(changes, FieldChangeSummary{
			FieldName: "Method",
			OldValue:  w.Method,
			NewValue:  input.Method,
		})
		w.Method = input.Method
	}

	if input.Events != nil && len(input.Events) > 0 {
		changes = append(changes, FieldChangeSummary{
			FieldName: "Events",
			OldValue:  w.Events,
			NewValue:  input.Events,
		})
		w.Events = input.Events
	}

	if input.DataTypes != nil && len(input.DataTypes) > 0 {
		changes = append(changes, FieldChangeSummary{
			FieldName: "DataTypes",
			OldValue:  w.DataTypes,
			NewValue:  input.DataTypes,
		})
		w.DataTypes = input.DataTypes
	}

	if input.Topics != nil && len(input.Topics) > 0 {
		changes = append(changes, FieldChangeSummary{
			FieldName: "Topics",
			OldValue:  w.Topics,
			NewValue:  input.Topics,
		})
		w.Topics = input.Topics
	}

	return changes
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

// Validate validates a WebhookCreationInput.
func (w *WebhookCreationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.URL, validation.Required, &urlValidator{}),
		validation.Field(&w.Method, validation.Required, validation.In(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete)),
		validation.Field(&w.ContentType, validation.Required, validation.In("application/json", "application/xml")),
		validation.Field(&w.Events, validation.Required),
		validation.Field(&w.DataTypes, validation.Required),
	)
}

// Validate validates a WebhookUpdateInput.
func (w *WebhookUpdateInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.URL, validation.Required, &urlValidator{}),
		validation.Field(&w.Method, validation.Required, validation.In(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete)),
		validation.Field(&w.ContentType, validation.Required, validation.In("application/json", "application/xml")),
		validation.Field(&w.Events, validation.Required),
		validation.Field(&w.DataTypes, validation.Required),
	)
}

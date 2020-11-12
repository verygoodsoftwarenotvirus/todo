package types

import (
	"context"
	"net/http"

	v "github.com/RussellLuo/validating/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
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

	// WebhookCreationInput represents what a user could set as input for creating a webhook.
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

	// WebhookUpdateInput represents what a user could set as input for updating a webhook.
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
		Webhooks []Webhook `json:"webhooks"`
	}

	// WebhookDataManager describes a structure capable of storing webhooks.
	WebhookDataManager interface {
		GetWebhook(ctx context.Context, webhookID, userID uint64) (*Webhook, error)
		GetAllWebhooksCount(ctx context.Context) (uint64, error)
		GetWebhooks(ctx context.Context, userID uint64, filter *QueryFilter) (*WebhookList, error)
		GetAllWebhooks(ctx context.Context) (*WebhookList, error)
		CreateWebhook(ctx context.Context, input *WebhookCreationInput) (*Webhook, error)
		UpdateWebhook(ctx context.Context, updated *Webhook) error
		ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error
	}

	// WebhookAuditManager describes a structure capable of .
	WebhookAuditManager interface {
		GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]AuditLogEntry, error)
		LogWebhookCreationEvent(ctx context.Context, webhook *Webhook)
		LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []FieldChangeSummary)
		LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64)
	}

	// WebhookDataServer describes a structure capable of serving traffic related to webhooks.
	WebhookDataServer interface {
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
			FieldName: "URL",
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

// ToListener creates a newsman Listener from a Webhook.
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

// Validate validates a WebhookCreationInput.
func (w *WebhookCreationInput) Validate() error {
	err := v.Validate(v.Schema{
		v.F("name", w.Name):                &minimumStringLengthValidator{minLength: 1},
		v.F("url", w.URL):                  &urlValidator{},
		v.F("method", &w.Method):           v.In(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete),
		v.F("contentType", &w.ContentType): v.In("application/json", "application/xml"),
		v.F("events", &w.Events):           &minimumStringSliceLengthValidator{minLength: 1},
		v.F("dataTypes", &w.DataTypes):     &minimumStringSliceLengthValidator{minLength: 1},
		v.F("topics", &w.Topics):           &minimumStringSliceLengthValidator{minLength: 1},
	})

	// for whatever reason, returning straight from v.Validate makes my tests fail /shrug
	if err != nil {
		return err
	}

	return nil
}

// Validate validates a WebhookUpdateInput.
func (w *WebhookUpdateInput) Validate() error {
	err := v.Validate(v.Schema{
		v.F("name", w.Name):                &minimumStringLengthValidator{minLength: 1},
		v.F("url", w.URL):                  &urlValidator{},
		v.F("method", &w.Method):           v.In(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete),
		v.F("contentType", &w.ContentType): v.In("application/json", "application/xml"),
		v.F("events", &w.Events):           &minimumStringSliceLengthValidator{minLength: 1},
		v.F("dataTypes", &w.DataTypes):     &minimumStringSliceLengthValidator{minLength: 1},
		v.F("topics", &w.Topics):           &minimumStringSliceLengthValidator{minLength: 1},
	})

	// for whatever reason, returning straight from v.Validate makes my tests fail /shrug
	if err != nil {
		return err
	}

	return nil
}

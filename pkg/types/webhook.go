package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Webhook represents a webhook listener, an endpoint to send an HTTP request to upon an event.
	Webhook struct {
		_ struct{}

		LastUpdatedOn    *uint64  `json:"lastUpdatedOn"`
		ArchivedOn       *uint64  `json:"archivedOn"`
		Name             string   `json:"name"`
		URL              string   `json:"url"`
		Method           string   `json:"method"`
		ContentType      string   `json:"contentType"`
		ID               string   `json:"id"`
		BelongsToAccount string   `json:"belongsToAccount"`
		Events           []string `json:"events"`
		DataTypes        []string `json:"dataTypes"`
		Topics           []string `json:"topics"`
		CreatedOn        uint64   `json:"createdOn"`
	}

	// WebhookCreationInput represents what a User could set as input for creating a webhook.
	WebhookCreationInput struct {
		_ struct{}

		ID               string   `json:"-"`
		Name             string   `json:"name"`
		ContentType      string   `json:"contentType"`
		URL              string   `json:"url"`
		Method           string   `json:"method"`
		BelongsToAccount string   `json:"-"`
		Events           []string `json:"events"`
		DataTypes        []string `json:"dataTypes"`
		Topics           []string `json:"topics"`
	}

	// WebhookList represents a list of webhooks.
	WebhookList struct {
		_ struct{}

		Webhooks []*Webhook `json:"webhooks"`
		Pagination
	}

	// WebhookDataManager describes a structure capable of storing webhooks.
	WebhookDataManager interface {
		GetWebhook(ctx context.Context, webhookID, accountID string) (*Webhook, error)
		GetAllWebhooksCount(ctx context.Context) (uint64, error)
		GetWebhooks(ctx context.Context, accountID string, filter *QueryFilter) (*WebhookList, error)
		CreateWebhook(ctx context.Context, input *WebhookCreationInput) (*Webhook, error)
		ArchiveWebhook(ctx context.Context, webhookID, accountID string) error
	}

	// WebhookDataService describes a structure capable of serving traffic related to webhooks.
	WebhookDataService interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

var _ validation.ValidatableWithContext = (*WebhookCreationInput)(nil)

// ValidateWithContext validates a WebhookCreationInput.
func (w *WebhookCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.URL, validation.Required, &urlValidator{}),
		validation.Field(&w.Method, validation.Required, validation.In(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete)),
		validation.Field(&w.ContentType, validation.Required, validation.In("application/json", "application/xml")),
		validation.Field(&w.Events, validation.Required),
		validation.Field(&w.DataTypes, validation.Required),
	)
}

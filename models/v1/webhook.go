package models

import (
	"context"
	"net/http"
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
	DeleteWebhook(ctx context.Context, id uint64, userID uint64) error
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
	CompletedOn *uint64  `json:"completed_on"`
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

func sliceToMap(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// Update merges an WebhookInput with an Webhook
func (i *Webhook) Update(input *WebhookInput) {
	if input.Name != "" {
		i.Name = input.Name
	}
	if input.ContentType != "" {
		i.ContentType = input.ContentType
	}
	if input.URL != "" {
		i.URL = input.URL
	}
	if input.Method != "" {
		i.Method = input.Method
	}

	if input.Events != nil && len(input.Events) > 0 {
		i.Events = input.Events
	}

	if input.DataTypes != nil && len(input.DataTypes) > 0 {
		i.DataTypes = input.DataTypes
	}

	if input.Topics != nil && len(input.Topics) > 0 {
		i.Topics = input.Topics
	}

}

// WebhookList represents a list of items
type WebhookList struct {
	Pagination
	Webhooks []Webhook `json:"items"`
}

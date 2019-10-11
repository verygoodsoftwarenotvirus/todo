package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	webhooksBasePath = "webhooks"
)

// BuildGetWebhookRequest builds an HTTP request for fetching a webhook
func (c *V1Client) BuildGetWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetWebhook retrieves a webhook
func (c *V1Client) GetWebhook(ctx context.Context, id uint64) (webhook *models.Webhook, err error) {
	req, err := c.BuildGetWebhookRequest(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &webhook)
	return webhook, err
}

// BuildGetWebhooksRequest builds an HTTP request for fetching webhooks
func (c *V1Client) BuildGetWebhooksRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	uri := c.BuildURL(filter.ToValues(), webhooksBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetWebhooks gets a list of webhooks
func (c *V1Client) GetWebhooks(ctx context.Context, filter *models.QueryFilter) (webhooks *models.WebhookList, err error) {
	req, err := c.BuildGetWebhooksRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &webhooks)
	return webhooks, err
}

// BuildCreateWebhookRequest builds an HTTP request for creating a webhook
func (c *V1Client) BuildCreateWebhookRequest(ctx context.Context, body *models.WebhookCreationInput) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath)

	return c.buildDataRequest(http.MethodPost, uri, body)
}

// CreateWebhook creates a webhook
func (c *V1Client) CreateWebhook(ctx context.Context, input *models.WebhookCreationInput) (webhook *models.Webhook, err error) {
	req, err := c.BuildCreateWebhookRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &webhook)
	return webhook, err
}

// BuildUpdateWebhookRequest builds an HTTP request for updating a webhook
func (c *V1Client) BuildUpdateWebhookRequest(ctx context.Context, updated *models.Webhook) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(updated.ID, 10))

	return c.buildDataRequest(http.MethodPut, uri, updated)
}

// UpdateWebhook updates a webhook
func (c *V1Client) UpdateWebhook(ctx context.Context, updated *models.Webhook) error {
	req, err := c.BuildUpdateWebhookRequest(ctx, updated)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &updated)
}

// BuildArchiveWebhookRequest builds an HTTP request for updating a webhook
func (c *V1Client) BuildArchiveWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodDelete, uri, nil)
}

// ArchiveWebhook archives a webhook
func (c *V1Client) ArchiveWebhook(ctx context.Context, id uint64) error {
	req, err := c.BuildArchiveWebhookRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

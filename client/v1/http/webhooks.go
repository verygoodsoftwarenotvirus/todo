package client

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	webhooksBasePath = "webhooks"
)

// BuildGetWebhookRequest builds an http Request for fetching an item
func (c *V1Client) BuildGetWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetWebhook gets an webhook
func (c *V1Client) GetWebhook(ctx context.Context, id uint64) (webhook *models.Webhook, err error) {
	req, err := c.BuildGetWebhookRequest(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &webhook)
	return webhook, err
}

// BuildGetWebhooksRequest builds an http Request for fetching items
func (c *V1Client) BuildGetWebhooksRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	uri := c.BuildURL(filter.ToValues(), webhooksBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetWebhooks gets a list of webhooks
func (c *V1Client) GetWebhooks(ctx context.Context, filter *models.QueryFilter) (webhooks *models.WebhookList, err error) {
	req, err := c.BuildGetWebhooksRequest(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &webhooks)
	return webhooks, err
}

// BuildCreateWebhookRequest builds an http Request for creating an item
func (c *V1Client) BuildCreateWebhookRequest(ctx context.Context, body *models.WebhookInput) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath)

	return c.buildDataRequest(http.MethodPost, uri, body)
}

// CreateWebhook creates an webhook
func (c *V1Client) CreateWebhook(ctx context.Context, input *models.WebhookInput) (webhook *models.Webhook, err error) {
	req, err := c.BuildCreateWebhookRequest(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.makeRequest(ctx, req, &webhook)
	return webhook, err
}

// BuildUpdateWebhookRequest builds an http Request for updating an item
func (c *V1Client) BuildUpdateWebhookRequest(ctx context.Context, updated *models.Webhook) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(updated.ID, 10))

	return c.buildDataRequest(http.MethodPut, uri, updated)
}

// UpdateWebhook updates an webhook
func (c *V1Client) UpdateWebhook(ctx context.Context, updated *models.Webhook) error {
	req, err := c.BuildUpdateWebhookRequest(ctx, updated)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	return c.makeRequest(ctx, req, &updated)
}

// BuildDeleteWebhookRequest builds an http Request for updating an item
func (c *V1Client) BuildDeleteWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodDelete, uri, nil)
}

// DeleteWebhook deletes an webhook
func (c *V1Client) DeleteWebhook(ctx context.Context, id uint64) error {
	req, err := c.BuildDeleteWebhookRequest(ctx, id)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	return c.makeRequest(ctx, req, nil)
}

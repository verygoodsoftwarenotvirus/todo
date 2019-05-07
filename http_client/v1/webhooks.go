package client

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	webhooksBasePath = "webhooks"
)

// GetWebhook gets an webhook
func (c *V1Client) GetWebhook(ctx context.Context, id uint64) (webhook *models.Webhook, err error) {
	logger := c.logger.WithValue("id", id)
	logger.Debug("GetWebhook called")

	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))
	err = c.get(ctx, uri, &webhook)
	return webhook, err
}

// GetWebhooks gets a list of webhooks
func (c *V1Client) GetWebhooks(ctx context.Context, filter *models.QueryFilter) (webhooks *models.WebhookList, err error) {
	logger := c.logger.WithValue("filter", filter)
	logger.Debug("GetWebhooks called")

	uri := c.BuildURL(filter.ToValues(), webhooksBasePath)
	err = c.get(ctx, uri, &webhooks)
	return webhooks, err
}

// CreateWebhook creates an webhook
func (c *V1Client) CreateWebhook(ctx context.Context, input *models.WebhookInput) (webhook *models.Webhook, err error) {
	logger := c.logger.WithValues(map[string]interface{}{
		"input_name": input.Name,
	})
	logger.Debug("CreateWebhook called")

	uri := c.BuildURL(nil, webhooksBasePath)
	err = c.post(ctx, uri, input, &webhook)
	return webhook, err
}

// UpdateWebhook updates an webhook
func (c *V1Client) UpdateWebhook(ctx context.Context, updated *models.Webhook) error {
	logger := c.logger.WithValue("id", updated.ID)
	logger.Debug("UpdateWebhook called")

	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(updated.ID, 10))
	err := c.put(ctx, uri, updated, &updated)
	return err
}

// DeleteWebhook deletes an webhook
func (c *V1Client) DeleteWebhook(ctx context.Context, id uint64) error {
	logger := c.logger.WithValue("id", id)
	logger.Debug("DeleteWebhook called")

	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))
	err := c.delete(ctx, uri)
	return err
}

package dbclient

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.WebhookDataManager = (*Client)(nil)

// GetWebhook fetches an webhook from the postgres database
func (c *Client) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	ctx, span := trace.StartSpan(ctx, "GetWebhook")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("webhook_id", strconv.FormatUint(webhookID, 10)))
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	}).Debug("GetWebhook called")

	return c.database.GetWebhook(ctx, webhookID, userID)
}

// GetWebhookCount fetches the count of webhooks from the postgres database that meet a particular filter
func (c *Client) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetWebhookCount")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	}).Debug("GetWebhookCount called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	return c.database.GetWebhookCount(ctx, filter, userID)
}

// GetAllWebhooksCount fetches the count of webhooks from the postgres database that meet a particular filter
func (c *Client) GetAllWebhooksCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetAllWebhooksCount")
	defer span.End()

	c.logger.WithValues(map[string]interface{}{
		"filter": filter,
	}).Debug("GetWebhookCount called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	return c.database.GetAllWebhooksCount(ctx, filter)
}

// GetWebhooks fetches a list of webhooks from the postgres database that meet a particular filter
func (c *Client) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	ctx, span := trace.StartSpan(ctx, "GetWebhooks")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	}).Debug("GetWebhookCount called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	return c.database.GetWebhooks(ctx, filter, userID)
}

// CreateWebhook creates an webhook in a postgres database
func (c *Client) CreateWebhook(ctx context.Context, input *models.WebhookInput) (*models.Webhook, error) {
	ctx, span := trace.StartSpan(ctx, "CreateWebhook")
	defer span.End()

	c.logger.WithValue("input", input).Debug("CreateWebhook called")

	return c.database.CreateWebhook(ctx, input)
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (c *Client) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	ctx, span := trace.StartSpan(ctx, "UpdateWebhook")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("webhook_id", strconv.FormatUint(input.ID, 10)))

	c.logger.WithValue("input", input).Debug("UpdateWebhook called")

	return c.database.UpdateWebhook(ctx, input)
}

// DeleteWebhook deletes an webhook from the database by its ID
func (c *Client) DeleteWebhook(ctx context.Context, webhookID uint64, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "DeleteWebhook")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("webhook_id", strconv.FormatUint(webhookID, 10)))
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	}).Debug("DeleteWebhook called")

	return c.database.DeleteWebhook(ctx, webhookID, userID)
}

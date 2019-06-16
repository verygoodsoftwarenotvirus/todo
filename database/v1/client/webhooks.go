package dbclient

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.WebhookDataManager = (*Client)(nil)

func attachWebhookIDToSpan(span *trace.Span, webhookID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("webhook_id", strconv.FormatUint(webhookID, 10)))
	}
}

// GetWebhook fetches an webhook from the postgres querier
func (c *Client) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	ctx, span := trace.StartSpan(ctx, "GetWebhook")
	defer span.End()

	attachUserIDToSpan(span, userID)
	attachWebhookIDToSpan(span, webhookID)

	c.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	}).Debug("GetWebhook called")

	return c.querier.GetWebhook(ctx, webhookID, userID)
}

// GetWebhookCount fetches the count of webhooks from the postgres querier that meet a particular filter
func (c *Client) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetWebhookCount")
	defer span.End()

	if filter == nil {
		filter = models.DefaultQueryFilter()
	}

	attachFilterToSpan(span, filter)
	attachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	}).Debug("GetWebhookCount called")

	return c.querier.GetWebhookCount(ctx, filter, userID)
}

// GetAllWebhooksCount fetches the count of webhooks from the postgres querier that meet a particular filter
func (c *Client) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetAllWebhooksCount")
	defer span.End()

	c.logger.Debug("GetAllWebhooksCount called")

	return c.querier.GetAllWebhooksCount(ctx)
}

// GetAllWebhooks fetches a list of webhooks from the postgres querier that meet a particular filter
func (c *Client) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllWebhooks")
	defer span.End()

	c.logger.Debug("GetWebhookCount called")

	return c.querier.GetAllWebhooks(ctx)
}

// GetAllWebhooksForUser fetches a list of webhooks from the postgres querier that meet a particular filter
func (c *Client) GetAllWebhooksForUser(ctx context.Context, userID uint64) ([]models.Webhook, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllWebhooksForUser")
	defer span.End()

	attachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("GetAllWebhooksForUser called")

	return c.querier.GetAllWebhooksForUser(ctx, userID)
}

// GetWebhooks fetches a list of webhooks from the postgres querier that meet a particular filter
func (c *Client) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	ctx, span := trace.StartSpan(ctx, "GetWebhooks")
	defer span.End()

	if filter == nil {
		filter = models.DefaultQueryFilter()
	}

	attachUserIDToSpan(span, userID)
	attachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetWebhookCount called")

	return c.querier.GetWebhooks(ctx, filter, userID)
}

// CreateWebhook creates an webhook in a postgres querier
func (c *Client) CreateWebhook(ctx context.Context, input *models.WebhookInput) (*models.Webhook, error) {
	ctx, span := trace.StartSpan(ctx, "CreateWebhook")
	defer span.End()

	attachUserIDToSpan(span, input.BelongsTo)
	c.logger.WithValue("user_id", input.BelongsTo).Debug("CreateWebhook called")

	return c.querier.CreateWebhook(ctx, input)
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (c *Client) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	ctx, span := trace.StartSpan(ctx, "UpdateWebhook")
	defer span.End()

	attachWebhookIDToSpan(span, input.ID)
	attachUserIDToSpan(span, input.BelongsTo)

	c.logger.WithValue("webhook_id", input.ID).Debug("UpdateWebhook called")

	return c.querier.UpdateWebhook(ctx, input)
}

// DeleteWebhook deletes an webhook from the querier by its ID
func (c *Client) DeleteWebhook(ctx context.Context, webhookID, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "DeleteWebhook")
	defer span.End()

	attachUserIDToSpan(span, userID)
	attachWebhookIDToSpan(span, webhookID)

	c.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	}).Debug("DeleteWebhook called")

	return c.querier.DeleteWebhook(ctx, webhookID, userID)
}

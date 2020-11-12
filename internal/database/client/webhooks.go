package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var _ models.WebhookDataManager = (*Client)(nil)

// GetWebhook fetches a webhook from the database.
func (c *Client) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	ctx, span := tracing.StartSpan(ctx, "GetWebhook")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachWebhookIDToSpan(span, webhookID)

	c.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	}).Debug("GetWebhook called")

	return c.querier.GetWebhook(ctx, webhookID, userID)
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (c *Client) GetWebhooks(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.WebhookList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetWebhooks")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetWebhookCount called")

	return c.querier.GetWebhooks(ctx, userID, filter)
}

// GetAllWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (c *Client) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetAllWebhooks")
	defer span.End()

	c.logger.Debug("GetWebhookCount called")

	return c.querier.GetAllWebhooks(ctx)
}

// GetAllWebhooksCount fetches the count of webhooks from the database that meet a particular filter.
func (c *Client) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAllWebhooksCount")
	defer span.End()

	c.logger.Debug("GetAllWebhooksCount called")

	return c.querier.GetAllWebhooksCount(ctx)
}

// CreateWebhook creates a webhook in a database.
func (c *Client) CreateWebhook(ctx context.Context, input *models.WebhookCreationInput) (*models.Webhook, error) {
	ctx, span := tracing.StartSpan(ctx, "CreateWebhook")
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.BelongsToUser)
	c.logger.WithValue("user_id", input.BelongsToUser).Debug("CreateWebhook called")

	return c.querier.CreateWebhook(ctx, input)
}

// UpdateWebhook updates a particular webhook.
// NOTE: this function expects the provided input to have a non-zero ID.
func (c *Client) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	ctx, span := tracing.StartSpan(ctx, "UpdateWebhook")
	defer span.End()

	tracing.AttachWebhookIDToSpan(span, input.ID)
	tracing.AttachUserIDToSpan(span, input.BelongsToUser)

	c.logger.WithValue("webhook_id", input.ID).Debug("UpdateWebhook called")

	return c.querier.UpdateWebhook(ctx, input)
}

// ArchiveWebhook archives a webhook from the database.
func (c *Client) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx, "ArchiveWebhook")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachWebhookIDToSpan(span, webhookID)

	c.logger.WithValues(map[string]interface{}{
		"webhook_id": webhookID,
		"user_id":    userID,
	}).Debug("ArchiveWebhook called")

	return c.querier.ArchiveWebhook(ctx, webhookID, userID)
}

// GetAuditLogEntriesForWebhook fetches a list of audit log entries from the database that relate to a given webhook.
func (c *Client) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]models.AuditLogEntry, error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntriesForWebhook")
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForWebhook called")

	return c.querier.GetAuditLogEntriesForWebhook(ctx, webhookID)
}

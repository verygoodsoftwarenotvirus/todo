package httpclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// GetWebhook retrieves a webhook.
func (c *Client) GetWebhook(ctx context.Context, webhookID uint64) (*types.Webhook, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.WebhookIDKey, webhookID)

	req, err := c.requestBuilder.BuildGetWebhookRequest(ctx, webhookID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building get webhook request")
	}

	var webhook *types.Webhook
	if err = c.fetchAndUnmarshal(ctx, req, &webhook); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving webhook")
	}

	return webhook, nil
}

// GetWebhooks gets a list of webhooks.
func (c *Client) GetWebhooks(ctx context.Context, filter *types.QueryFilter) (*types.WebhookList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.loggerWithFilter(filter)

	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetWebhooksRequest(ctx, filter)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building webhooks list request")
	}

	var webhooks *types.WebhookList
	if err = c.fetchAndUnmarshal(ctx, req, &webhooks); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving webhooks")
	}

	return webhooks, nil
}

// CreateWebhook creates a webhook.
func (c *Client) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.NameKey, input.Name)
	logger.Debug("creating webhook")

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildCreateWebhookRequest(ctx, input)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building create webhook request")
	}

	var webhook *types.Webhook
	if err = c.fetchAndUnmarshal(ctx, req, &webhook); err != nil {
		return nil, observability.PrepareError(err, logger, span, "creating webhook")
	}

	logger.Debug("webhook created")

	return webhook, nil
}

// UpdateWebhook updates a webhook.
func (c *Client) UpdateWebhook(ctx context.Context, updated *types.Webhook) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.WebhookIDKey, updated.ID)

	req, err := c.requestBuilder.BuildUpdateWebhookRequest(ctx, updated)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building update webhook request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, &updated); err != nil {
		return observability.PrepareError(err, logger, span, "updating webhook")
	}

	return nil
}

// ArchiveWebhook archives a webhook.
func (c *Client) ArchiveWebhook(ctx context.Context, webhookID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.WebhookIDKey, webhookID)

	req, err := c.requestBuilder.BuildArchiveWebhookRequest(ctx, webhookID)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building archive webhook request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "archiving webhook")
	}

	return nil
}

// GetAuditLogForWebhook retrieves a list of audit log entries pertaining to a webhook.
func (c *Client) GetAuditLogForWebhook(ctx context.Context, webhookID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.WebhookIDKey, webhookID)

	req, err := c.requestBuilder.BuildGetAuditLogForWebhookRequest(ctx, webhookID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building get audit log entries for webhook request")
	}

	var entries []*types.AuditLogEntry
	if err = c.fetchAndUnmarshal(ctx, req, &entries); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving audit log entries for webhook")
	}

	return entries, nil
}

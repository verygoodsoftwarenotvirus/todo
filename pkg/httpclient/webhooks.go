package httpclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetWebhook retrieves a webhook.
func (c *Client) GetWebhook(ctx context.Context, id uint64) (webhook *types.Webhook, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.BuildGetWebhookRequest(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &webhook)

	return webhook, err
}

// GetWebhooks gets a list of webhooks.
func (c *Client) GetWebhooks(ctx context.Context, filter *types.QueryFilter) (webhooks *types.WebhookList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetWebhooksRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &webhooks)

	return webhooks, err
}

// CreateWebhook creates a webhook.
func (c *Client) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (webhook *types.Webhook, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildCreateWebhookRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &webhook)

	return webhook, err
}

// UpdateWebhook updates a webhook.
func (c *Client) UpdateWebhook(ctx context.Context, updated *types.Webhook) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	req, err := c.BuildUpdateWebhookRequest(ctx, updated)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &updated)
}

// ArchiveWebhook archives a webhook.
func (c *Client) ArchiveWebhook(ctx context.Context, id uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.BuildArchiveWebhookRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// GetAuditLogForWebhook retrieves a list of audit log entries pertaining to a webhook.
func (c *Client) GetAuditLogForWebhook(ctx context.Context, webhookID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.BuildGetAuditLogForWebhookRequest(ctx, webhookID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

package httpclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetWebhook retrieves a webhook.
func (c *Client) GetWebhook(ctx context.Context, id uint64) (webhook *types.Webhook, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetWebhookRequest(ctx, id)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &webhook); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, fmt.Errorf("fetching webhook: %w", retrieveErr)
	}

	return webhook, nil
}

// GetWebhooks gets a list of webhooks.
func (c *Client) GetWebhooks(ctx context.Context, filter *types.QueryFilter) (webhooks *types.WebhookList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.requestBuilder.BuildGetWebhooksRequest(ctx, filter)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &webhooks); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, fmt.Errorf("fetching webhooks: %w", retrieveErr)
	}

	return webhooks, nil
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
		tracing.AttachErrorToSpan(span, validationErr)
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.requestBuilder.BuildCreateWebhookRequest(ctx, input)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if createErr := c.executeRequest(ctx, req, &webhook); createErr != nil {
		tracing.AttachErrorToSpan(span, createErr)
		return nil, fmt.Errorf("creating webhook: %w", createErr)
	}

	return webhook, nil
}

// UpdateWebhook updates a webhook.
func (c *Client) UpdateWebhook(ctx context.Context, updated *types.Webhook) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	req, err := c.requestBuilder.BuildUpdateWebhookRequest(ctx, updated)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building request: %w", err)
	}

	if updateErr := c.executeRequest(ctx, req, &updated); updateErr != nil {
		tracing.AttachErrorToSpan(span, updateErr)
		return fmt.Errorf("updating webhook: %w", updateErr)
	}

	return nil
}

// ArchiveWebhook archives a webhook.
func (c *Client) ArchiveWebhook(ctx context.Context, id uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildArchiveWebhookRequest(ctx, id)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building request: %w", err)
	}

	if archiveErr := c.executeRequest(ctx, req, nil); archiveErr != nil {
		tracing.AttachErrorToSpan(span, archiveErr)
		return fmt.Errorf("archiving webhook: %w", archiveErr)
	}

	return nil
}

// GetAuditLogForWebhook retrieves a list of audit log entries pertaining to a webhook.
func (c *Client) GetAuditLogForWebhook(ctx context.Context, webhookID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetAuditLogForWebhookRequest(ctx, webhookID)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, retrieveErr
	}

	return entries, nil
}

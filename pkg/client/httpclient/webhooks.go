package httpclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// GetWebhook retrieves a webhook.
func (c *Client) GetWebhook(ctx context.Context, webhookID string) (*types.Webhook, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" {
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
func (c *Client) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (string, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return "", ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.NameKey, input.Name)
	logger.Debug("creating webhook")

	if err := input.ValidateWithContext(ctx); err != nil {
		return "", observability.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildCreateWebhookRequest(ctx, input)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "building create webhook request")
	}

	var pwr *types.PreWriteResponse
	if err = c.fetchAndUnmarshal(ctx, req, &pwr); err != nil {
		return "", observability.PrepareError(err, logger, span, "creating webhook")
	}

	logger.Debug("webhook created")

	return pwr.ID, nil
}

// ArchiveWebhook archives a webhook.
func (c *Client) ArchiveWebhook(ctx context.Context, webhookID string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" {
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

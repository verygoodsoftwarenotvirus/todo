package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	webhooksBasePath = "webhooks"
)

// BuildGetWebhookRequest builds an HTTP request for fetching a webhook.
func (c *Client) BuildGetWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetWebhooksRequest builds an HTTP request for fetching webhooks.
func (c *Client) BuildGetWebhooksRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), webhooksBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateWebhookRequest builds an HTTP request for creating a webhook.
func (c *Client) BuildCreateWebhookRequest(ctx context.Context, input *types.WebhookCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, webhooksBasePath)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildUpdateWebhookRequest builds an HTTP request for updating a webhook.
func (c *Client) BuildUpdateWebhookRequest(ctx context.Context, updated *types.Webhook) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return nil, ErrNilInputProvided
	}

	uri := c.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(updated.ID, 10))

	return c.buildDataRequest(ctx, http.MethodPut, uri, updated)
}

// BuildArchiveWebhookRequest builds an HTTP request for updating a webhook.
func (c *Client) BuildArchiveWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForWebhookRequest builds an HTTP request for fetching a list of audit log entries pertaining to a webhook.
func (c *Client) BuildGetAuditLogForWebhookRequest(ctx context.Context, webhookID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(webhookID, 10), "audit")

	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

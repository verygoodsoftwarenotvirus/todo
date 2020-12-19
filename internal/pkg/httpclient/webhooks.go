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
func (c *V1Client) BuildGetWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetWebhook retrieves a webhook.
func (c *V1Client) GetWebhook(ctx context.Context, id uint64) (webhook *types.Webhook, err error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetWebhookRequest(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &webhook)

	return webhook, err
}

// BuildGetWebhooksRequest builds an HTTP request for fetching webhooks.
func (c *V1Client) BuildGetWebhooksRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(filter.ToValues(), webhooksBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetWebhooks gets a list of webhooks.
func (c *V1Client) GetWebhooks(ctx context.Context, filter *types.QueryFilter) (webhooks *types.WebhookList, err error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetWebhooksRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &webhooks)

	return webhooks, err
}

// BuildCreateWebhookRequest builds an HTTP request for creating a webhook.
func (c *V1Client) BuildCreateWebhookRequest(ctx context.Context, body *types.WebhookCreationInput) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(nil, webhooksBasePath)

	return c.buildDataRequest(ctx, http.MethodPost, uri, body)
}

// CreateWebhook creates a webhook.
func (c *V1Client) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (webhook *types.Webhook, err error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildCreateWebhookRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &webhook)

	return webhook, err
}

// BuildUpdateWebhookRequest builds an HTTP request for updating a webhook.
func (c *V1Client) BuildUpdateWebhookRequest(ctx context.Context, updated *types.Webhook) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(updated.ID, 10))

	return c.buildDataRequest(ctx, http.MethodPut, uri, updated)
}

// UpdateWebhook updates a webhook.
func (c *V1Client) UpdateWebhook(ctx context.Context, updated *types.Webhook) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildUpdateWebhookRequest(ctx, updated)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &updated)
}

// BuildArchiveWebhookRequest builds an HTTP request for updating a webhook.
func (c *V1Client) BuildArchiveWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveWebhook archives a webhook.
func (c *V1Client) ArchiveWebhook(ctx context.Context, id uint64) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildArchiveWebhookRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildGetAuditLogForWebhookRequest builds an HTTP request for fetching a list of audit log entries pertaining to a webhook.
func (c *V1Client) BuildGetAuditLogForWebhookRequest(ctx context.Context, webhookID uint64) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		webhooksBasePath,
		strconv.FormatUint(webhookID, 10),
		"audit",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogForWebhook retrieves a list of audit log entries pertaining to a webhook.
func (c *V1Client) GetAuditLogForWebhook(ctx context.Context, webhookID uint64) (entries []types.AuditLogEntry, err error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAuditLogForWebhookRequest(ctx, webhookID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

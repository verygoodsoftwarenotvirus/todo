package requests

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
func (b *Builder) BuildGetWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetWebhooksRequest builds an HTTP request for fetching webhooks.
func (b *Builder) BuildGetWebhooksRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.BuildURL(ctx, filter.ToValues(), webhooksBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateWebhookRequest builds an HTTP request for creating a webhook.
func (b *Builder) BuildCreateWebhookRequest(ctx context.Context, input *types.WebhookCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		b.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := b.BuildURL(ctx, nil, webhooksBasePath)

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildUpdateWebhookRequest builds an HTTP request for updating a webhook.
func (b *Builder) BuildUpdateWebhookRequest(ctx context.Context, updated *types.Webhook) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return nil, ErrNilInputProvided
	}

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(updated.ID, 10))

	return b.buildDataRequest(ctx, http.MethodPut, uri, updated)
}

// BuildArchiveWebhookRequest builds an HTTP request for updating a webhook.
func (b *Builder) BuildArchiveWebhookRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(id, 10))

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForWebhookRequest builds an HTTP request for fetching a list of audit log entries pertaining to a webhook.
func (b *Builder) BuildGetAuditLogForWebhookRequest(ctx context.Context, webhookID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(webhookID, 10), "audit")

	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

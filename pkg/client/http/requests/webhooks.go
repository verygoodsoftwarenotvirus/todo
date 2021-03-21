package requests

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/errs"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	webhooksBasePath = "webhooks"
)

// BuildGetWebhookRequest builds an HTTP request for fetching a webhook.
func (b *Builder) BuildGetWebhookRequest(ctx context.Context, webhookID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachWebhookIDToSpan(span, webhookID)

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(webhookID, 10))

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetWebhooksRequest builds an HTTP request for fetching a list of webhooks.
func (b *Builder) BuildGetWebhooksRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachQueryFilterToSpan(span, filter)
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

	logger := b.logger.WithValue(keys.NameKey, input.Name)

	if err := input.Validate(ctx); err != nil {
		return nil, errs.PrepareError(err, logger, span, "validating input")
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

	tracing.AttachWebhookIDToSpan(span, updated.ID)

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(updated.ID, 10))

	return b.buildDataRequest(ctx, http.MethodPut, uri, updated)
}

// BuildArchiveWebhookRequest builds an HTTP request for archiving a webhook.
func (b *Builder) BuildArchiveWebhookRequest(ctx context.Context, webhookID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachWebhookIDToSpan(span, webhookID)

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(webhookID, 10))

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForWebhookRequest builds an HTTP request for fetching a list of audit log entries pertaining to a webhook.
func (b *Builder) BuildGetAuditLogForWebhookRequest(ctx context.Context, webhookID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachWebhookIDToSpan(span, webhookID)

	uri := b.BuildURL(ctx, nil, webhooksBasePath, strconv.FormatUint(webhookID, 10), "audit")

	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

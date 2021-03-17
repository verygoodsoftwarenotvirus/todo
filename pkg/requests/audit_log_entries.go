package requests

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	auditLogBasePath = "audit_log"
)

// BuildGetAuditLogEntriesRequest builds an HTTP request for fetching entries.
func (b *Builder) BuildGetAuditLogEntriesRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.BuildURL(
		ctx,
		filter.ToValues(),
		adminBasePath,
		auditLogBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetAuditLogEntryRequest builds an HTTP request for fetching entries.
func (b *Builder) BuildGetAuditLogEntryRequest(ctx context.Context, entryID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.BuildURL(
		ctx,
		nil,
		adminBasePath,
		auditLogBasePath,
		strconv.FormatUint(entryID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

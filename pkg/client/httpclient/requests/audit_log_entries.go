package requests

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	auditLogBasePath = "audit_log"
)

// BuildGetAuditLogEntryRequest builds an HTTP request for fetching a given audit log entry.
func (b *Builder) BuildGetAuditLogEntryRequest(ctx context.Context, entryID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger.WithValue(keys.AuditLogEntryIDKey, entryID)
	tracing.AttachAuditLogEntryIDToSpan(span, entryID)

	uri := b.BuildURL(
		ctx,
		nil,
		adminBasePath,
		auditLogBasePath,
		id(entryID),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildGetAuditLogEntriesRequest builds an HTTP request for fetching audit log entries.
func (b *Builder) BuildGetAuditLogEntriesRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(b.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	uri := b.BuildURL(
		ctx,
		filter.ToValues(),
		adminBasePath,
		auditLogBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

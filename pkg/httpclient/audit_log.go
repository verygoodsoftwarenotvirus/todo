package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	auditLogBasePath = "audit_log"
)

// BuildGetAuditLogEntriesRequest builds an HTTP request for fetching entries.
func (c *Client) BuildGetAuditLogEntriesRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		ctx,
		filter.ToValues(),
		adminBasePath,
		auditLogBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogEntries retrieves a list of entries.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (entries *types.AuditLogEntryList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAuditLogEntriesRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	c.logger.WithRequest(req).Debug("Fetching audit log entries")

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

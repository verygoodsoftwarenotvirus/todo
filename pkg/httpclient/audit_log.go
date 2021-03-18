package httpclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetAuditLogEntries retrieves a list of entries.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (entries *types.AuditLogEntryList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.requestBuilder.BuildGetAuditLogEntriesRequest(ctx, filter)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	c.logger.WithRequest(req).Debug("Fetching audit log entries")

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, retrieveErr
	}

	return entries, nil
}

// GetAuditLogEntry retrieves an entry.
func (c *Client) GetAuditLogEntry(ctx context.Context, entryID uint64) (entry *types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.requestBuilder.BuildGetAuditLogEntryRequest(ctx, entryID)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	c.logger.WithRequest(req).Debug("Fetching audit log entry")

	if retrieveErr := c.retrieve(ctx, req, &entry); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, retrieveErr
	}

	return entry, nil
}

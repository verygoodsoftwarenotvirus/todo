package httpclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// GetAuditLogEntry retrieves an entry.
func (c *Client) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if entryID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AuditLogEntryIDKey, entryID)

	req, err := c.requestBuilder.BuildGetAuditLogEntryRequest(ctx, entryID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building get audit log entry request")
	}

	c.logger.WithRequest(req).Debug("Fetching audit log entry")

	var entry *types.AuditLogEntry
	if err = c.fetchAndUnmarshal(ctx, req, &entry); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving audit log entry")
	}

	return entry, nil
}

// GetAuditLogEntries retrieves a list of audit log entries.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.loggerWithFilter(filter)

	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetAuditLogEntriesRequest(ctx, filter)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building fetch audit log entries request")
	}

	logger = logger.WithRequest(req)

	var entries *types.AuditLogEntryList
	if err = c.fetchAndUnmarshal(ctx, req, &entries); err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching audit log entries")
	}

	return entries, nil
}

package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var _ models.AuditLogEntryDataManager = (*Client)(nil)

// GetAuditLogEntry fetches an audit log entry from the database.
func (c *Client) GetAuditLogEntry(ctx context.Context, entryID uint64) (*models.AuditLogEntry, error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntry")
	defer span.End()

	tracing.AttachAuditLogEntryIDToSpan(span, entryID)
	c.logger.WithValue("audit_log_entry_id", entryID).Debug("GetAuditLogEntry called")

	return c.querier.GetAuditLogEntry(ctx, entryID)
}

// GetAllAuditLogEntriesCount fetches the count of audit log entries from the database that meet a particular filter.
func (c *Client) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAllAuditLogEntriesCount")
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntriesCount called")

	return c.querier.GetAllAuditLogEntriesCount(ctx)
}

// GetAllAuditLogEntries fetches a list of all audit log entries in the database.
func (c *Client) GetAllAuditLogEntries(ctx context.Context, results chan []models.AuditLogEntry) error {
	ctx, span := tracing.StartSpan(ctx, "GetAllAuditLogEntries")
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntries called")

	return c.querier.GetAllAuditLogEntries(ctx, results)
}

// GetAuditLogEntries fetches a list of audit log entries from the database that meet a particular filter.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *models.QueryFilter) (*models.AuditLogEntryList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntries")
	defer span.End()

	tracing.AttachFilterToSpan(span, filter)
	c.logger.Debug("GetAuditLogEntries called")

	return c.querier.GetAuditLogEntries(ctx, filter)
}

// CreateAuditLogEntry creates an audit log entry in the database.
func (c *Client) CreateAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) {
	ctx, span := tracing.StartSpan(ctx, "CreateAuditLogEntry")
	defer span.End()

	tracing.AttachAuditLogEntryEventTypeToSpan(span, string(input.EventType))
	c.logger.WithValue("event_type", input.EventType).Debug("CreateAuditLogEntry called")

	c.querier.CreateAuditLogEntry(ctx, input)
}

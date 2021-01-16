package superclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AuditLogEntryDataManager = (*Client)(nil)

// GetAuditLogEntry fetches an audit log entry from the database.
func (c *Client) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAuditLogEntryIDToSpan(span, entryID)
	c.logger.WithValue(keys.AuditLogEntryIDKey, entryID).Debug("GetAuditLogEntry called")

	return c.querier.GetAuditLogEntry(ctx, entryID)
}

// GetAllAuditLogEntriesCount fetches the count of audit log entries from the database that meet a particular filter.
func (c *Client) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntriesCount called")

	return c.querier.GetAllAuditLogEntriesCount(ctx)
}

// GetAllAuditLogEntries fetches a list of all audit log entries in the database.
func (c *Client) GetAllAuditLogEntries(ctx context.Context, results chan []*types.AuditLogEntry, bucketSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntries called")

	return c.querier.GetAllAuditLogEntries(ctx, results, bucketSize)
}

// GetAuditLogEntries fetches a list of audit log entries from the database that meet a particular filter.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
	}

	c.logger.Debug("GetAuditLogEntries called")

	return c.querier.GetAuditLogEntries(ctx, filter)
}

// LogOAuth2ClientCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, client.BelongsToUser).Debug("LogOAuth2ClientCreationEvent called")

	c.querier.LogOAuth2ClientCreationEvent(ctx, client)
}

// LogOAuth2ClientArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogOAuth2ClientArchiveEvent called")

	c.querier.LogOAuth2ClientArchiveEvent(ctx, userID, clientID)
}

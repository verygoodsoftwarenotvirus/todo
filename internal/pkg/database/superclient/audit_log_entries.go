package superclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AuditLogEntryDataManager = (*Client)(nil)

// scanAuditLogEntry takes a database Scanner (i.e. *sql.Row) and scans the result into an AuditLogEntry struct.
func (c *Client) scanAuditLogEntry(scan database.Scanner, includeCounts bool) (entry *types.AuditLogEntry, totalCount uint64, err error) {
	entry = &types.AuditLogEntry{}

	targetVars := []interface{}{
		&entry.ID,
		&entry.EventType,
		&entry.Context,
		&entry.CreatedOn,
	}

	if includeCounts {
		targetVars = append(targetVars, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, scanErr
	}

	return entry, totalCount, nil
}

// scanAuditLogEntries takes some database rows and turns them into a slice of .
func (c *Client) scanAuditLogEntries(rows database.ResultIterator, includeCounts bool) (entries []*types.AuditLogEntry, totalCount uint64, err error) {
	for rows.Next() {
		x, tc, scanErr := c.scanAuditLogEntry(rows, includeCounts)
		if scanErr != nil {
			return nil, 0, scanErr
		}

		if includeCounts {
			if totalCount == 0 {
				totalCount = tc
			}
		}

		entries = append(entries, x)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing database rows")
		return nil, 0, closeErr
	}

	return entries, totalCount, nil
}

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

// createAuditLogEntry creates an audit log entry in the database.
func (c *Client) createAuditLogEntry(ctx context.Context, input *types.AuditLogEntryCreationInput) {
	x := &types.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	c.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType).Debug("createAuditLogEntry called")
	query, args := c.sqlQueryBuilder.BuildCreateAuditLogEntryQuery(x)

	// create the audit log entry.
	if _, err := c.db.ExecContext(ctx, query, args...); err != nil {
		c.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType).Error(err, "executing audit log entry creation query")
	}
}

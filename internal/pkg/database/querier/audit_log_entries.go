package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AuditLogEntryDataManager = (*Client)(nil)
)

// scanAuditLogEntry takes a database Scanner (i.e. *sql.Row) and scans the result into an AuditLogEntry struct.
func (c *Client) scanAuditLogEntry(scan database.Scanner, includeCounts bool) (entry *types.AuditLogEntry, totalCount uint64, err error) {
	entry = &types.AuditLogEntry{}

	targetVars := []interface{}{
		&entry.ID,
		&entry.ExternalID,
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

// scanAuditLogEntries takes some database rows and turns them into a slice of AuditLogEntry pointers.
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

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntryQuery(entryID)
	row := c.db.QueryRowContext(ctx, query, args...)

	entry, _, err := c.scanAuditLogEntry(row, false)
	if err != nil {
		return nil, fmt.Errorf("scanning audit log entry: %w", err)
	}

	return entry, nil
}

// GetAllAuditLogEntriesCount fetches the count of audit log entries from the database that meet a particular filter.
func (c *Client) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntriesCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllAuditLogEntriesCountQuery())
}

// GetAllAuditLogEntries fetches a list of all audit log entries in the database.
func (c *Client) GetAllAuditLogEntries(ctx context.Context, results chan []*types.AuditLogEntry, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntries called")

	count, countErr := c.GetAllAuditLogEntriesCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of entries: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfAuditLogEntriesQuery(begin, end)
			logger := c.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			auditLogEntries, _, scanErr := c.scanAuditLogEntries(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			results <- auditLogEntries
		}(beginID, endID)
	}

	return nil
}

// GetAuditLogEntries fetches a list of audit log entries from the database that meet a particular filter.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (x *types.AuditLogEntryList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.AuditLogEntryList{}

	c.logger.Debug("GetAuditLogEntries called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesQuery(filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	if x.Entries, x.TotalCount, err = c.scanAuditLogEntries(rows, true); err != nil {
		return nil, fmt.Errorf("scanning audit log entry: %w", err)
	}

	return x, nil
}

// createAuditLogEntryInTransaction creates an audit log entry in the database.
func (c *Client) createAuditLogEntryInTransaction(ctx context.Context, transaction *sql.Tx, input *types.AuditLogEntryCreationInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType)
	query, args := c.sqlQueryBuilder.BuildCreateAuditLogEntryQuery(input)

	tracing.AttachAuditLogEntryEventTypeToSpan(span, input.EventType)
	logger.Debug("audit log entry created")

	// create the audit log entry.
	if err := c.performWriteQueryIgnoringReturn(ctx, transaction, "audit log entry creation", query, args); err != nil {
		logger.Error(err, "executing audit log entry creation query")
		c.rollbackTransaction(transaction)

		return err
	}

	return nil
}

// createAuditLogEntry creates an audit log entry in the database.
func (c *Client) createAuditLogEntry(ctx context.Context, querier database.Querier, input *types.AuditLogEntryCreationInput) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAuditLogEntryEventTypeToSpan(span, input.EventType)
	logger := c.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType)

	query, args := c.sqlQueryBuilder.BuildCreateAuditLogEntryQuery(input)

	// create the audit log entry.
	if err := c.performWriteQueryIgnoringReturn(ctx, querier, "audit log entry creation", query, args); err != nil {
		logger.Error(err, "executing audit log entry creation query")
	}

	logger.Debug("audit log entry created")
}

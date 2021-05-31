package querier

import (
	"context"
	"database/sql"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	_ types.AuditLogEntryDataManager = (*SQLQuerier)(nil)
)

// scanAuditLogEntry takes a database Scanner (i.e. *sql.Row) and scans the result into an AuditLogEntry struct.
func (q *SQLQuerier) scanAuditLogEntry(ctx context.Context, scan database.Scanner, includeCounts bool) (entry *types.AuditLogEntry, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)
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

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, observability.PrepareError(err, logger, span, "scanning API client database result")
	}

	return entry, totalCount, nil
}

// scanAuditLogEntries takes some database rows and turns them into a slice of AuditLogEntry pointers.
func (q *SQLQuerier) scanAuditLogEntries(ctx context.Context, rows database.ResultIterator, includeCounts bool) (entries []*types.AuditLogEntry, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		x, tc, scanErr := q.scanAuditLogEntry(ctx, rows, includeCounts)
		if scanErr != nil {
			return nil, 0, observability.PrepareError(scanErr, logger, span, "scanning audit log entries")
		}

		if includeCounts {
			if totalCount == 0 {
				totalCount = tc
			}
		}

		entries = append(entries, x)
	}

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return entries, totalCount, nil
}

// GetAuditLogEntry fetches an audit log entry from the database.
func (q *SQLQuerier) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if entryID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAuditLogEntryIDToSpan(span, entryID)
	logger := q.logger.WithValue(keys.AuditLogEntryIDKey, entryID)

	query, args := q.sqlQueryBuilder.BuildGetAuditLogEntryQuery(ctx, entryID)
	row := q.getOneRow(ctx, q.db, "audit log entry", query, args...)

	entry, _, err := q.scanAuditLogEntry(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entry")
	}

	return entry, nil
}

// GetAllAuditLogEntriesCount fetches the count of audit log entries from the database that meet a particular filter.
func (q *SQLQuerier) GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetAllAuditLogEntriesCountQuery(ctx), "fetching count of audit logs entries")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of audit log entries")
	}

	return count, nil
}

// GetAllAuditLogEntries fetches a list of all audit log entries in the database.
func (q *SQLQuerier) GetAllAuditLogEntries(ctx context.Context, results chan []*types.AuditLogEntry, batchSize uint16) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if results == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValue("batch_size", batchSize)

	count, err := q.GetAllAuditLogEntriesCount(ctx)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching count of entries")
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := q.sqlQueryBuilder.BuildGetBatchOfAuditLogEntriesQuery(ctx, begin, end)
			logger = logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := q.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			auditLogEntries, _, scanErr := q.scanAuditLogEntries(ctx, rows, false)
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
func (q *SQLQuerier) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (x *types.AuditLogEntryList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachQueryFilterToSpan(span, filter)
	logger := filter.AttachToLogger(q.logger)

	x = &types.AuditLogEntryList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetAuditLogEntriesQuery(ctx, filter)

	rows, err := q.performReadQuery(ctx, q.db, "audit log entries", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	if x.Entries, x.TotalCount, err = q.scanAuditLogEntries(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entry")
	}

	return x, nil
}

// createAuditLogEntryInTransaction creates an audit log entry in the database.
func (q *SQLQuerier) createAuditLogEntryInTransaction(ctx context.Context, transaction *sql.Tx, input *types.AuditLogEntryCreationInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	if transaction == nil {
		return ErrNilTransactionProvided
	}

	logger := q.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType).
		WithValue(keys.AuditLogEntryContextKey, input.Context)
	query, args := q.sqlQueryBuilder.BuildCreateAuditLogEntryQuery(ctx, input)

	tracing.AttachAuditLogEntryEventTypeToSpan(span, input.EventType)

	// create the audit log entry.
	if err := q.performWriteQueryIgnoringReturn(ctx, transaction, "audit log entry creation", query, args); err != nil {
		logger.Error(err, "executing audit log entry creation query")
		q.rollbackTransaction(ctx, transaction)

		return err
	}

	logger.Info("audit log entry created")

	return nil
}

// createAuditLogEntry creates an audit log entry in the database.
func (q *SQLQuerier) createAuditLogEntry(ctx context.Context, querier database.Querier, input *types.AuditLogEntryCreationInput) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if input == nil {
		observability.NoteEvent(logger, span, "early return due to nil input")
		return
	}

	if querier == nil {
		observability.NoteEvent(logger, span, "early return due to nil querier")
		return
	}

	tracing.AttachAuditLogEntryEventTypeToSpan(span, input.EventType)
	logger = logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType)

	query, args := q.sqlQueryBuilder.BuildCreateAuditLogEntryQuery(ctx, input)

	// create the audit log entry.
	id, err := q.performWriteQuery(ctx, querier, true, "audit log entry creation", query, args)
	if err != nil {
		logger.Error(err, "executing audit log entry creation query")
	}

	tracing.AttachAuditLogEntryIDToSpan(span, id)

	logger.Info("audit log entry created")
}

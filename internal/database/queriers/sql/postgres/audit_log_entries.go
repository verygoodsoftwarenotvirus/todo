package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"

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

const getAuditLogEntryQuery = `
	SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.id = $1
`

// GetAuditLogEntry fetches an audit log entry from the database.
func (q *SQLQuerier) GetAuditLogEntry(ctx context.Context, entryID string) (*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if entryID == "" {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAuditLogEntryIDToSpan(span, entryID)
	logger := q.logger.WithValue(keys.AuditLogEntryIDKey, entryID)

	args := []interface{}{entryID}

	row := q.getOneRow(ctx, q.db, "audit log entry", getAuditLogEntryQuery, args)

	entry, _, err := q.scanAuditLogEntry(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entry")
	}

	return entry, nil
}

const getAllAuditLogEntryCountQuery = `
	SELECT COUNT(audit_log.id) FROM audit_log
`

// GetAllAuditLogEntriesCount fetches the count of audit log entries from the database that meet a particular filter.
func (q *SQLQuerier) GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, getAllAuditLogEntryCountQuery, "fetching count of audit logs entries")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of audit log entries")
	}

	return count, nil
}

// buildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (q *SQLQuerier) buildGetAuditLogEntriesQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	countQueryBuilder := q.sqlBuilder.Select(allCountQuery).
		From(querybuilding.AuditLogEntriesTableName)

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(span, err)

	builder := q.sqlBuilder.Select(append(querybuilding.AuditLogEntriesTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(querybuilding.AuditLogEntriesTableName).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, querybuilding.AuditLogEntriesTableName, builder)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(span, err)

	return query, append(countQueryArgs, selectArgs...)
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

	query, args := q.buildGetAuditLogEntriesQuery(ctx, filter)

	rows, err := q.performReadQuery(ctx, q.db, "audit log entries", query, args)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	if x.Entries, x.TotalCount, err = q.scanAuditLogEntries(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entry")
	}

	return x, nil
}

const createAuditLogEntryQuery = `
	INSERT INTO audit_log (id,event_type,context) VALUES ($1,$2,$3)
`

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

	logger := q.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType)
	logger = logger.WithValue(keys.AuditLogEntryContextKey, input.Context)

	args := []interface{}{
		input.ID,
		input.EventType,
		input.Context,
	}

	tracing.AttachAuditLogEntryEventTypeToSpan(span, input.EventType)

	// create the audit log entry.
	if err := q.performWriteQuery(ctx, transaction, "audit log entry creation", createAuditLogEntryQuery, args); err != nil {
		logger.Error(err, "executing audit log entry creation query")
		q.rollbackTransaction(ctx, transaction)

		return err
	}

	logger.Info("audit log entry created")

	return nil
}

// createAuditLogEntry creates an audit log entry in the database.
func (q *SQLQuerier) createAuditLogEntry(ctx context.Context, querier database.SQLQueryExecutor, input *types.AuditLogEntryCreationInput) {
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

	args := []interface{}{
		input.ID,
		input.EventType,
		input.Context,
	}

	// create the audit log entry.
	if writeErr := q.performWriteQuery(ctx, querier, "audit log entry creation", createAuditLogEntryQuery, args); writeErr != nil {
		logger.Error(writeErr, "executing audit log entry creation query")
	}

	tracing.AttachAuditLogEntryIDToSpan(span, input.ID)

	logger.Info("audit log entry created")
}

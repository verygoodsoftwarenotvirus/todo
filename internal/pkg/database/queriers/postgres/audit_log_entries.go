package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.AuditLogEntryDataManager = (*Postgres)(nil)

// scanAuditLogEntry takes a database Scanner (i.e. *sql.Row) and scans the result into an AuditLogEntry struct.
func (q *Postgres) scanAuditLogEntry(scan database.Scanner, includeCounts bool) (entry *types.AuditLogEntry, totalCount uint64, err error) {
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
func (q *Postgres) scanAuditLogEntries(rows database.ResultIterator, includeCounts bool) (entries []*types.AuditLogEntry, totalCount uint64, err error) {
	for rows.Next() {
		x, tc, scanErr := q.scanAuditLogEntry(rows, includeCounts)
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
		q.logger.Error(closeErr, "closing database rows")
		return nil, 0, closeErr
	}

	return entries, totalCount, nil
}

// BuildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Postgres) BuildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): entryID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntry fetches an audit log entry from the database.
func (q *Postgres) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	query, args := q.BuildGetAuditLogEntryQuery(entryID)
	row := q.db.QueryRowContext(ctx, query, args...)

	entry, _, err := q.scanAuditLogEntry(row, false)

	return entry, err
}

// BuildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Postgres) BuildGetAllAuditLogEntriesCountQuery() string {
	allAuditLogEntriesCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.AuditLogEntriesTableName)).
		From(queriers.AuditLogEntriesTableName).
		ToSql()
	q.logQueryBuildingError(err)

	return allAuditLogEntriesCountQuery
}

// GetAllAuditLogEntriesCount will fetch the count of  from the database.
func (q *Postgres) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.BuildGetAllAuditLogEntriesCountQuery()).Scan(&count)
	return count, err
}

// BuildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (q *Postgres) BuildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAllAuditLogEntries fetches every audit log entry from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (q *Postgres) GetAllAuditLogEntries(ctx context.Context, resultChannel chan []*types.AuditLogEntry, batchSize uint16) error {
	count, countErr := q.GetAllAuditLogEntriesCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of webhooks: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := q.BuildGetBatchOfAuditLogEntriesQuery(begin, end)
			logger := q.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := q.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			auditLogEntries, _, err := q.scanAuditLogEntries(rows, false)
			if err != nil {
				logger.Error(err, "scanning database rows")
				return
			}

			resultChannel <- auditLogEntries
		}(beginID, endID)
	}

	return nil
}

// BuildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Postgres) BuildGetAuditLogEntriesQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.
		Select(allCountQuery).
		From(queriers.AuditLogEntriesTableName)

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(queriers.AuditLogEntriesTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.AuditLogEntriesTableName).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.AuditLogEntriesTableName)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetAuditLogEntries fetches a list of  from the database that meet a particular filter.
func (q *Postgres) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	query, args := q.BuildGetAuditLogEntriesQuery(filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database: %w", err)
	}

	auditLogEntries, totalCount, err := q.scanAuditLogEntries(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AuditLogEntryList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: totalCount,
		},
		Entries: auditLogEntries,
	}

	return list, nil
}

// BuildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (q *Postgres) BuildCreateAuditLogEntryQuery(input *types.AuditLogEntry) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(queriers.AuditLogEntriesTableName).
		Columns(
			queriers.AuditLogEntriesTableEventTypeColumn,
			queriers.AuditLogEntriesTableContextColumn,
		).
		Values(
			input.EventType,
			input.Context,
		).
		Suffix(fmt.Sprintf("RETURNING %s, %s", queriers.IDColumn, queriers.CreatedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// CreateAuditLogEntry creates an audit log entry in the database.
func (q *Postgres) CreateAuditLogEntry(ctx context.Context, input *types.AuditLogEntryCreationInput) {
	x := &types.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	query, args := q.BuildCreateAuditLogEntryQuery(x)
	q.logger.WithValue("event_type", input.EventType).Debug("CreateAuditLogEntry called")

	// create the audit log entry.
	if err := q.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		q.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType).Error(err, "executing audit log entry creation query")
	}
}

package sqlite

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

var (
	_ types.AuditLogEntryDataManager = (*Sqlite)(nil)
)

// scanAuditLogEntry takes a database Scanner (i.e. *sql.Row) and scans the result into an AuditLogEntry struct.
func (c *Sqlite) scanAuditLogEntry(scan database.Scanner, includeCounts bool) (entry *types.AuditLogEntry, totalCount uint64, err error) {
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
func (c *Sqlite) scanAuditLogEntries(rows database.ResultIterator, includeCounts bool) (entries []*types.AuditLogEntry, totalCount uint64, err error) {
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

// buildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (c *Sqlite) buildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): entryID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntry fetches an audit log entry from the database.
func (c *Sqlite) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	query, args := c.buildGetAuditLogEntryQuery(entryID)
	row := c.db.QueryRowContext(ctx, query, args...)

	entry, _, err := c.scanAuditLogEntry(row, false)

	return entry, err
}

// buildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (c *Sqlite) buildGetAllAuditLogEntriesCountQuery() string {
	allAuditLogEntriesCountQuery, _, err := c.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.AuditLogEntriesTableName)).
		From(queriers.AuditLogEntriesTableName).
		ToSql()
	c.logQueryBuildingError(err)

	return allAuditLogEntriesCountQuery
}

// GetAllAuditLogEntriesCount will fetch the count of  from the database.
func (c *Sqlite) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	err = c.db.QueryRowContext(ctx, c.buildGetAllAuditLogEntriesCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (c *Sqlite) buildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := c.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetAllAuditLogEntries fetches every audit log entry from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (c *Sqlite) GetAllAuditLogEntries(ctx context.Context, resultChannel chan []*types.AuditLogEntry, batchSize uint16) error {
	count, countErr := c.GetAllAuditLogEntriesCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of entries: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.buildGetBatchOfAuditLogEntriesQuery(begin, end)
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

			resultChannel <- auditLogEntries
		}(beginID, endID)
	}

	return nil
}

// buildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (c *Sqlite) buildGetAuditLogEntriesQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := c.sqlBuilder.
		Select(allCountQuery).
		From(queriers.AuditLogEntriesTableName)

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	c.logQueryBuildingError(err)

	builder := c.sqlBuilder.
		Select(append(queriers.AuditLogEntriesTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.AuditLogEntriesTableName).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.AuditLogEntriesTableName)
	}

	query, selectArgs, err := builder.ToSql()
	c.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetAuditLogEntries fetches a list of  from the database that meet a particular filter.
func (c *Sqlite) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	query, args := c.buildGetAuditLogEntriesQuery(filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, count, err := c.scanAuditLogEntries(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AuditLogEntryList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Entries: auditLogEntries,
	}

	return list, nil
}

// buildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (c *Sqlite) buildCreateAuditLogEntryQuery(input *types.AuditLogEntry) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Insert(queriers.AuditLogEntriesTableName).
		Columns(
			queriers.AuditLogEntriesTableEventTypeColumn,
			queriers.AuditLogEntriesTableContextColumn,
		).
		Values(
			input.EventType,
			input.Context,
		).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// createAuditLogEntry creates an audit log entry in the database.
func (c *Sqlite) createAuditLogEntry(ctx context.Context, input *types.AuditLogEntryCreationInput) {
	x := &types.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	query, args := c.buildCreateAuditLogEntryQuery(x)
	c.logger.Debug("createAuditLogEntry called")

	// create the audit log entry.
	if _, err := c.db.ExecContext(ctx, query, args...); err != nil {
		c.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType).Error(err, "executing audit log entry creation query")
	}
}

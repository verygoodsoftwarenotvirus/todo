package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// scanAuditLogEntry takes a database Scanner (i.e. *sql.Row) and scans the result into an AuditLogEntry struct.
func (m *MariaDB) scanAuditLogEntry(scan database.Scanner) (*types.AuditLogEntry, error) {
	x := &types.AuditLogEntry{}

	targetVars := []interface{}{
		&x.ID,
		&x.EventType,
		&x.Context,
		&x.CreatedOn,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	return x, nil
}

// scanAuditLogEntries takes a logger and some database rows and turns them into a slice of .
func (m *MariaDB) scanAuditLogEntries(rows database.ResultIterator) ([]types.AuditLogEntry, error) {
	var (
		list []types.AuditLogEntry
	)

	for rows.Next() {
		x, err := m.scanAuditLogEntry(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		m.logger.Error(closeErr, "closing database rows")
	}

	return list, nil
}

// buildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (m *MariaDB) buildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): entryID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntry fetches an audit log entry from the database.
func (m *MariaDB) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	query, args := m.buildGetAuditLogEntryQuery(entryID)
	row := m.db.QueryRowContext(ctx, query, args...)
	return m.scanAuditLogEntry(row)
}

// buildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (m *MariaDB) buildGetAllAuditLogEntriesCountQuery() string {
	allAuditLogEntriesCountQuery, _, err := m.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.AuditLogEntriesTableName)).
		From(queriers.AuditLogEntriesTableName).
		ToSql()
	m.logQueryBuildingError(err)

	return allAuditLogEntriesCountQuery
}

// GetAllAuditLogEntriesCount will fetch the count of  from the database.
func (m *MariaDB) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	err = m.db.QueryRowContext(ctx, m.buildGetAllAuditLogEntriesCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (m *MariaDB) buildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := m.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetAllAuditLogEntries fetches every audit log entry from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (m *MariaDB) GetAllAuditLogEntries(ctx context.Context, resultChannel chan []types.AuditLogEntry) error {
	count, err := m.GetAllAuditLogEntriesCount(ctx)
	if err != nil {
		return err
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := m.buildGetBatchOfAuditLogEntriesQuery(begin, end)
			logger := m.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := m.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			auditLogEntries, err := m.scanAuditLogEntries(rows)
			if err != nil {
				logger.Error(err, "scanning database rows")
				return
			}

			resultChannel <- auditLogEntries
		}(beginID, endID)
	}

	return nil
}

// buildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (m *MariaDB) buildGetAuditLogEntriesQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, queriers.AuditLogEntriesTableName)
	}

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntries fetches a list of  from the database that meet a particular filter.
func (m *MariaDB) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	query, args := m.buildGetAuditLogEntriesQuery(filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching audit log entries from database: %w", err)
	}

	auditLogEntries, err := m.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AuditLogEntryList{
		Pagination: types.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Entries: auditLogEntries,
	}

	return list, nil
}

// buildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (m *MariaDB) buildCreateAuditLogEntryQuery(input *types.AuditLogEntry) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
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

	m.logQueryBuildingError(err)

	return query, args
}

// createAuditLogEntry creates an audit log entry in the database.
func (m *MariaDB) createAuditLogEntry(ctx context.Context, input *types.AuditLogEntryCreationInput) {
	x := &types.AuditLogEntry{
		EventType: input.EventType,
		Context:   input.Context,
	}

	query, args := m.buildCreateAuditLogEntryQuery(x)
	m.logger.Debug("createAuditLogEntry called")

	// create the audit log entry.
	if _, err := m.db.ExecContext(ctx, query, args...); err != nil {
		m.logger.WithValue("event_type", input.EventType).Error(err, "executing audit log entry creation query")
	}
}

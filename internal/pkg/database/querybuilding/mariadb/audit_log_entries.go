package mariadb

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.AuditLogEntrySQLQueryBuilder = (*MariaDB)(nil)

// BuildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *MariaDB) BuildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): entryID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *MariaDB) BuildGetAllAuditLogEntriesCountQuery() string {
	allAuditLogEntriesCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AuditLogEntriesTableName)).
		From(querybuilding.AuditLogEntriesTableName).
		ToSql()
	q.logQueryBuildingError(err)

	return allAuditLogEntriesCountQuery
}

// BuildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (q *MariaDB) BuildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *MariaDB) BuildGetAuditLogEntriesQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.
		Select(allCountQuery).
		From(querybuilding.AuditLogEntriesTableName)

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(querybuilding.AuditLogEntriesTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(querybuilding.AuditLogEntriesTableName).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, builder, querybuilding.AuditLogEntriesTableName)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// BuildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (q *MariaDB) BuildCreateAuditLogEntryQuery(input *types.AuditLogEntryCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(querybuilding.AuditLogEntriesTableName).
		Columns(
			querybuilding.AuditLogEntriesTableEventTypeColumn,
			querybuilding.AuditLogEntriesTableContextColumn,
		).
		Values(
			input.EventType,
			input.Context,
		).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

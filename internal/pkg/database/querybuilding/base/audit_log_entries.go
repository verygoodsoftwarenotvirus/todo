package base

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.AuditLogEntrySQLQueryBuilder = (*QueryBuilder)(nil)
)

// BuildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *QueryBuilder) BuildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): entryID,
		}),
	)
}

// BuildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *QueryBuilder) BuildGetAllAuditLogEntriesCountQuery() string {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AuditLogEntriesTableName)).
		From(querybuilding.AuditLogEntriesTableName),
	)
}

// BuildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (q *QueryBuilder) BuildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): endID,
		}),
	)
}

// BuildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (q *QueryBuilder) BuildGetAuditLogEntriesQuery(filter *types.QueryFilter) (query string, args []interface{}) {
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
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, querybuilding.AuditLogEntriesTableName, builder)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// BuildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (q *QueryBuilder) BuildCreateAuditLogEntryQuery(input *types.AuditLogEntryCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.AuditLogEntriesTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.AuditLogEntriesTableEventTypeColumn,
			querybuilding.AuditLogEntriesTableContextColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.EventType,
			input.Context,
		),
	)
}

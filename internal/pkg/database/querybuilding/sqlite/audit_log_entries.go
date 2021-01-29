package sqlite

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.AuditLogEntrySQLQueryBuilder = (*Sqlite)(nil)
)

// BuildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (c *Sqlite) BuildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): entryID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (c *Sqlite) BuildGetAllAuditLogEntriesCountQuery() string {
	allAuditLogEntriesCountQuery, _, err := c.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AuditLogEntriesTableName)).
		From(querybuilding.AuditLogEntriesTableName).
		ToSql()
	c.logQueryBuildingError(err)

	return allAuditLogEntriesCountQuery
}

// BuildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (c *Sqlite) BuildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := c.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): endID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (c *Sqlite) BuildGetAuditLogEntriesQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := c.sqlBuilder.
		Select(allCountQuery).
		From(querybuilding.AuditLogEntriesTableName)

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	c.logQueryBuildingError(err)

	builder := c.sqlBuilder.
		Select(append(querybuilding.AuditLogEntriesTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(querybuilding.AuditLogEntriesTableName).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, builder, querybuilding.AuditLogEntriesTableName)
	}

	query, selectArgs, err := builder.ToSql()
	c.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// BuildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (c *Sqlite) BuildCreateAuditLogEntryQuery(input *types.AuditLogEntryCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
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

	c.logQueryBuildingError(err)

	return query, args
}

// createAuditLogEntry creates an audit log entry in the database.
func (c *Sqlite) createAuditLogEntry(ctx context.Context, input *types.AuditLogEntryCreationInput) {
	query, args := c.BuildCreateAuditLogEntryQuery(input)
	c.logger.Debug("createAuditLogEntry called")

	// create the audit log entry.
	if _, err := c.db.ExecContext(ctx, query, args...); err != nil {
		c.logger.WithValue(keys.AuditLogEntryEventTypeKey, input.EventType).Error(err, "executing audit log entry creation query")
	}
}

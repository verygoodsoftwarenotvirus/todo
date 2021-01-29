package mariadb

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.AccountSubscriptionPlanSQLQueryBuilder = (*MariaDB)(nil)

// BuildGetAccountSubscriptionPlanQuery constructs a SQL query for fetching an plan with a given ID belong to a user with a given ID.
func (q *MariaDB) BuildGetAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.AccountSubscriptionPlansTableColumns...).
		From(querybuilding.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.IDColumn):         planID,
			fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllAccountSubscriptionPlansCountQuery returns a query that fetches the total number of plans in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *MariaDB) BuildGetAllAccountSubscriptionPlansCountQuery() string {
	allPlansCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AccountSubscriptionPlansTableName)).
		From(querybuilding.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()
	q.logQueryBuildingError(err)

	return allPlansCountQuery
}

// BuildGetAccountSubscriptionPlansQuery builds a SQL query selecting plans that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *MariaDB) BuildGetAccountSubscriptionPlansQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.AccountSubscriptionPlansTableName,
		"",
		querybuilding.AccountSubscriptionPlansTableColumns,
		0,
		true,
		filter,
	)
}

// BuildCreateAccountSubscriptionPlanQuery takes an plan and returns a creation query for that plan and the relevant arguments.
func (q *MariaDB) BuildCreateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlanCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(querybuilding.AccountSubscriptionPlansTableName).
		Columns(
			querybuilding.AccountSubscriptionPlansTableNameColumn,
			querybuilding.AccountSubscriptionPlansTableDescriptionColumn,
			querybuilding.AccountSubscriptionPlansTablePriceColumn,
			querybuilding.AccountSubscriptionPlansTablePeriodColumn,
		).
		Values(
			input.Name,
			input.Description,
			input.Price,
			input.Period.String(),
		).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateAccountSubscriptionPlanQuery takes an plan and returns an update SQL query, with the relevant query parameters.
func (q *MariaDB) BuildUpdateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.AccountSubscriptionPlansTableName).
		Set(querybuilding.AccountSubscriptionPlansTableNameColumn, input.Name).
		Set(querybuilding.AccountSubscriptionPlansTableDescriptionColumn, input.Description).
		Set(querybuilding.AccountSubscriptionPlansTablePriceColumn, input.Price).
		Set(querybuilding.AccountSubscriptionPlansTablePeriodColumn, input.Period.String()).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn: input.ID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveAccountSubscriptionPlanQuery returns a SQL query which marks a given plan belonging to a given user as archived.
func (q *MariaDB) BuildArchiveAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.AccountSubscriptionPlansTableName).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         planID,
			querybuilding.ArchivedOnColumn: nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery returns a SQL query which retrieves audit log entries for a given account subscription plan.
func (q *MariaDB) BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	builder := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Expr(
			fmt.Sprintf(
				jsonPluckQuery,
				querybuilding.AuditLogEntriesTableName,
				querybuilding.AuditLogEntriesTableContextColumn,
				planID,
				audit.AccountSubscriptionPlanAssignmentKey,
			),
		)).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

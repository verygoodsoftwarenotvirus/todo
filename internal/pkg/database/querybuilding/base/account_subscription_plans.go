package base

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.AccountSubscriptionPlanSQLQueryBuilder = (*QueryBuilder)(nil)
)

// BuildGetAccountSubscriptionPlanQuery constructs a SQL query for fetching an plan with a given ID belong to a user with a given ID.
func (q *QueryBuilder) BuildGetAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AccountSubscriptionPlansTableColumns...).
		From(querybuilding.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.IDColumn):         planID,
			fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetAllAccountSubscriptionPlansCountQuery returns a query that fetches the total number of account subscription plans in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *QueryBuilder) BuildGetAllAccountSubscriptionPlansCountQuery() string {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AccountSubscriptionPlansTableName)).
		From(querybuilding.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetAccountSubscriptionPlansQuery builds a SQL query selecting account subscription plans that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (q *QueryBuilder) BuildGetAccountSubscriptionPlansQuery(filter *types.QueryFilter) (query string, args []interface{}) {
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
func (q *QueryBuilder) BuildCreateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlanCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.AccountSubscriptionPlansTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.AccountSubscriptionPlansTableNameColumn,
			querybuilding.AccountSubscriptionPlansTableDescriptionColumn,
			querybuilding.AccountSubscriptionPlansTablePriceColumn,
			querybuilding.AccountSubscriptionPlansTablePeriodColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Name,
			input.Description,
			input.Price,
			input.Period.String(),
		),
	)
}

// BuildUpdateAccountSubscriptionPlanQuery takes an plan and returns an update SQL query, with the relevant query parameters.
func (q *QueryBuilder) BuildUpdateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountSubscriptionPlansTableName).
		Set(querybuilding.AccountSubscriptionPlansTableNameColumn, input.Name).
		Set(querybuilding.AccountSubscriptionPlansTableDescriptionColumn, input.Description).
		Set(querybuilding.AccountSubscriptionPlansTablePriceColumn, input.Price).
		Set(querybuilding.AccountSubscriptionPlansTablePeriodColumn, input.Period.String()).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         input.ID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildArchiveAccountSubscriptionPlanQuery returns a SQL query which marks a given plan belonging to a given user as archived.
func (q *QueryBuilder) BuildArchiveAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountSubscriptionPlansTableName).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         planID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery constructs a SQL query for fetching audit log entries
// associated with a given plan.
func (q *QueryBuilder) BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	planIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.AccountSubscriptionPlanAssignmentKey)

	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{planIDKey: planID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

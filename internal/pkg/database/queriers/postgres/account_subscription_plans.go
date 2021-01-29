package postgres

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.AccountSubscriptionPlanSQLQueryBuilder = (*Postgres)(nil)

// BuildGetAccountSubscriptionPlanQuery constructs a SQL query for fetching an plan with a given ID belong to a user with a given ID.
func (q *Postgres) BuildGetAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(queriers.AccountSubscriptionPlansTableColumns...).
		From(queriers.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountSubscriptionPlansTableName, queriers.IDColumn):         planID,
			fmt.Sprintf("%s.%s", queriers.AccountSubscriptionPlansTableName, queriers.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetAllAccountSubscriptionPlansCountQuery returns a query that fetches the total number of plans in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Postgres) BuildGetAllAccountSubscriptionPlansCountQuery() string {
	allPlansCountQuery, _ := q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.AccountSubscriptionPlansTableName)).
		From(queriers.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountSubscriptionPlansTableName, queriers.ArchivedOnColumn): nil,
		}),
	)

	return allPlansCountQuery
}

// BuildGetAccountSubscriptionPlansQuery builds a SQL query selecting plans that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Postgres) BuildGetAccountSubscriptionPlansQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		queriers.AccountSubscriptionPlansTableName,
		"",
		queriers.AccountSubscriptionPlansTableColumns,
		0,
		true,
		filter,
	)
}

// BuildCreateAccountSubscriptionPlanQuery takes an plan and returns a creation query for that plan and the relevant arguments.
func (q *Postgres) BuildCreateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlanCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(queriers.AccountSubscriptionPlansTableName).
		Columns(
			queriers.AccountSubscriptionPlansTableNameColumn,
			queriers.AccountSubscriptionPlansTableDescriptionColumn,
			queriers.AccountSubscriptionPlansTablePriceColumn,
			queriers.AccountSubscriptionPlansTablePeriodColumn,
		).
		Values(
			input.Name,
			input.Description,
			input.Price,
			input.Period.String(),
		).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.IDColumn)),
	)
}

// BuildUpdateAccountSubscriptionPlanQuery takes an plan and returns an update SQL query, with the relevant query parameters.
func (q *Postgres) BuildUpdateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(queriers.AccountSubscriptionPlansTableName).
		Set(queriers.AccountSubscriptionPlansTableNameColumn, input.Name).
		Set(queriers.AccountSubscriptionPlansTableDescriptionColumn, input.Description).
		Set(queriers.AccountSubscriptionPlansTablePriceColumn, input.Price).
		Set(queriers.AccountSubscriptionPlansTablePeriodColumn, input.Period.String()).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn: input.ID,
		}),
	)
}

// BuildArchiveAccountSubscriptionPlanQuery returns a SQL query which marks a given plan belonging to a given user as archived.
func (q *Postgres) BuildArchiveAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(queriers.AccountSubscriptionPlansTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:         planID,
			queriers.ArchivedOnColumn: nil,
		}),
	)
}

// BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery constructs a SQL query for fetching audit log entries
// associated with a given plan.
func (q *Postgres) BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	planIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.AccountSubscriptionPlanAssignmentKey)

	return q.buildQuery(q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{planIDKey: planID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn)),
	)
}

package sqlite

import (
	"context"
	"fmt"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ querybuilding.AccountSubscriptionPlanSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildGetAccountSubscriptionPlanQuery constructs a SQL query for fetching an account subscription plan with a given ID belong to a user with a given ID.
func (b *Sqlite) BuildGetAccountSubscriptionPlanQuery(ctx context.Context, accountSubscriptionPlanID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AccountSubscriptionPlansTableColumns...).
			From(querybuilding.AccountSubscriptionPlansTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.IDColumn):         accountSubscriptionPlanID,
				fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.ArchivedOnColumn): nil,
			}),
	)
}

// BuildGetAllAccountSubscriptionPlansCountQuery returns a query that fetches the total number of account subscription plans in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (b *Sqlite) BuildGetAllAccountSubscriptionPlansCountQuery(ctx context.Context) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQueryOnly(span, b.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AccountSubscriptionPlansTableName)).
		From(querybuilding.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountSubscriptionPlansTableName, querybuilding.ArchivedOnColumn): nil,
		}))
}

// BuildGetAccountSubscriptionPlansQuery builds a SQL query selecting account subscription plans that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (b *Sqlite) BuildGetAccountSubscriptionPlansQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}
	return b.buildListQuery(ctx, querybuilding.AccountSubscriptionPlansTableName, "", querybuilding.AccountSubscriptionPlansTableColumns, 0, true, filter)
}

// BuildCreateAccountSubscriptionPlanQuery takes an account subscription plan and returns a creation query for that plan and the relevant arguments.
func (b *Sqlite) BuildCreateAccountSubscriptionPlanQuery(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.AccountSubscriptionPlansTableName).
			Columns(
				querybuilding.ExternalIDColumn,
				querybuilding.AccountSubscriptionPlansTableNameColumn,
				querybuilding.AccountSubscriptionPlansTableDescriptionColumn,
				querybuilding.AccountSubscriptionPlansTablePriceColumn,
				querybuilding.AccountSubscriptionPlansTablePeriodColumn,
			).
			Values(
				b.externalIDGenerator.NewExternalID(),
				input.Name,
				input.Description,
				input.Price,
				input.Period.String(),
			),
	)
}

// BuildUpdateAccountSubscriptionPlanQuery takes an account subscription plan and returns an update SQL query, with the relevant query parameters.
func (b *Sqlite) BuildUpdateAccountSubscriptionPlanQuery(ctx context.Context, input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountSubscriptionPlansTableName).
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
func (b *Sqlite) BuildArchiveAccountSubscriptionPlanQuery(ctx context.Context, accountSubscriptionPlanID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountSubscriptionPlansTableName).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:         accountSubscriptionPlanID,
				querybuilding.ArchivedOnColumn: nil,
			}),
	)
}

// BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery constructs a SQL query for fetching audit log entries
// associated with a given plan.
func (b *Sqlite) BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(ctx context.Context, accountSubscriptionPlanID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	planIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.AccountSubscriptionPlanAssignmentKey)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Eq{planIDKey: accountSubscriptionPlanID}).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

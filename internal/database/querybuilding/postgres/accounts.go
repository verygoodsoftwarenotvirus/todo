package postgres

import (
	"context"
	"fmt"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ querybuilding.AccountSQLQueryBuilder = (*Postgres)(nil)

// BuildGetAccountQuery constructs a SQL query for fetching an account with a given ID belong to a user with a given ID.
func (b *Postgres) BuildGetAccountQuery(ctx context.Context, accountID, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	columns := append(querybuilding.AccountsTableColumns, querybuilding.AccountsUserMembershipTableColumns...)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(columns...).
			From(querybuilding.AccountsTableName).
			Join(fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				querybuilding.AccountsUserMembershipTableName,
				querybuilding.AccountsUserMembershipTableName,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
				querybuilding.AccountsTableName,
				querybuilding.IDColumn,
			)).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.IDColumn):                         accountID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.AccountsTableUserOwnershipColumn): userID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.ArchivedOnColumn):                 nil,
			}),
	)
}

// BuildGetAllAccountsCountQuery returns a query that fetches the total number of accounts in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (b *Postgres) BuildGetAllAccountsCountQuery(ctx context.Context) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	allAccountsCountQuery, _, err := b.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AccountsTableName)).
		From(querybuilding.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()
	b.logQueryBuildingError(span, err)

	return allAccountsCountQuery
}

// BuildGetBatchOfAccountsQuery returns a query that fetches every account in the database within a bucketed range.
func (b *Postgres) BuildGetBatchOfAccountsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AccountsTableColumns...).
			From(querybuilding.AccountsTableName).
			Where(squirrel.Gt{
				fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.IDColumn): beginID,
			}).
			Where(squirrel.Lt{
				fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.IDColumn): endID,
			}),
	)
}

// BuildGetAccountsQuery builds a SQL query selecting accounts that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (b *Postgres) BuildGetAccountsQuery(ctx context.Context, userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	var includeArchived bool
	if filter != nil {
		includeArchived = filter.IncludeArchived
	}

	columns := append(querybuilding.AccountsTableColumns, querybuilding.AccountsUserMembershipTableColumns...)
	filteredCountQuery, filteredCountQueryArgs := b.buildFilteredCountQuery(ctx, querybuilding.AccountsTableName, querybuilding.AccountsTableUserOwnershipColumn, userID, forAdmin, includeArchived, filter)
	totalCountQuery, totalCountQueryArgs := b.buildTotalCountQuery(ctx, querybuilding.AccountsTableName, querybuilding.AccountsTableUserOwnershipColumn, userID, forAdmin, includeArchived)

	builder := b.sqlBuilder.
		Select(append(
			columns,
			fmt.Sprintf("(%s) as total_count", totalCountQuery),
			fmt.Sprintf("(%s) as filtered_count", filteredCountQuery),
		)...).
		From(querybuilding.AccountsTableName).
		Join(fmt.Sprintf(
			"%s ON %s.%s = %s.%s",
			querybuilding.AccountsUserMembershipTableName,
			querybuilding.AccountsUserMembershipTableName,
			querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
			querybuilding.AccountsTableName,
			querybuilding.IDColumn,
		))

	if !forAdmin {
		builder = builder.Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.ArchivedOnColumn):                 nil,
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.AccountsTableUserOwnershipColumn): userID,
		})
	}

	builder = builder.GroupBy(fmt.Sprintf(
		"(%s.%s, %s.%s)",
		querybuilding.AccountsTableName,
		querybuilding.IDColumn,
		querybuilding.AccountsUserMembershipTableName,
		querybuilding.IDColumn,
	))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, querybuilding.AccountsTableName, builder)
	}

	query, selectArgs := b.buildQuery(span, builder)

	return query, append(append(filteredCountQueryArgs, totalCountQueryArgs...), selectArgs...)
}

// BuildAccountCreationQuery takes an account and returns a creation query for that account and the relevant arguments.
func (b *Postgres) BuildAccountCreationQuery(ctx context.Context, input *types.AccountCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.AccountsTableName).
			Columns(
				querybuilding.ExternalIDColumn,
				querybuilding.AccountsTableNameColumn,
				querybuilding.AccountsTableBillingStatusColumn,
				querybuilding.AccountsTableContactEmailColumn,
				querybuilding.AccountsTableContactPhoneColumn,
				querybuilding.AccountsTableUserOwnershipColumn,
			).
			Values(
				b.externalIDGenerator.NewExternalID(),
				input.Name,
				types.UnpaidAccountBillingStatus,
				input.ContactEmail,
				input.ContactPhone,
				input.BelongsToUser,
			).
			Suffix(fmt.Sprintf("RETURNING %s", querybuilding.IDColumn)),
	)
}

// BuildUpdateAccountQuery takes an account and returns an update SQL query, with the relevant query parameters.
func (b *Postgres) BuildUpdateAccountQuery(ctx context.Context, input *types.Account) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, input.ID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsTableName).
			Set(querybuilding.AccountsTableNameColumn, input.Name).
			Set(querybuilding.AccountsTableContactEmailColumn, input.ContactEmail).
			Set(querybuilding.AccountsTableContactPhoneColumn, input.ContactPhone).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                         input.ID,
				querybuilding.ArchivedOnColumn:                 nil,
				querybuilding.AccountsTableUserOwnershipColumn: input.BelongsToUser,
			}),
	)
}

// BuildArchiveAccountQuery returns a SQL query which marks a given account belonging to a given user as archived.
func (b *Postgres) BuildArchiveAccountQuery(ctx context.Context, accountID, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsTableName).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                         accountID,
				querybuilding.ArchivedOnColumn:                 nil,
				querybuilding.AccountsTableUserOwnershipColumn: userID,
			}),
	)
}

// BuildGetAuditLogEntriesForAccountQuery constructs a SQL query for fetching audit log entries
// associated with a given account.
func (b *Postgres) BuildGetAuditLogEntriesForAccountQuery(ctx context.Context, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, accountID)

	accountIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.AccountAssignmentKey)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Eq{accountIDKey: accountID}).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

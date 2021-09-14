package mysql

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ querybuilding.ItemSQLQueryBuilder = (*MySQL)(nil)

// BuildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (b *MySQL) BuildItemExistsQuery(ctx context.Context, itemID, accountID string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn)).
			Prefix(querybuilding.ExistencePrefix).
			From(querybuilding.ItemsTableName).
			Suffix(querybuilding.ExistenceSuffix).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):                         itemID,
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn):                 nil,
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableAccountOwnershipColumn): accountID,
			}),
	)
}

// BuildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (b *MySQL) BuildGetItemQuery(ctx context.Context, itemID, accountID string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.ItemsTableColumns...).
			From(querybuilding.ItemsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):                         itemID,
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn):                 nil,
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableAccountOwnershipColumn): accountID,
			}),
	)
}

// BuildGetTotalItemCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (b *MySQL) BuildGetTotalItemCountQuery(ctx context.Context) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQueryOnly(
		span,
		b.sqlBuilder.Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.ItemsTableName)).
			From(querybuilding.ItemsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
			}),
	)
}

// BuildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (b *MySQL) BuildGetBatchOfItemsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.ItemsTableColumns...).
			From(querybuilding.ItemsTableName).
			Where(squirrel.Gt{
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn): beginID,
			}).
			Where(squirrel.Lt{
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn): endID,
			}),
	)
}

// BuildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (b *MySQL) BuildGetItemsQuery(ctx context.Context, accountID string, includeArchived bool, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
	}

	return b.buildListQuery(
		ctx,
		querybuilding.ItemsTableName,
		nil,
		where,
		querybuilding.ItemsTableAccountOwnershipColumn,
		querybuilding.ItemsTableColumns,
		accountID,
		includeArchived,
		filter,
	)
}

// BuildGetItemsWithIDsQuery builds a SQL query selecting items that belong to a given account,
// and have IDs that exist within a given set of IDs. Returns both the query and the relevant
// args to pass to the query executor. This function is primarily intended for use with a search
// index, which would provide a slice of string IDs to query against. This function accepts a
// slice of uint64s instead of a slice of strings in order to ensure all the provided strings
// are valid database IDs, because there's no way in squirrel to escape them in the unnest join,
// and if we accept strings we could leave ourselves vulnerable to SQL injection attacks.
func (b *MySQL) BuildGetItemsWithIDsQuery(ctx context.Context, accountID string, limit uint8, ids []string, restrictToAccount bool) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, accountID)

	whenThenStatement := joinStringIDs(ids)
	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):         ids,
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
	}

	if restrictToAccount {
		where[fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableAccountOwnershipColumn)] = accountID
	}

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.ItemsTableColumns...).
			From(querybuilding.ItemsTableName).
			Where(where).
			OrderBy(fmt.Sprintf("CASE %s.%s %s", querybuilding.ItemsTableName, querybuilding.IDColumn, whenThenStatement)).
			Limit(uint64(limit)),
	)
}

// BuildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (b *MySQL) BuildCreateItemQuery(ctx context.Context, input *types.ItemDatabaseCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.ItemsTableName).
			Columns(
				querybuilding.IDColumn,
				querybuilding.ItemsTableNameColumn,
				querybuilding.ItemsTableDetailsColumn,
				querybuilding.ItemsTableAccountOwnershipColumn,
			).
			Values(
				input.ID,
				input.Name,
				input.Details,
				input.BelongsToAccount,
			),
	)
}

// BuildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (b *MySQL) BuildUpdateItemQuery(ctx context.Context, input *types.Item) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, input.ID)
	tracing.AttachAccountIDToSpan(span, input.BelongsToAccount)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.ItemsTableName).
			Set(querybuilding.ItemsTableNameColumn, input.Name).
			Set(querybuilding.ItemsTableDetailsColumn, input.Details).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                         input.ID,
				querybuilding.ArchivedOnColumn:                 nil,
				querybuilding.ItemsTableAccountOwnershipColumn: input.BelongsToAccount,
			}),
	)
}

// BuildArchiveItemQuery returns a SQL query which marks a given item belonging to a given account as archived.
func (b *MySQL) BuildArchiveItemQuery(ctx context.Context, itemID, accountID string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.ItemsTableName).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                         itemID,
				querybuilding.ArchivedOnColumn:                 nil,
				querybuilding.ItemsTableAccountOwnershipColumn: accountID,
			}),
	)
}

// BuildGetAuditLogEntriesForItemQuery constructs a SQL query for fetching audit log entries relating to an item with a given ID.
func (b *MySQL) BuildGetAuditLogEntriesForItemQuery(ctx context.Context, itemID string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)

	itemIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		itemID,
		audit.ItemAssignmentKey,
	)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Expr(itemIDKey)).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

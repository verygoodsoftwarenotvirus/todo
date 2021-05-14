package mariadb

import (
	"context"
	"fmt"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ querybuilding.ItemSQLQueryBuilder = (*MariaDB)(nil)

// BuildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (b *MariaDB) BuildItemExistsQuery(ctx context.Context, itemID, accountID uint64) (query string, args []interface{}) {
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
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableAccountOwnershipColumn): accountID,
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn):                 nil,
			}),
	)
}

// BuildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (b *MariaDB) BuildGetItemQuery(ctx context.Context, itemID, accountID uint64) (query string, args []interface{}) {
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
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableAccountOwnershipColumn): accountID,
				fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn):                 nil,
			}),
	)
}

// BuildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (b *MariaDB) BuildGetAllItemsCountQuery(ctx context.Context) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQueryOnly(span, b.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.ItemsTableName)).
		From(querybuilding.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
		}))
}

// BuildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (b *MariaDB) BuildGetBatchOfItemsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
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
func (b *MariaDB) BuildGetItemsQuery(ctx context.Context, accountID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}
	return b.buildListQuery(ctx, querybuilding.ItemsTableName, querybuilding.ItemsTableAccountOwnershipColumn, querybuilding.ItemsTableColumns, accountID, forAdmin, filter)
}

// BuildGetItemsWithIDsQuery builds a SQL query selecting items that belong to a given account,
// and have IDs that exist within a given set of IDs. Returns both the query and the relevant
// args to pass to the query executor. This function is primarily intended for use with a search
// index, which would provide a slice of string IDs to query against. This function accepts a
// slice of uint64s instead of a slice of strings in order to ensure all the provided strings
// are valid database IDs, because there's no way in squirrel to escape them in the unnest join,
// and if we accept strings we could leave ourselves vulnerable to SQL injection attacks.
func (b *MariaDB) BuildGetItemsWithIDsQuery(ctx context.Context, accountID uint64, limit uint8, ids []uint64, forAdmin bool) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, accountID)

	whenThenStatement := buildWhenThenStatement(ids)

	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):         ids,
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
	}

	if !forAdmin {
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
func (b *MariaDB) BuildCreateItemQuery(ctx context.Context, input *types.ItemCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.ItemsTableName).
			Columns(
				querybuilding.ExternalIDColumn,
				querybuilding.ItemsTableNameColumn,
				querybuilding.ItemsTableDetailsColumn,
				querybuilding.ItemsTableAccountOwnershipColumn,
			).
			Values(
				b.externalIDGenerator.NewExternalID(),
				input.Name,
				input.Details,
				input.BelongsToAccount,
			),
	)
}

// BuildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (b *MariaDB) BuildUpdateItemQuery(ctx context.Context, input *types.Item) (query string, args []interface{}) {
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

// BuildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (b *MariaDB) BuildArchiveItemQuery(ctx context.Context, itemID, accountID uint64) (query string, args []interface{}) {
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

// BuildGetAuditLogEntriesForItemQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (b *MariaDB) BuildGetAuditLogEntriesForItemQuery(ctx context.Context, itemID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					querybuilding.AuditLogEntriesTableName,
					querybuilding.AuditLogEntriesTableContextColumn,
					itemID,
					audit.ItemAssignmentKey,
				),
			)).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

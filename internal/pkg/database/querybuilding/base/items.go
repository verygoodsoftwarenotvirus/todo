package sqlite

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.ItemSQLQueryBuilder = (*BaseQueryBuilder)(nil)
)

// BuildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (q *BaseQueryBuilder) BuildItemExistsQuery(itemID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn)).
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
func (q *BaseQueryBuilder) BuildGetItemQuery(itemID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.ItemsTableColumns...).
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
func (q *BaseQueryBuilder) BuildGetAllItemsCountQuery() string {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.ItemsTableName)).
		From(querybuilding.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *BaseQueryBuilder) BuildGetBatchOfItemsQuery(beginID, endID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.ItemsTableColumns...).
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
func (q *BaseQueryBuilder) BuildGetItemsQuery(accountID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.ItemsTableName,
		querybuilding.ItemsTableAccountOwnershipColumn,
		querybuilding.ItemsTableColumns,
		accountID,
		forAdmin,
		filter,
	)
}

func buildWhenThenStatement(ids []uint64) string {
	statement := ""

	for i, id := range ids {
		if i != 0 {
			statement += " "
		}

		statement += fmt.Sprintf("WHEN %d THEN %d", id, i)
	}

	statement += " END"

	return statement
}

// BuildGetItemsWithIDsQuery builds a SQL query selecting items that belong to a given account,
// and have IDs that exist within a given set of IDs. Returns both the query and the relevant
// args to pass to the query executor. This function is primarily intended for use with a search
// index, which would provide a slice of string IDs to query against. This function accepts a
// slice of uint64s instead of a slice of strings in order to ensure all the provided strings
// are valid database IDs, because there's no way in squirrel to escape them in the unnest join,
// and if we accept strings we could leave ourselves vulnerable to SQL injection attacks.
func (q *BaseQueryBuilder) BuildGetItemsWithIDsQuery(accountID uint64, limit uint8, ids []uint64, forAdmin bool) (query string, args []interface{}) {
	whenThenStatement := buildWhenThenStatement(ids)

	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):         ids,
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
	}

	if !forAdmin {
		where[fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableAccountOwnershipColumn)] = accountID
	}

	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.ItemsTableColumns...).
		From(querybuilding.ItemsTableName).
		Where(where).
		OrderBy(fmt.Sprintf("CASE %s.%s %s", querybuilding.ItemsTableName, querybuilding.IDColumn, whenThenStatement)).
		Limit(uint64(limit)),
	)
}

// BuildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (q *BaseQueryBuilder) BuildCreateItemQuery(input *types.ItemCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.ItemsTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.ItemsTableNameColumn,
			querybuilding.ItemsTableDetailsColumn,
			querybuilding.ItemsTableAccountOwnershipColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Name,
			input.Details,
			input.BelongsToAccount,
		),
	)
}

// BuildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (q *BaseQueryBuilder) BuildUpdateItemQuery(input *types.Item) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.ItemsTableName).
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
func (q *BaseQueryBuilder) BuildArchiveItemQuery(itemID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.ItemsTableName).
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
func (q *BaseQueryBuilder) BuildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{}) {
	itemIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.ItemAssignmentKey)

	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{itemIDKey: itemID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

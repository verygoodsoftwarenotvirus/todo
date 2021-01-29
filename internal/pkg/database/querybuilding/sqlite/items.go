package sqlite

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.ItemSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (c *Sqlite) BuildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn)).
		Prefix(querybuilding.ExistencePrefix).
		From(querybuilding.ItemsTableName).
		Suffix(querybuilding.ExistenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn):              nil,
		}).ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (c *Sqlite) BuildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(querybuilding.ItemsTableColumns...).
		From(querybuilding.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn):              nil,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (c *Sqlite) BuildGetAllItemsCountQuery() string {
	var err error

	allItemsCountQuery, _, err := c.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.ItemsTableName)).
		From(querybuilding.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()
	c.logQueryBuildingError(err)

	return allItemsCountQuery
}

// BuildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (c *Sqlite) BuildGetBatchOfItemsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := c.sqlBuilder.
		Select(querybuilding.ItemsTableColumns...).
		From(querybuilding.ItemsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn): endID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (c *Sqlite) BuildGetItemsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	return c.buildListQuery(
		querybuilding.ItemsTableName,
		querybuilding.ItemsTableUserOwnershipColumn,
		querybuilding.ItemsTableColumns,
		userID,
		forAdmin,
		filter,
	)
}

// BuildGetItemsWithIDsQuery builds a SQL query selecting items that belong to a given user,
// and have IDs that exist within a given set of IDs. Returns both the query and the relevant
// args to pass to the query executor. This function is primarily intended for use with a search
// index, which would provide a slice of string IDs to query against. This function accepts a
// slice of uint64s instead of a slice of strings in order to ensure all the provided strings
// are valid database IDs, because there's no way in squirrel to escape them in the unnest join,
// and if we accept strings we could leave ourselves vulnerable to SQL injection attacks.
func (c *Sqlite) BuildGetItemsWithIDsQuery(userID uint64, limit uint8, ids []uint64, forAdmin bool) (query string, args []interface{}) {
	var (
		err               error
		whenThenStatement string
	)

	for i, id := range ids {
		if i != 0 {
			whenThenStatement += " "
		}

		whenThenStatement += fmt.Sprintf("WHEN %d THEN %d", id, i)
	}

	whenThenStatement += " END"

	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.IDColumn):         ids,
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
	}
	if !forAdmin {
		where[fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ItemsTableUserOwnershipColumn)] = userID
	}

	builder := c.sqlBuilder.
		Select(querybuilding.ItemsTableColumns...).
		From(querybuilding.ItemsTableName).
		Where(where).
		OrderBy(fmt.Sprintf("CASE %s.%s %s", querybuilding.ItemsTableName, querybuilding.IDColumn, whenThenStatement)).
		Limit(uint64(limit))

	query, args, err = builder.ToSql()
	c.logQueryBuildingError(err)

	return query, args
}

// BuildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (c *Sqlite) BuildCreateItemQuery(input *types.ItemCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Insert(querybuilding.ItemsTableName).
		Columns(
			querybuilding.ItemsTableNameColumn,
			querybuilding.ItemsTableDetailsColumn,
			querybuilding.ItemsTableUserOwnershipColumn,
		).
		Values(
			input.Name,
			input.Details,
			input.BelongsToUser,
		).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (c *Sqlite) BuildUpdateItemQuery(input *types.Item) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Update(querybuilding.ItemsTableName).
		Set(querybuilding.ItemsTableNameColumn, input.Name).
		Set(querybuilding.ItemsTableDetailsColumn, input.Details).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                      input.ID,
			querybuilding.ItemsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (c *Sqlite) BuildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Update(querybuilding.ItemsTableName).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                      itemID,
			querybuilding.ArchivedOnColumn:              nil,
			querybuilding.ItemsTableUserOwnershipColumn: userID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForItemQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (c *Sqlite) BuildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{}) {
	var err error

	itemIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.ItemAssignmentKey)
	builder := c.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{itemIDKey: itemID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	query, args, err = builder.ToSql()
	c.logQueryBuildingError(err)

	return query, args
}

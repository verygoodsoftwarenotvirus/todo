package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.ItemDataManager  = (*Sqlite)(nil)
	_ types.ItemAuditManager = (*Sqlite)(nil)
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct.
func (c *Sqlite) scanItem(scan database.Scanner, includeCounts bool) (x *types.Item, filteredCount, totalCount uint64, err error) {
	x = &types.Item{}

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	return x, filteredCount, totalCount, nil
}

// scanItems takes some database rows and turns them into a slice of items.
func (c *Sqlite) scanItems(rows database.ResultIterator, includeCounts bool) (items []*types.Item, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		x, fc, tc, scanErr := c.scanItem(rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, scanErr
		}

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}

		items = append(items, x)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing database rows")
		return nil, 0, 0, closeErr
	}

	return items, filteredCount, totalCount, nil
}

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (c *Sqlite) buildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn)).
		Prefix(queriers.ExistencePrefix).
		From(queriers.ItemsTableName).
		Suffix(queriers.ExistenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
		}).ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// ItemExists queries the database to see if a given item belonging to a given user exists.
func (c *Sqlite) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	query, args := c.buildItemExistsQuery(itemID, userID)

	err = c.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (c *Sqlite) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the database.
func (c *Sqlite) GetItem(ctx context.Context, itemID, userID uint64) (*types.Item, error) {
	query, args := c.buildGetItemQuery(itemID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	item, _, _, err := c.scanItem(row, false)

	return item, err
}

// buildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (c *Sqlite) buildGetAllItemsCountQuery() string {
	var err error

	allItemsCountQuery, _, err := c.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.ItemsTableName)).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	c.logQueryBuildingError(err)

	return allItemsCountQuery
}

// GetAllItemsCount will fetch the count of items from the database.
func (c *Sqlite) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = c.db.QueryRowContext(ctx, c.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (c *Sqlite) buildGetBatchOfItemsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := c.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetAllItems fetches every item from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (c *Sqlite) GetAllItems(ctx context.Context, resultChannel chan []*types.Item, batchSize uint16) error {
	count, countErr := c.GetAllItemsCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of items: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.buildGetBatchOfItemsQuery(begin, end)
			logger := c.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			items, _, _, scanErr := c.scanItems(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			resultChannel <- items
		}(beginID, endID)
	}

	return nil
}

// buildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (c *Sqlite) buildGetItemsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	return c.buildListQuery(
		queriers.ItemsTableName,
		queriers.ItemsTableUserOwnershipColumn,
		queriers.ItemsTableColumns,
		userID,
		forAdmin,
		filter,
	)
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (c *Sqlite) GetItems(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := c.buildGetItemsQuery(userID, false, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for items: %w", err)
	}

	items, filteredCount, totalCount, err := c.scanItems(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: filteredCount,
			TotalCount:    totalCount,
		},
		Items: items,
	}

	return list, nil
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (c *Sqlite) GetItemsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := c.buildGetItemsQuery(0, true, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, filteredCount, totalCount, err := c.scanItems(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: filteredCount,
			TotalCount:    totalCount,
		},
		Items: items,
	}

	return list, nil
}

// buildGetItemsWithIDsQuery builds a SQL query selecting items that belong to a given user,
// and have IDs that exist within a given set of IDs. Returns both the query and the relevant
// args to pass to the query executor. This function is primarily intended for use with a search
// index, which would provide a slice of string IDs to query against. This function accepts a
// slice of uint64s instead of a slice of strings in order to ensure all the provided strings
// are valid database IDs, because there's no way in squirrel to escape them in the unnest join,
// and if we accept strings we could leave ourselves vulnerable to SQL injection attacks.
func (c *Sqlite) buildGetItemsWithIDsQuery(userID uint64, limit uint8, ids []uint64) (query string, args []interface{}) {
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

	builder := c.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      ids,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		}).
		OrderBy(fmt.Sprintf("CASE %s.%s %s", queriers.ItemsTableName, queriers.IDColumn, whenThenStatement)).
		Limit(uint64(limit))

	query, args, err = builder.ToSql()
	c.logQueryBuildingError(err)

	return query, args
}

// GetItemsWithIDs fetches a list of items from the database that exist within a given set of IDs.
func (c *Sqlite) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]*types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := c.buildGetItemsWithIDsQuery(userID, limit, ids)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, _, _, err := c.scanItems(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return items, nil
}

// buildGetItemsWithIDsForAdminQuery builds a SQL query selecting items that exist within a given set of IDs.
// Returns both the query and the relevant args to pass to the query executor.
// This function is primarily intended for use with a search index, which would provide a slice of string IDs to query against.
// This function accepts a slice of uint64s instead of a slice of strings in order to ensure all the provided strings
// are valid database IDs, because there's no way in squirrel to escape them in the unnest join,
// and if we accept strings we could leave ourselves vulnerable to SQL injection attacks.
func (c *Sqlite) buildGetItemsWithIDsForAdminQuery(limit uint8, ids []uint64) (query string, args []interface{}) {
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

	builder := c.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):         ids,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("CASE %s.%s %s", queriers.ItemsTableName, queriers.IDColumn, whenThenStatement)).
		Limit(uint64(limit))

	query, args, err = builder.ToSql()
	c.logQueryBuildingError(err)

	return query, args
}

// GetItemsWithIDsForAdmin fetches a list of items from the database that exist within a given set of IDs.
func (c *Sqlite) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]*types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := c.buildGetItemsWithIDsForAdminQuery(limit, ids)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, _, _, err := c.scanItems(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return items, nil
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (c *Sqlite) buildCreateItemQuery(input *types.Item) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Insert(queriers.ItemsTableName).
		Columns(
			queriers.ItemsTableNameColumn,
			queriers.ItemsTableDetailsColumn,
			queriers.ItemsTableUserOwnershipColumn,
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

// CreateItem creates an item in the database.
func (c *Sqlite) CreateItem(ctx context.Context, input *types.ItemCreationInput) (*types.Item, error) {
	x := &types.Item{
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := c.buildCreateItemQuery(x)

	// create the item.
	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing item creation query: %w", err)
	}

	x.ID, x.CreatedOn = c.getIDFromResult(res), c.timeTeller.Now()

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (c *Sqlite) buildUpdateItemQuery(input *types.Item) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.ItemsTableNameColumn, input.Name).
		Set(queriers.ItemsTableDetailsColumn, input.Details).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      input.ID,
			queriers.ItemsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (c *Sqlite) UpdateItem(ctx context.Context, input *types.Item) error {
	query, args := c.buildUpdateItemQuery(input)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (c *Sqlite) buildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      itemID,
			queriers.ArchivedOnColumn:              nil,
			queriers.ItemsTableUserOwnershipColumn: userID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// ArchiveItem marks an item as archived in the database.
func (c *Sqlite) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	query, args := c.buildArchiveItemQuery(itemID, userID)

	res, err := c.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogItemCreationEvent saves a ItemCreationEvent in the audit log table.
func (c *Sqlite) LogItemCreationEvent(ctx context.Context, item *types.Item) {
	c.createAuditLogEntry(ctx, audit.BuildItemCreationEventEntry(item))
}

// LogItemUpdateEvent saves a ItemUpdateEvent in the audit log table.
func (c *Sqlite) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []types.FieldChangeSummary) {
	c.createAuditLogEntry(ctx, audit.BuildItemUpdateEventEntry(userID, itemID, changes))
}

// LogItemArchiveEvent saves a ItemArchiveEvent in the audit log table.
func (c *Sqlite) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	c.createAuditLogEntry(ctx, audit.BuildItemArchiveEventEntry(userID, itemID))
}

// buildGetAuditLogEntriesForItemQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (c *Sqlite) buildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{}) {
	var err error

	itemIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.ItemAssignmentKey)
	builder := c.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{itemIDKey: itemID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	query, args, err = builder.ToSql()
	c.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForItem fetches an audit log entry from the database.
func (c *Sqlite) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]*types.AuditLogEntry, error) {
	query, args := c.buildGetAuditLogEntriesForItemQuery(itemID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

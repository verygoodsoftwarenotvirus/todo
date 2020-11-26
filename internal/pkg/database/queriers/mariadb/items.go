package mariadb

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

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct.
func (q *MariaDB) scanItem(scan database.Scanner, includeCount bool) (*types.Item, uint64, error) {
	var (
		x     = &types.Item{}
		count uint64
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	return x, count, nil
}

// scanItems takes a logger and some database rows and turns them into a slice of items.
func (q *MariaDB) scanItems(rows database.ResultIterator, includeCount bool) ([]types.Item, uint64, error) {
	var (
		list  []types.Item
		count uint64
	)

	for rows.Next() {
		x, c, err := q.scanItem(rows, includeCount)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 && includeCount {
			count = c
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		q.logger.Error(closeErr, "closing database rows")
	}

	return list, count, nil
}

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (q *MariaDB) buildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn)).
		Prefix(queriers.ExistencePrefix).
		From(queriers.ItemsTableName).
		Suffix(queriers.ExistenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ItemExists queries the database to see if a given item belonging to a given user exists.
func (q *MariaDB) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	query, args := q.buildItemExistsQuery(itemID, userID)

	err = q.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (q *MariaDB) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the database.
func (q *MariaDB) GetItem(ctx context.Context, itemID, userID uint64) (*types.Item, error) {
	query, args := q.buildGetItemQuery(itemID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	item, _, err := q.scanItem(row, false)

	return item, err
}

// buildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *MariaDB) buildGetAllItemsCountQuery() string {
	var err error

	allItemsCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.ItemsTableName)).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	q.logQueryBuildingError(err)

	return allItemsCountQuery
}

// GetAllItemsCount will fetch the count of items from the database.
func (q *MariaDB) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *MariaDB) buildGetBatchOfItemsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAllItems fetches every item from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (q *MariaDB) GetAllItems(ctx context.Context, resultChannel chan []types.Item) error {
	count, err := q.GetAllItemsCount(ctx)
	if err != nil {
		return err
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := q.buildGetBatchOfItemsQuery(begin, end)
			logger := q.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := q.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			items, _, err := q.scanItems(rows, false)
			if err != nil {
				logger.Error(err, "scanning database rows")
				return
			}

			resultChannel <- items
		}(beginID, endID)
	}

	return nil
}

// buildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *MariaDB) buildGetItemsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.
		Select(allCountQuery).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		})

	if filter != nil {
		countQueryBuilder = queriers.ApplyFilterToSubCountQueryBuilder(filter, countQueryBuilder, queriers.ItemsTableName)
	}

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(queriers.ItemsTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.ItemsTableName)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (q *MariaDB) GetItems(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := q.buildGetItemsQuery(userID, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, count, err := q.scanItems(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Items: items,
	}

	return list, nil
}

// buildGetItemsForAdminQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *MariaDB) buildGetItemsForAdminQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	if filter == nil {
		filter = types.DefaultQueryFilter()
	}

	where := squirrel.Eq{}
	if !filter.IncludeArchived {
		where[fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn)] = nil
	}

	countQueryBuilder := queriers.ApplyFilterToSubCountQueryBuilder(
		filter,
		q.sqlBuilder.
			Select(allCountQuery).
			From(queriers.ItemsTableName).
			Where(where),
		queriers.ItemsTableName,
	)

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := queriers.ApplyFilterToQueryBuilder(
		filter,
		q.sqlBuilder.
			Select(append(queriers.ItemsTableColumns, fmt.Sprintf("(%s)", countQuery))...).
			From(queriers.ItemsTableName).
			Where(where).
			OrderBy(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn)),
		queriers.ItemsTableName,
	)

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (q *MariaDB) GetItemsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := q.buildGetItemsForAdminQuery(filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, count, err := q.scanItems(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
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
func (q *MariaDB) buildGetItemsWithIDsQuery(userID uint64, limit uint8, ids []uint64) (query string, args []interface{}) {
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

	builder := q.sqlBuilder.
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
	q.logQueryBuildingError(err)

	return query, args
}

// GetItemsWithIDs fetches a list of items from the database that exist within a given set of IDs.
func (q *MariaDB) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := q.buildGetItemsWithIDsQuery(userID, limit, ids)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, _, err := q.scanItems(rows, false)
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
func (q *MariaDB) buildGetItemsWithIDsForAdminQuery(limit uint8, ids []uint64) (query string, args []interface{}) {
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

	builder := q.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):         ids,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("CASE %s.%s %s", queriers.ItemsTableName, queriers.IDColumn, whenThenStatement)).
		Limit(uint64(limit))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// GetItemsWithIDsForAdmin fetches a list of items from the database that exist within a given set of IDs.
func (q *MariaDB) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := q.buildGetItemsWithIDsForAdminQuery(limit, ids)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, _, err := q.scanItems(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return items, nil
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (q *MariaDB) buildCreateItemQuery(input *types.Item) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
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

	q.logQueryBuildingError(err)

	return query, args
}

// CreateItem creates an item in the database.
func (q *MariaDB) CreateItem(ctx context.Context, input *types.ItemCreationInput) (*types.Item, error) {
	x := &types.Item{
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := q.buildCreateItemQuery(x)

	// create the item.
	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing item creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	q.logIDRetrievalError(err)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = q.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (q *MariaDB) buildUpdateItemQuery(input *types.Item) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.ItemsTableNameColumn, input.Name).
		Set(queriers.ItemsTableDetailsColumn, input.Details).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      input.ID,
			queriers.ItemsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (q *MariaDB) UpdateItem(ctx context.Context, input *types.Item) error {
	query, args := q.buildUpdateItemQuery(input)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (q *MariaDB) buildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      itemID,
			queriers.ArchivedOnColumn:              nil,
			queriers.ItemsTableUserOwnershipColumn: userID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchiveItem marks an item as archived in the database.
func (q *MariaDB) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	query, args := q.buildArchiveItemQuery(itemID, userID)

	res, err := q.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogItemCreationEvent saves a ItemCreationEvent in the audit log table.
func (q *MariaDB) LogItemCreationEvent(ctx context.Context, item *types.Item) {
	q.createAuditLogEntry(ctx, audit.BuildItemCreationEventEntry(item))
}

// LogItemUpdateEvent saves a ItemUpdateEvent in the audit log table.
func (q *MariaDB) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []types.FieldChangeSummary) {
	q.createAuditLogEntry(ctx, audit.BuildItemUpdateEventEntry(userID, itemID, changes))
}

// LogItemArchiveEvent saves a ItemArchiveEvent in the audit log table.
func (q *MariaDB) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildItemArchiveEventEntry(userID, itemID))
}

// buildGetAuditLogEntriesForItemQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *MariaDB) buildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{}) {
	var err error

	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					queriers.AuditLogEntriesTableName,
					queriers.AuditLogEntriesTableContextColumn,
					itemID,
					audit.ItemAssignmentKey,
				),
			),
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					queriers.AuditLogEntriesTableName,
					queriers.AuditLogEntriesTableContextColumn,
					itemID,
					audit.ActorAssignmentKey,
				),
			),
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForItem fetches an audit log entry from the database.
func (q *MariaDB) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]types.AuditLogEntry, error) {
	query, args := q.buildGetAuditLogEntriesForItemQuery(itemID)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

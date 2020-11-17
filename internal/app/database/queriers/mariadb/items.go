package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct.
func (m *MariaDB) scanItem(scan database.Scanner) (*types.Item, error) {
	x := &types.Item{}

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	return x, nil
}

// scanItems takes a logger and some database rows and turns them into a slice of items.
func (m *MariaDB) scanItems(rows database.ResultIterator) ([]types.Item, error) {
	var (
		list []types.Item
	)

	for rows.Next() {
		x, err := m.scanItem(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		m.logger.Error(closeErr, "closing database rows")
	}

	return list, nil
}

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (m *MariaDB) buildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn)).
		Prefix(existencePrefix).
		From(queriers.ItemsTableName).
		Suffix(existenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		}).ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ItemExists queries the database to see if a given item belonging to a given user exists.
func (m *MariaDB) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	query, args := m.buildItemExistsQuery(itemID, userID)

	err = m.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (m *MariaDB) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the database.
func (m *MariaDB) GetItem(ctx context.Context, itemID, userID uint64) (*types.Item, error) {
	query, args := m.buildGetItemQuery(itemID, userID)
	row := m.db.QueryRowContext(ctx, query, args...)
	return m.scanItem(row)
}

// buildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (m *MariaDB) buildGetAllItemsCountQuery() string {
	var err error

	allItemsCountQuery, _, err := m.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.ItemsTableName)).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	m.logQueryBuildingError(err)

	return allItemsCountQuery
}

// GetAllItemsCount will fetch the count of items from the database.
func (m *MariaDB) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = m.db.QueryRowContext(ctx, m.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (m *MariaDB) buildGetBatchOfItemsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := m.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetAllItems fetches every item from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (m *MariaDB) GetAllItems(ctx context.Context, resultChannel chan []types.Item) error {
	count, err := m.GetAllItemsCount(ctx)
	if err != nil {
		return err
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := m.buildGetBatchOfItemsQuery(begin, end)
			logger := m.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := m.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			items, err := m.scanItems(rows)
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
func (m *MariaDB) buildGetItemsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, queriers.ItemsTableName)
	}

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (m *MariaDB) GetItems(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := m.buildGetItemsQuery(userID, filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, err := m.scanItems(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Items: items,
	}

	return list, nil
}

// buildGetItemsForAdminQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (m *MariaDB) buildGetItemsForAdminQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	if filter == nil {
		filter = types.DefaultQueryFilter()
	}

	where := squirrel.Eq{}
	if filter != nil && filter.IncludeArchived {
		where[fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn)] = nil
	}

	builder := m.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(where).
		OrderBy(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn))

	filter.ApplyToQueryBuilder(builder, queriers.ItemsTableName)

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (m *MariaDB) GetItemsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := m.buildGetItemsForAdminQuery(filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, err := m.scanItems(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
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
func (m *MariaDB) buildGetItemsWithIDsQuery(userID uint64, limit uint8, ids []uint64) (query string, args []interface{}) {
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

	builder := m.sqlBuilder.
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
	m.logQueryBuildingError(err)

	return query, args
}

// GetItemsWithIDs fetches a list of items from the database that exist within a given set of IDs.
func (m *MariaDB) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := m.buildGetItemsWithIDsQuery(userID, limit, ids)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, err := m.scanItems(rows)
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
func (m *MariaDB) buildGetItemsWithIDsForAdminQuery(limit uint8, ids []uint64) (query string, args []interface{}) {
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

	builder := m.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):         ids,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("CASE %s.%s %s", queriers.ItemsTableName, queriers.IDColumn, whenThenStatement)).
		Limit(uint64(limit))

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetItemsWithIDsForAdmin fetches a list of items from the database that exist within a given set of IDs.
func (m *MariaDB) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := m.buildGetItemsWithIDsForAdminQuery(limit, ids)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, err := m.scanItems(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return items, nil
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (m *MariaDB) buildCreateItemQuery(input *types.Item) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
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

	m.logQueryBuildingError(err)

	return query, args
}

// CreateItem creates an item in the database.
func (m *MariaDB) CreateItem(ctx context.Context, input *types.ItemCreationInput) (*types.Item, error) {
	x := &types.Item{
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := m.buildCreateItemQuery(x)

	// create the item.
	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing item creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	m.logIDRetrievalError(err)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = m.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (m *MariaDB) buildUpdateItemQuery(input *types.Item) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.ItemsTableNameColumn, input.Name).
		Set(queriers.ItemsTableDetailsColumn, input.Details).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      input.ID,
			queriers.ItemsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (m *MariaDB) UpdateItem(ctx context.Context, input *types.Item) error {
	query, args := m.buildUpdateItemQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (m *MariaDB) buildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      itemID,
			queriers.ArchivedOnColumn:              nil,
			queriers.ItemsTableUserOwnershipColumn: userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ArchiveItem marks an item as archived in the database.
func (m *MariaDB) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	query, args := m.buildArchiveItemQuery(itemID, userID)

	res, err := m.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogItemCreationEvent saves a ItemCreationEvent in the audit log table.
func (m *MariaDB) LogItemCreationEvent(ctx context.Context, item *types.Item) {
	m.createAuditLogEntry(ctx, audit.BuildItemCreationEventEntry(item))
}

// LogItemUpdateEvent saves a ItemUpdateEvent in the audit log table.
func (m *MariaDB) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []types.FieldChangeSummary) {
	m.createAuditLogEntry(ctx, audit.BuildItemUpdateEventEntry(userID, itemID, changes))
}

// LogItemArchiveEvent saves a ItemArchiveEvent in the audit log table.
func (m *MariaDB) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildItemArchiveEventEntry(userID, itemID))
}

// buildGetAuditLogEntriesForItemQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (m *MariaDB) buildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
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
	m.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForItem fetches an audit log entry from the database.
func (m *MariaDB) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]types.AuditLogEntry, error) {
	query, args := m.buildGetAuditLogEntriesForItemQuery(itemID)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := m.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

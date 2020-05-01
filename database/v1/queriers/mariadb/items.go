package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
)

const (
	itemsTableName           = "items"
	itemsUserOwnershipColumn = "belongs_to_user"
)

var (
	itemsTableColumns = []string{
		fmt.Sprintf("%s.%s", itemsTableName, "id"),
		fmt.Sprintf("%s.%s", itemsTableName, "name"),
		fmt.Sprintf("%s.%s", itemsTableName, "details"),
		fmt.Sprintf("%s.%s", itemsTableName, "created_on"),
		fmt.Sprintf("%s.%s", itemsTableName, "updated_on"),
		fmt.Sprintf("%s.%s", itemsTableName, "archived_on"),
		fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn),
	}
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct
func (m *MariaDB) scanItem(scan database.Scanner, includeCount bool) (*models.Item, uint64, error) {
	x := &models.Item{}
	var count uint64

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.UpdatedOn,
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
func (m *MariaDB) scanItems(rows database.ResultIterator) ([]models.Item, uint64, error) {
	var (
		list  []models.Item
		count uint64
	)

	for rows.Next() {
		x, c, err := m.scanItem(rows, true)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 {
			count = c
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		m.logger.Error(closeErr, "closing database rows")
	}

	return list, count, nil
}

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists
func (m *MariaDB) buildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(fmt.Sprintf("%s.id", itemsTableName)).
		Prefix(existencePrefix).
		From(itemsTableName).
		Suffix(existenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", itemsTableName):                           itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn): userID,
		}).ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ItemExists queries the database to see if a given item belonging to a given user exists.
func (m *MariaDB) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	query, args := m.buildItemExistsQuery(itemID, userID)

	err = m.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}

	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (m *MariaDB) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", itemsTableName):                           itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn): userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the database.
func (m *MariaDB) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	query, args := m.buildGetItemQuery(itemID, userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	item, _, err := m.scanItem(row, false)
	return item, err
}

var (
	allItemsCountQueryBuilder sync.Once
	allItemsCountQuery        string
)

// buildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (m *MariaDB) buildGetAllItemsCountQuery() string {
	allItemsCountQueryBuilder.Do(func() {
		var err error

		allItemsCountQuery, _, err = m.sqlBuilder.
			Select(fmt.Sprintf(countQuery, itemsTableName)).
			From(itemsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", itemsTableName): nil,
			}).
			ToSql()
		m.logQueryBuildingError(err)
	})

	return allItemsCountQuery
}

// GetAllItemsCount will fetch the count of items from the database.
func (m *MariaDB) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = m.db.QueryRowContext(ctx, m.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (m *MariaDB) buildGetItemsQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(append(itemsTableColumns, fmt.Sprintf(countQuery, itemsTableName))...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", itemsTableName):                  nil,
			fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn): userID,
		}).
		GroupBy(fmt.Sprintf("%s.id", itemsTableName))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, itemsTableName)
	}

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (m *MariaDB) GetItems(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.ItemList, error) {
	query, args := m.buildGetItemsQuery(userID, filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for items")
	}

	items, count, err := m.scanItems(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &models.ItemList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Items: items,
	}

	return list, nil
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (m *MariaDB) buildCreateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Insert(itemsTableName).
		Columns(
			"name",
			"details",
			itemsUserOwnershipColumn,
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
func (m *MariaDB) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	x := &models.Item{
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
	x.ID = uint64(id)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = m.timeTeller.Now()

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (m *MariaDB) buildUpdateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(itemsTableName).
		Set("name", input.Name).
		Set("details", input.Details).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                     input.ID,
			itemsUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (m *MariaDB) UpdateItem(ctx context.Context, input *models.Item) error {
	query, args := m.buildUpdateItemQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (m *MariaDB) buildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(itemsTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                     itemID,
			"archived_on":            nil,
			itemsUserOwnershipColumn: userID,
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

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	itemsTableName            = "items"
	itemsTableOwnershipColumn = "belongs_to_user"
)

var (
	itemsTableColumns = []string{
		fmt.Sprintf("%s.%s", itemsTableName, "id"),
		fmt.Sprintf("%s.%s", itemsTableName, "name"),
		fmt.Sprintf("%s.%s", itemsTableName, "details"),
		fmt.Sprintf("%s.%s", itemsTableName, "created_on"),
		fmt.Sprintf("%s.%s", itemsTableName, "updated_on"),
		fmt.Sprintf("%s.%s", itemsTableName, "archived_on"),
		fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn),
	}
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct
func scanItem(scan database.Scanner) (*models.Item, error) {
	x := &models.Item{}

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	); err != nil {
		return nil, err
	}

	return x, nil
}

// scanItems takes a logger and some database rows and turns them into a slice of items
func scanItems(logger logging.Logger, rows *sql.Rows) ([]models.Item, error) {
	var list []models.Item

	for rows.Next() {
		x, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		logger.Error(closeErr, "closing database rows")
	}

	return list, nil
}

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (s *Sqlite) buildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(fmt.Sprintf("%s.id", itemsTableName)).
		Prefix(existencePrefix).
		From(itemsTableName).
		Suffix(existenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", itemsTableName):                            itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn): userID,
		}).ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ItemExists queries the database to see if a given item belonging to a given user exists
func (s *Sqlite) ItemExists(ctx context.Context, itemID, userID uint64) (bool, error) {
	var exists bool
	query, args := s.buildItemExistsQuery(itemID, userID)
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (s *Sqlite) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", itemsTableName):                            itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn): userID,
		}).ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the sqlite database
func (s *Sqlite) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	query, args := s.buildGetItemQuery(itemID, userID)
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanItem(row)
}

// buildGetItemCountQuery takes a QueryFilter and a user ID and returns a SQL query (and the relevant arguments) for
// fetching the number of items belonging to a given user that meet a given query
func (s *Sqlite) buildGetItemCountQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
		Select(fmt.Sprintf(countQuery, itemsTableName)).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", itemsTableName):                   nil,
			fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn): userID,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetItemCount will fetch the count of items from the database that meet a particular filter and belong to a particular user.
func (s *Sqlite) GetItemCount(ctx context.Context, userID uint64, filter *models.QueryFilter) (count uint64, err error) {
	query, args := s.buildGetItemCountQuery(userID, filter)
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

var (
	allItemsCountQueryBuilder sync.Once
	allItemsCountQuery        string
)

// buildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (s *Sqlite) buildGetAllItemsCountQuery() string {
	allItemsCountQueryBuilder.Do(func() {
		var err error

		allItemsCountQuery, _, err = s.sqlBuilder.
			Select(fmt.Sprintf(countQuery, itemsTableName)).
			From(itemsTableName).
			Where(squirrel.Eq{fmt.Sprintf("%s.archived_on", itemsTableName): nil}).
			ToSql()
		s.logQueryBuildingError(err)
	})

	return allItemsCountQuery
}

// GetAllItemsCount will fetch the count of items from the database
func (s *Sqlite) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = s.db.QueryRowContext(ctx, s.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (s *Sqlite) buildGetItemsQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", itemsTableName):                   nil,
			fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn): userID,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetItems fetches a list of items from the database that meet a particular filter
func (s *Sqlite) GetItems(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.ItemList, error) {
	query, args := s.buildGetItemsQuery(userID, filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for items")
	}

	list, err := scanItems(s.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	count, err := s.GetItemCount(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("fetching item count: %w", err)
	}

	x := &models.ItemList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Items: list,
	}

	return x, nil
}

// GetAllItemsForUser fetches every item belonging to a user
func (s *Sqlite) GetAllItemsForUser(ctx context.Context, userID uint64) ([]models.Item, error) {
	query, args := s.buildGetItemsQuery(userID, nil)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "fetching items for user")
	}

	list, err := scanItems(s.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("parsing database results: %w", err)
	}

	return list, nil
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (s *Sqlite) buildCreateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Insert(itemsTableName).
		Columns(
			"name",
			"details",
			itemsTableOwnershipColumn,
		).
		Values(
			input.Name,
			input.Details,
			input.BelongsToUser,
		).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// buildItemCreationTimeQuery takes an item and returns a creation query for that item and the relevant arguments
func (s *Sqlite) buildItemCreationTimeQuery(itemID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(fmt.Sprintf("%s.created_on", itemsTableName)).
		From(itemsTableName).
		Where(squirrel.Eq{fmt.Sprintf("%s.id", itemsTableName): itemID}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// CreateItem creates an item in the database
func (s *Sqlite) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	x := &models.Item{
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := s.buildCreateItemQuery(x)

	// create the item
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing item creation query: %w", err)
	}

	// fetch the last inserted ID
	id, idErr := res.LastInsertId()
	if idErr == nil {
		x.ID = uint64(id)

		query, args := s.buildItemCreationTimeQuery(x.ID)
		s.logCreationTimeRetrievalError(s.db.QueryRowContext(ctx, query, args...).Scan(&x.CreatedOn))
	}

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters
func (s *Sqlite) buildUpdateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(itemsTableName).
		Set("name", input.Name).
		Set("details", input.Details).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                      input.ID,
			itemsTableOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (s *Sqlite) UpdateItem(ctx context.Context, input *models.Item) error {
	query, args := s.buildUpdateItemQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (s *Sqlite) buildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(itemsTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                      itemID,
			"archived_on":             nil,
			itemsTableOwnershipColumn: userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveItem marks an item as archived in the database
func (s *Sqlite) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	query, args := s.buildArchiveItemQuery(itemID, userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

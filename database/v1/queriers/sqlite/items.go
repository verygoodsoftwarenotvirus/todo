package sqlite

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
	itemsTableNameColumn     = "name"
	itemsTableDetailsColumn  = "details"
	itemsUserOwnershipColumn = "belongs_to_user"
)

var (
	itemsTableColumns = []string{
		fmt.Sprintf("%s.%s", itemsTableName, idColumn),
		fmt.Sprintf("%s.%s", itemsTableName, itemsTableNameColumn),
		fmt.Sprintf("%s.%s", itemsTableName, itemsTableDetailsColumn),
		fmt.Sprintf("%s.%s", itemsTableName, createdOnColumn),
		fmt.Sprintf("%s.%s", itemsTableName, lastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", itemsTableName, archivedOnColumn),
		fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn),
	}
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct
func (s *Sqlite) scanItem(scan database.Scanner) (*models.Item, error) {
	x := &models.Item{}

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	return x, nil
}

// scanItems takes a logger and some database rows and turns them into a slice of items.
func (s *Sqlite) scanItems(rows database.ResultIterator) ([]models.Item, error) {
	var (
		list []models.Item
	)

	for rows.Next() {
		x, err := s.scanItem(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		s.logger.Error(closeErr, "closing database rows")
	}

	return list, nil
}

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists
func (s *Sqlite) buildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", itemsTableName, idColumn)).
		Prefix(existencePrefix).
		From(itemsTableName).
		Suffix(existenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", itemsTableName, idColumn):                 itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn): userID,
		}).ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ItemExists queries the database to see if a given item belonging to a given user exists.
func (s *Sqlite) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	query, args := s.buildItemExistsQuery(itemID, userID)

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}

	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (s *Sqlite) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", itemsTableName, idColumn):                 itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn): userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the database.
func (s *Sqlite) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	query, args := s.buildGetItemQuery(itemID, userID)
	row := s.db.QueryRowContext(ctx, query, args...)
	return s.scanItem(row)
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
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", itemsTableName, archivedOnColumn): nil,
			}).
			ToSql()
		s.logQueryBuildingError(err)
	})

	return allItemsCountQuery
}

// GetAllItemsCount will fetch the count of items from the database.
func (s *Sqlite) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = s.db.QueryRowContext(ctx, s.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetAllItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (s *Sqlite) buildGetAllItemsQuery(beginID, endID uint64) (query string, args []interface{}) {
	allItemsQuery, args, err := s.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", itemsTableName, idColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", itemsTableName, idColumn): endID,
		}).
		ToSql()
	s.logQueryBuildingError(err)

	return allItemsQuery, args
}

// GetAllItems fetches all items from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks,
func (s *Sqlite) GetAllItems(ctx context.Context, resultChannel chan []models.Item) error {
	count, err := s.GetAllItemsCount(ctx)
	if err != nil {
		return err
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin uint64, end uint64) {
			query, args := s.buildGetAllItemsQuery(begin, end)
			logger := s.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := s.db.Query(query, args...)
			if err == sql.ErrNoRows {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			items, err := s.scanItems(rows)
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
func (s *Sqlite) buildGetItemsQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", itemsTableName, archivedOnColumn):         nil,
			fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn): userID,
		}).
		OrderBy(fmt.Sprintf("%s.%s", itemsTableName, idColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, itemsTableName)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (s *Sqlite) GetItems(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.ItemList, error) {
	query, args := s.buildGetItemsQuery(userID, filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for items")
	}

	items, err := s.scanItems(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &models.ItemList{
		Pagination: models.Pagination{
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
func (s *Sqlite) buildGetItemsWithIDsQuery(userID uint64, limit uint8, ids []uint64) (query string, args []interface{}) {
	var err error

	var whenThenStatement string
	for i, id := range ids {
		if i != 0 {
			whenThenStatement += " "
		}
		whenThenStatement += fmt.Sprintf("WHEN %d THEN %d", id, i)
	}
	whenThenStatement += " END"

	builder := s.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", itemsTableName, idColumn):                 ids,
			fmt.Sprintf("%s.%s", itemsTableName, archivedOnColumn):         nil,
			fmt.Sprintf("%s.%s", itemsTableName, itemsUserOwnershipColumn): userID,
		}).
		OrderBy(fmt.Sprintf("CASE %s.%s %s", itemsTableName, idColumn, whenThenStatement)).
		Limit(uint64(limit))

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetItemsWithIDs fetches a list of items from the database that exist within a given set of IDs.
func (s *Sqlite) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]models.Item, error) {
	if limit == 0 {
		limit = uint8(models.DefaultLimit)
	}

	query, args := s.buildGetItemsWithIDsQuery(userID, limit, ids)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for items")
	}

	items, err := s.scanItems(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return items, nil
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (s *Sqlite) buildCreateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Insert(itemsTableName).
		Columns(
			itemsTableNameColumn,
			itemsTableDetailsColumn,
			itemsUserOwnershipColumn,
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

// CreateItem creates an item in the database.
func (s *Sqlite) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	x := &models.Item{
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := s.buildCreateItemQuery(x)

	// create the item.
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing item creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	s.logIDRetrievalError(err)
	x.ID = uint64(id)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = s.timeTeller.Now()

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (s *Sqlite) buildUpdateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(itemsTableName).
		Set(itemsTableNameColumn, input.Name).
		Set(itemsTableDetailsColumn, input.Details).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn:                 input.ID,
			itemsUserOwnershipColumn: input.BelongsToUser,
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
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(archivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn:                 itemID,
			archivedOnColumn:         nil,
			itemsUserOwnershipColumn: userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveItem marks an item as archived in the database.
func (s *Sqlite) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	query, args := s.buildArchiveItemQuery(itemID, userID)

	res, err := s.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

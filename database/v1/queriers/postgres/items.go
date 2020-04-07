package postgres

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
func scanItem(scan database.Scanner, includeCount bool) (*models.Item, uint64, error) {
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

// scanItems takes a logger and some database rows and turns them into a slice of items
func scanItems(logger logging.Logger, rows *sql.Rows) ([]models.Item, uint64, error) {
	var (
		list  []models.Item
		count uint64
	)

	for rows.Next() {
		x, c, err := scanItem(rows, true)
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
		logger.Error(closeErr, "closing database rows")
	}

	return list, count, nil
}

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (p *Postgres) buildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(fmt.Sprintf("%s.id", itemsTableName)).
		Prefix(existencePrefix).
		From(itemsTableName).
		Suffix(existenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", itemsTableName):                            itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn): userID,
		}).ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// ItemExists queries the database to see if a given item belonging to a given user exists
func (p *Postgres) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	query, args := p.buildItemExistsQuery(itemID, userID)
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (p *Postgres) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", itemsTableName):                            itemID,
			fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn): userID,
		}).ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the postgres database
func (p *Postgres) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	query, args := p.buildGetItemQuery(itemID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	x, _, err := scanItem(row, false)
	return x, err
}

var (
	allItemsCountQueryBuilder sync.Once
	allItemsCountQuery        string
)

// buildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (p *Postgres) buildGetAllItemsCountQuery() string {
	allItemsCountQueryBuilder.Do(func() {
		var err error

		allItemsCountQuery, _, err = p.sqlBuilder.
			Select(fmt.Sprintf(countQuery, itemsTableName)).
			From(itemsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", itemsTableName): nil,
			}).
			ToSql()
		p.logQueryBuildingError(err)
	})

	return allItemsCountQuery
}

// GetAllItemsCount will fetch the count of items from the database
func (p *Postgres) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = p.db.QueryRowContext(ctx, p.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (p *Postgres) buildGetItemsQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := p.sqlBuilder.
		Select(append(itemsTableColumns, fmt.Sprintf(countQuery, itemsTableName))...).
		From(itemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", itemsTableName):                   nil,
			fmt.Sprintf("%s.%s", itemsTableName, itemsTableOwnershipColumn): userID,
		}).
		GroupBy(fmt.Sprintf("%s.id", itemsTableName))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, itemsTableName)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetItems fetches a list of items from the database that meet a particular filter
func (p *Postgres) GetItems(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.ItemList, error) {
	query, args := p.buildGetItemsQuery(userID, filter)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for items")
	}

	items, count, err := scanItems(p.logger, rows)
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
func (p *Postgres) buildCreateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
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
		Suffix("RETURNING id, created_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// CreateItem creates an item in the database
func (p *Postgres) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	x := &models.Item{
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := p.buildCreateItemQuery(x)

	// create the item
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, fmt.Errorf("error executing item creation query: %w", err)
	}

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters
func (p *Postgres) buildUpdateItemQuery(input *models.Item) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(itemsTableName).
		Set("name", input.Name).
		Set("details", input.Details).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                      input.ID,
			itemsTableOwnershipColumn: input.BelongsToUser,
		}).
		Suffix("RETURNING updated_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (p *Postgres) UpdateItem(ctx context.Context, input *models.Item) error {
	query, args := p.buildUpdateItemQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.UpdatedOn)
}

// buildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (p *Postgres) buildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(itemsTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                      itemID,
			"archived_on":             nil,
			itemsTableOwnershipColumn: userID,
		}).
		Suffix("RETURNING archived_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// ArchiveItem marks an item as archived in the database
func (p *Postgres) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	query, args := p.buildArchiveItemQuery(itemID, userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

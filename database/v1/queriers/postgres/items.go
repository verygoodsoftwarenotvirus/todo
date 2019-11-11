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
	itemsTableName = "items"
)

var (
	itemsTableColumns = []string{
		"id",
		"name",
		"details",
		"created_on",
		"updated_on",
		"archived_on",
		"belongs_to",
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
		&x.BelongsTo,
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

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (p *Postgres) buildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			"id":         itemID,
			"belongs_to": userID,
		}).ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetItem fetches an item from the postgres database
func (p *Postgres) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	query, args := p.buildGetItemQuery(itemID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)
	return scanItem(row)
}

// buildGetItemCountQuery takes a QueryFilter and a user ID and returns a SQL query (and the relevant arguments) for
// fetching the number of items belonging to a given user that meet a given query
func (p *Postgres) buildGetItemCountQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
		Select(CountQuery).
		From(itemsTableName).
		Where(squirrel.Eq{
			"archived_on": nil,
			"belongs_to":  userID,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetItemCount will fetch the count of items from the database that meet a particular filter and belong to a particular user.
func (p *Postgres) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	query, args := p.buildGetItemCountQuery(filter, userID)
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
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
			Select(CountQuery).
			From(itemsTableName).
			Where(squirrel.Eq{"archived_on": nil}).
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
func (p *Postgres) buildGetItemsQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			"archived_on": nil,
			"belongs_to":  userID,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetItems fetches a list of items from the database that meet a particular filter
func (p *Postgres) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	query, args := p.buildGetItemsQuery(filter, userID)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying database for items")
	}

	list, err := scanItems(p.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	count, err := p.GetItemCount(ctx, filter, userID)
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
func (p *Postgres) GetAllItemsForUser(ctx context.Context, userID uint64) ([]models.Item, error) {
	query, args := p.buildGetItemsQuery(nil, userID)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "fetching items for user")
	}

	list, err := scanItems(p.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("parsing database results: %w", err)
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
			"belongs_to",
		).
		Values(
			input.Name,
			input.Details,
			input.BelongsTo,
		).
		Suffix("RETURNING id, created_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// CreateItem creates an item in the database
func (p *Postgres) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	x := &models.Item{
		Name:      input.Name,
		Details:   input.Details,
		BelongsTo: input.BelongsTo,
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
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":         input.ID,
			"belongs_to": input.BelongsTo,
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
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":          itemID,
			"archived_on": nil,
			"belongs_to":  userID,
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

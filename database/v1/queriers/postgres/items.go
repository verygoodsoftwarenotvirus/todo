package postgres

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
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
		"completed_on",
		"belongs_to",
	}
)

func scanItem(scan database.Scanner) (*models.Item, error) {
	var (
		x = &models.Item{}
	)

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.CompletedOn,
		&x.BelongsTo,
	); err != nil {
		return nil, err
	}

	return x, nil
}

func (p *Postgres) scanItems(rows *sql.Rows) ([]models.Item, error) {
	var list []models.Item

	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

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

	return list, nil
}

const getItemQuery = `
	SELECT
		id,
		name,
		details,
		created_on,
		updated_on,
		completed_on,
		belongs_to
	FROM
		items
	WHERE
		id = $1
		AND belongs_to = $2
`

// GetItem fetches an item from the postgres database
func (p *Postgres) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	row := p.database.QueryRowContext(ctx, getItemQuery, itemID, userID)
	i, err := scanItem(row)
	return i, err
}

// GetItemCount will fetch the count of items from the postgres database that meet a particular filter and belong to a particular user.
func (p *Postgres) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	builder := p.sqlBuilder.
		Select("COUNT(*)").
		From(itemsTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":   userID,
			"completed_on": nil,
		}))

	builder = filter.ApplyToQueryBuilder(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "generating query")
	}

	err = p.database.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

const getAllItemsCountQuery = `
	SELECT
		COUNT(*)
	FROM
		items
	WHERE
		completed_on IS NULL
`

// GetAllItemsCount will fetch the count of items from the postgres database that meet a particular filter
func (p *Postgres) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = p.database.QueryRowContext(ctx, getAllItemsCountQuery).Scan(&count)
	return
}

// GetItems fetches a list of items from the postgres database that meet a particular filter
func (p *Postgres) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	builder := p.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":   userID,
			"completed_on": nil,
		}))

	builder = filter.ApplyToQueryBuilder(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "generating query")
	}

	rows, err := p.database.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	list, err := p.scanItems(rows)
	if err != nil {
		return nil, err
	}

	count, err := p.GetItemCount(ctx, filter, userID)
	if err != nil {
		return nil, err
	}

	x := &models.ItemList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			TotalCount: count,
			Limit:      filter.Limit,
		},
		Items: list,
	}

	return x, err
}

const createItemQuery = `
	INSERT INTO items
	(
		name,
		details,
		belongs_to
	)
	VALUES
	(
		$1, $2, $3
	)
	RETURNING
		id,
		created_on
`

// CreateItem creates an item in a postgres database
func (p *Postgres) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	i := &models.Item{
		Name:      input.Name,
		Details:   input.Details,
		BelongsTo: input.BelongsTo,
	}

	// create the item
	if err := p.database.
		QueryRow(createItemQuery, input.Name, input.Details, input.BelongsTo).
		Scan(&i.ID, &i.CreatedOn); err != nil {
		return nil, errors.Wrap(err, "error executing item creation query")
	}

	return i, nil
}

const updateItemQuery = `
	UPDATE items SET
		name = $1,
		details = $2,
		updated_on = extract(epoch FROM NOW())
	WHERE
		id = $3
		AND belongs_to = $4
	RETURNING
		updated_on
`

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (p *Postgres) UpdateItem(ctx context.Context, input *models.Item) error {
	// update the item
	err := p.database.
		QueryRowContext(
			ctx,
			updateItemQuery,
			input.Name,
			input.Details,
			input.ID,
			input.BelongsTo,
		).Scan(&input.UpdatedOn)
	return err
}

const archiveItemQuery = `
	UPDATE items SET
		updated_on = extract(epoch FROM NOW()),
		completed_on = extract(epoch FROM NOW())
	WHERE
		id = $1
		AND completed_on IS NULL
		AND belongs_to = $2
	RETURNING
		completed_on
`

// DeleteItem deletes an item from the database by its ID
func (p *Postgres) DeleteItem(ctx context.Context, itemID uint64, userID uint64) error {
	_, err := p.database.ExecContext(ctx, archiveItemQuery, itemID, userID)
	return err
}

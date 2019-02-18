package postgres

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

func (p Postgres) scanItem(scan Scannable) (*models.Item, error) {
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
	i, err := p.scanItem(row)
	return i, err
}

const getItemCountQuery = `
	SELECT
		COUNT(*)
	FROM
		items
	WHERE
		completed_on IS NULL
` // FINISHME: finish adding filters to this query

// GetItemCount fetches the count of items from the postgres database that meet a particular filter
func (p *Postgres) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	err = p.database.QueryRowContext(ctx, getItemCountQuery).Scan(&count)
	return
}

const getItemsQuery = `
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
		completed_on IS NULL
	LIMIT $1
	OFFSET $2
` // FINISHME: finish adding filters to this query

// GetItems fetches a list of items from the postgres database that meet a particular filter
func (p *Postgres) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	if filter == nil {
		p.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}

	var list []models.Item
	rows, err := p.database.QueryContext(ctx, getItemsQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		item, ierr := p.scanItem(rows)
		if ierr != nil {
			return nil, ierr
		}
		list = append(list, *item)
	}

	if err = rows.Err(); err != nil {
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
	RETURNING
		updated_on
`

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (p *Postgres) UpdateItem(ctx context.Context, input *models.Item) error {
	// update the item
	err := p.database.QueryRowContext(ctx, updateItemQuery, input.Name, input.Details, input.ID).Scan(&input.UpdatedOn)
	return err
}

const archiveItemQuery = `
	UPDATE items SET
		updated_on = extract(epoch FROM NOW()),
		completed_on = extract(epoch FROM NOW())
	WHERE
		id = $1
		AND completed_on IS NULL
	RETURNING
		completed_on
`

// DeleteItem deletes an item from the database by its ID
func (p *Postgres) DeleteItem(ctx context.Context, id uint64, userID uint64) error {
	_, err := p.database.ExecContext(ctx, archiveItemQuery, id)
	return err
}

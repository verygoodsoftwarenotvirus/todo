package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

func scanItem(scan Scannable) (*models.Item, error) {
	x := &models.Item{}
	err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.CompletedOn,
		&x.BelongsTo,
	)
	if err != nil {
		return nil, err
	}
	return x, nil
}

func (s *Sqlite) scanItems(rows *sql.Rows) ([]models.Item, error) {
	var list []models.Item

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
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
		id = ?
		AND belongs_to = ?
`

// GetItem fetches an item from the sqlite database
func (s *Sqlite) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	row := s.database.QueryRowContext(ctx, getItemQuery, itemID, userID)
	i, err := scanItem(row)
	return i, err
}

const getItemCountQuery = `
	SELECT
		COUNT(*)
	FROM
		items
	WHERE
		completed_on IS NULL
`

// GetItemCount fetches the count of items from the sqlite database that meet a particular filter
func (s *Sqlite) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	var count uint64
	err := s.database.QueryRowContext(ctx, getItemCountQuery).Scan(&count)
	return count, err
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
	LIMIT ?
	OFFSET ?
`

// GetItems fetches a list of items from the sqlite database that meet a particular filter
func (s *Sqlite) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	rows, err := s.database.QueryContext(ctx, getItemsQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, errors.Wrap(err, "querying database for items")
	}

	list, err := s.scanItems(rows)
	if err != nil {
		return nil, errors.Wrap(err, "scanning items")
	}

	count, err := s.GetItemCount(ctx, filter, userID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching item count")
	}

	x := &models.ItemList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			TotalCount: count,
			Limit:      filter.Limit,
		},
		Items: list,
	}

	return x, nil
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
		?, ?, ?
	)
`

// CreateItem creates an item in a sqlite database
func (s *Sqlite) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	// create the item
	res, err := s.database.ExecContext(ctx, createItemQuery, input.Name, input.Details, input.BelongsTo)
	if err != nil {
		return nil, errors.Wrap(err, "error executing item creation query")
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching last inserted item ID")
	}

	// fetch full updated item
	i, err := s.GetItem(ctx, uint64(id), input.BelongsTo)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching newly created item")
	}

	return i, nil
}

const updateItemQuery = `
	UPDATE items SET
		name = ?,
		details = ?,
		updated_on = (strftime('%s','now'))
	WHERE
		id = ?
`

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (s *Sqlite) UpdateItem(ctx context.Context, input *models.Item) error {
	if _, err := s.database.ExecContext(ctx, updateItemQuery, input.Name, input.Details, input.ID); err != nil {
		return errors.Wrap(err, "updating item")
	}

	return nil
}

const archiveItemQuery = `
	UPDATE items SET
		updated_on = (strftime('%s','now')),
		completed_on = (strftime('%s','now'))
	WHERE
		id = ?
		AND completed_on IS NULL
`

// DeleteItem deletes an item from the database by its ID
func (s *Sqlite) DeleteItem(ctx context.Context, id uint64, userID uint64) error {
	_, err := s.database.ExecContext(ctx, archiveItemQuery, id)
	return err
}

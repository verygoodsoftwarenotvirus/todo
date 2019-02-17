package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	getItemQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on, belongs_to
		FROM
			items
		WHERE
			id = ? AND belongs_to = ?
	`
	getItemCountQuery = `
		SELECT
			COUNT(*)
		FROM
			items
		WHERE completed_on IS NULL
	`
	getItemsQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on, belongs_to
		FROM
			items
		WHERE
			completed_on IS NULL
		LIMIT ?
		OFFSET ?
	`
	createItemQuery = `
		INSERT INTO items
		(
			name, details, belongs_to
		)
		VALUES
		(
			?, ?, ?
		)
	`
	updateItemQuery = `
		UPDATE items SET
			name = ?,
			details = ?,
			updated_on = (strftime('%s','now'))
		WHERE id = ?
	`
	archiveItemQuery = `
		UPDATE items SET
			updated_on = (strftime('%s','now')),
			completed_on = (strftime('%s','now'))
		WHERE id = ? AND completed_on IS NULL
	`
)

func scanItem(scan database.Scannable) (*models.Item, error) {
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

// GetItem fetches an item from the sqlite database
func (s *Sqlite) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetItem")
	defer span.Finish()

	row := s.database.QueryRow(getItemQuery, itemID, userID)
	i, err := scanItem(row)
	return i, err
}

// GetItemCount fetches the count of items from the sqlite database that meet a particular filter
func (s *Sqlite) GetItemCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetItemCount")
	defer span.Finish()

	var count uint64
	err := s.database.QueryRow(getItemCountQuery).Scan(&count)
	return count, err
}

// GetItems fetches a list of items from the sqlite database that meet a particular filter
func (s *Sqlite) GetItems(ctx context.Context, filter *models.QueryFilter) (*models.ItemList, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetItems")
	defer span.Finish()

	filter = s.prepareFilter(filter, span)

	rows, err := s.database.Query(getItemsQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, errors.Wrap(err, "querying database for items")
	}

	list, err := s.scanItems(rows)
	if err != nil {
		return nil, errors.Wrap(err, "scanning items")
	}

	count, err := s.GetItemCount(ctx, filter)
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

// CreateItem creates an item in a sqlite database
func (s *Sqlite) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "CreateItem")
	defer span.Finish()

	logger := s.logger.WithValue("belongs_to", input.BelongsTo)

	// create the item
	res, err := s.database.Exec(createItemQuery, input.Name, input.Details, input.BelongsTo)
	if err != nil {
		logger.Error(err, "error executing item creation query")
		return nil, err
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error(err, "error fetching last inserted item ID")
		return nil, err
	}

	logger = logger.WithValue("item_id", id)

	// fetch full updated item
	i, err := s.GetItem(ctx, uint64(id), input.BelongsTo)
	if err != nil {
		logger.Error(err, "error fetching newly created item")
		return nil, err
	}

	logger.Debug("returning from CreateItem")
	return i, nil
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (s *Sqlite) UpdateItem(ctx context.Context, input *models.Item) error {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "UpdateItem")
	defer span.Finish()

	// update the item
	if _, err := s.database.Exec(updateItemQuery, input.Name, input.Details, input.ID); err != nil {
		return errors.Wrap(err, "updating item")
	}

	return nil
}

// DeleteItem deletes an item from the database by its ID
func (s *Sqlite) DeleteItem(ctx context.Context, id uint64) error {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "DeleteItem")
	defer span.Finish()

	_, err := s.database.Exec(archiveItemQuery, id)
	return err
}

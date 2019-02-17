package postgres

import (
	"context"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	getItemQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on, belongs_to
		FROM
			items
		WHERE
			id = $1 AND belongs_to = $2
	`
	getItemCountQuery = `
		SELECT
			COUNT(*)
		FROM
			items
		WHERE 
			completed_on IS NULL 
			AND belongs_to = $1
	`  // FINISHME: finish adding filters to this query
	getItemsQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on, belongs_to
		FROM
			items
		WHERE
			completed_on IS NULL 
			AND belongs_to = $1
		LIMIT $2
		OFFSET $3
	`  // FINISHME: finish adding filters to this query
	createItemQuery = `
		INSERT INTO items
		(
			name, details, belongs_to
		)
		VALUES
		(
			$1, $2, $3
		)
		RETURNING
			id, created_on
	`
	updateItemQuery = `
		UPDATE items SET
			name = $1,
			details = $2,
			updated_on = extract(epoch FROM NOW())
		WHERE 
			id = $3
		RETURNING
			updated_on
	`
	archiveItemQuery = `
		UPDATE items SET
			updated_on = extract(epoch FROM NOW()),
			completed_on = extract(epoch FROM NOW())
		WHERE 
			id = $1 
			AND belongs_to = $2
			AND completed_on IS NULL
		RETURNING
			completed_on
	`
)

func (p Postgres) scanItem(scan database.Scannable) (*models.Item, error) {
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

// GetItem fetches an item from the postgres database
func (p *Postgres) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetItem")
	span.SetTag("item_id", itemID)
	span.SetTag("user_id", userID)
	defer span.Finish()

	p.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("GetItem called")

	row := p.database.QueryRow(getItemQuery, itemID, userID)
	i, err := p.scanItem(row)
	return i, err
}

// GetItemCount fetches the count of items from the postgres database that meet a particular filter
func (p *Postgres) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetItemCount")
	span.SetTag("user_id", userID)
	defer span.Finish()

	p.logger.WithValue("filter", filter).Debug("GetItemCount called")

	err = p.database.QueryRow(getItemCountQuery, userID).Scan(&count)
	return
}

// GetItems fetches a list of items from the postgres database that meet a particular filter
func (p *Postgres) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetItems")
	span.SetTag("user_id", userID)
	defer span.Finish()

	p.logger.WithValue("filter", filter).Debug("GetItems called")

	if filter == nil {
		p.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.Page = uint64(math.Max(1, float64(filter.Page)))
	queryPage := uint(filter.Limit * (filter.Page - 1))

	var list []models.Item

	rows, err := p.database.Query(getItemsQuery, userID, filter.Limit, queryPage)
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

// CreateItem creates an item in a postgres database
func (p *Postgres) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "CreateItem")
	defer span.Finish()

	p.logger.WithValue("input", input).Debug("CreateItem called")

	i := &models.Item{
		Name:      input.Name,
		Details:   input.Details,
		BelongsTo: input.BelongsTo,
	}

	// create the item
	err := p.database.
		QueryRow(createItemQuery, input.Name, input.Details, input.BelongsTo).
		Scan(&i.ID, &i.CreatedOn)
	if err != nil {
		p.logger.Error(err, "error executing item creation query")
		return nil, err
	}

	return i, nil
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (p *Postgres) UpdateItem(ctx context.Context, input *models.Item) error {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "UpdateItem")
	span.SetTag("item_id", input.ID)
	defer span.Finish()

	p.logger.WithValue("input", input).Debug("UpdateItem called")

	// update the item
	err := p.database.QueryRow(updateItemQuery, input.Name, input.Details, input.ID).Scan(&input.UpdatedOn)
	return err
}

// DeleteItem deletes an item from the database by its ID
func (p *Postgres) DeleteItem(ctx context.Context, itemID uint64, userID uint64) error {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "DeleteItem")
	span.SetTag("item_id", itemID)
	span.SetTag("user_id", userID)
	defer span.Finish()

	p.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("DeleteItem called")

	_, err := p.database.Exec(archiveItemQuery, itemID, userID)
	return err
}

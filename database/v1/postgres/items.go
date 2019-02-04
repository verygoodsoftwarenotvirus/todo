package postgres

import (
	"context"
	"math"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
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
		WHERE completed_on IS NULL
	`
	getItemsQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on, belongs_to
		FROM
			items
		WHERE
			completed_on IS NULL
		LIMIT $1
		OFFSET $2
	`
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
		WHERE id = $3
		RETURNING
			updated_on
	`
	archiveItemQuery = `
		UPDATE items SET
			updated_on = to_timestamp(extract(epoch FROM NOW())),
			completed_on = to_timestamp(extract(epoch FROM NOW()))
		WHERE id = $1 AND completed_on IS NULL
		RETURNING
			completed_on
	`
)

func scanItem(scan database.Scannable) (*models.Item, error) {
	var (
		x = &models.Item{}

		co time.Time
		uo *time.Time
		ao *time.Time
	)

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.Details,
		&co,
		&uo,
		&ao,
		&x.BelongsTo,
	); err != nil {
		return nil, err
	}

	x.CreatedOn = timeToUInt64(co)
	if uo != nil {
		x.UpdatedOn = timeToPUInt64(uo)
	}
	if ao != nil {
		x.CompletedOn = timeToPUInt64(ao)
	}

	return x, nil
}

// GetItem fetches an item from the postgres database
func (p *Postgres) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	p.logger.WithFields(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debugln("GetItem called")

	row := p.database.QueryRow(getItemQuery, itemID, userID)
	i, err := scanItem(row)
	return i, err
}

// GetItemCount fetches the count of items from the postgres database that meet a particular filter
func (p *Postgres) GetItemCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	p.logger.WithField("filter", filter).Debugln("GetItemCount called")

	err = p.database.QueryRow(getItemCountQuery).Scan(&count)
	return
}

// GetItems fetches a list of items from the postgres database that meet a particular filter
func (p *Postgres) GetItems(ctx context.Context, filter *models.QueryFilter) (*models.ItemList, error) {
	p.logger.WithField("filter", filter).Debugln("GetItems called")

	if filter == nil {
		p.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.Page = uint64(math.Max(1, float64(filter.Page)))
	queryPage := uint(filter.Limit * (filter.Page - 1))

	list := []models.Item{}

	rows, err := p.database.Query(getItemsQuery, filter.Limit, queryPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

	count, err := p.GetItemCount(ctx, filter)
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
	p.logger.WithField("input", input).Debugln("CreateItem called")

	i := &models.Item{
		Name:      input.Name,
		Details:   input.Details,
		BelongsTo: input.BelongsTo,
	}

	// create the item
	var t time.Time
	if err := p.database.
		QueryRow(createItemQuery, input.Name, input.Details, input.BelongsTo).
		Scan(&i.ID, &t); err != nil {
		p.logger.Errorf("error executing item creation query: %v", err)
		return nil, err
	}
	i.CreatedOn = timeToUInt64(t)

	return i, nil
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (p *Postgres) UpdateItem(ctx context.Context, input *models.Item) error {
	p.logger.WithField("input", input).Debugln("UpdateItem called")

	// update the item
	var t *time.Time
	err := p.database.QueryRow(updateItemQuery, input.Name, input.Details, input.ID).Scan(&t)
	uo := uint64(t.Unix())
	input.UpdatedOn = &uo
	return err
}

// DeleteItem deletes an item from the database by its ID
func (p *Postgres) DeleteItem(ctx context.Context, id uint64) error {
	p.logger.WithField("id", id).Debugln("DeleteItem called")

	_, err := p.database.Exec(archiveItemQuery, id)
	return err
}

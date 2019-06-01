package dbclient

import (
	"context"
	"database/sql"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
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

var _ models.ItemDataManager = (*Client)(nil)

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

func scanItems(logger logging.Logger, rows *sql.Rows) ([]models.Item, error) {
	var list []models.Item

	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(err, "closing rows")
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
func (c *Client) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "GetItem")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(itemID, 10)))
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("GetItem called")

	row := c.db.QueryRowContext(ctx, getItemQuery, itemID, userID)
	i, err := scanItem(row)
	return i, err
}

// GetItemCount fetches the count of items from the postgres database that meet a particular filter
func (c *Client) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetItemCount")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	}).Debug("GetItemCount called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	builder := c.sqlBuilder.
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

	return count, c.db.QueryRowContext(ctx, query, args...).Scan(&count)
}

const getAllItemsCountQuery = `
	SELECT
		COUNT(*)
	FROM
		items
	WHERE
		completed_on IS NULL
`

// GetAllItemsCount fetches the count of items from the postgres database that meet a particular filter
func (c *Client) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetAllItemsCount")
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	err = c.db.QueryRowContext(ctx, getAllItemsCountQuery).Scan(&count)
	return count, err
}

// GetItems fetches a list of items from the postgres database that meet a particular filter
func (c *Client) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	ctx, span := trace.StartSpan(ctx, "GetItems")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	}).Debug("GetItems called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	builder := filter.ApplyToQueryBuilder(c.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":   userID,
			"completed_on": nil,
		})))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "generating query")
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			c.logger.Error(err, "closing rows")
		}
	}()

	list, err := scanItems(c.logger, rows)
	if err != nil {
		return nil, err
	}

	count, err := c.GetItemCount(ctx, filter, userID)
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
func (c *Client) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "CreateItem")
	defer span.End()

	c.logger.WithValue("input", input).Debug("CreateItem called")

	i := &models.Item{
		Name:      input.Name,
		Details:   input.Details,
		BelongsTo: input.BelongsTo,
	}

	// create the item
	if err := c.db.
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
func (c *Client) UpdateItem(ctx context.Context, input *models.Item) error {
	ctx, span := trace.StartSpan(ctx, "UpdateItem")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(input.ID, 10)))

	c.logger.WithValue("input", input).Debug("UpdateItem called")

	// update the item
	err := c.db.
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
func (c *Client) DeleteItem(ctx context.Context, itemID uint64, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "DeleteItem")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(itemID, 10)))
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("DeleteItem called")

	_, err := c.db.ExecContext(ctx, archiveItemQuery, itemID, userID)
	return err
}

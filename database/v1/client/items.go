package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var _ models.ItemHandler = (*Client)(nil)

// GetItem fetches an item from the postgres database
func (c *Client) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetItem")
	span.SetTag("item_id", itemID)
	span.SetTag("user_id", userID)
	defer span.Finish()

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("GetItem called")

	return c.database.GetItem(ctx, itemID, userID)
}

// GetItemCount fetches the count of items from the postgres database that meet a particular filter
func (c *Client) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetItemCount")
	span.SetTag("user_id", userID)
	defer span.Finish()

	c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	}).Debug("GetItemCount called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	return c.database.GetItemCount(ctx, filter, userID)
}

// GetItems fetches a list of items from the postgres database that meet a particular filter
func (c *Client) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetItems")
	span.SetTag("user_id", userID)
	defer span.Finish()

	c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	}).Debug("GetItemCount called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	return c.database.GetItems(ctx, filter, userID)
}

// CreateItem creates an item in a postgres database
func (c *Client) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateItem")
	defer span.Finish()

	c.logger.WithValue("input", input).Debug("CreateItem called")

	return c.database.CreateItem(ctx, input)
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (c *Client) UpdateItem(ctx context.Context, input *models.Item) error {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "UpdateItem")
	span.SetTag("item_id", input.ID)
	defer span.Finish()

	c.logger.WithValue("input", input).Debug("UpdateItem called")

	return c.database.UpdateItem(ctx, input)
}

// DeleteItem deletes an item from the database by its ID
func (c *Client) DeleteItem(ctx context.Context, itemID uint64, userID uint64) error {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "DeleteItem")
	span.SetTag("item_id", itemID)
	span.SetTag("user_id", userID)
	defer span.Finish()

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("DeleteItem called")

	return c.database.DeleteItem(ctx, itemID, userID)
}

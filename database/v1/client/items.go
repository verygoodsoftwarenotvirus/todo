package dbclient

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.ItemDataManager = (*Client)(nil)

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

	return c.database.GetItem(ctx, itemID, userID)
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

	return c.database.GetItemCount(ctx, filter, userID)
}

// GetAllItemsCount fetches the count of items from the postgres database that meet a particular filter
func (c *Client) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetAllItemsCount")
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	return c.database.GetAllItemsCount(ctx)
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

	return c.database.GetItems(ctx, filter, userID)
}

// CreateItem creates an item in a postgres database
func (c *Client) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "CreateItem")
	defer span.End()

	c.logger.WithValue("input", input).Debug("CreateItem called")

	return c.database.CreateItem(ctx, input)
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (c *Client) UpdateItem(ctx context.Context, input *models.Item) error {
	ctx, span := trace.StartSpan(ctx, "UpdateItem")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(input.ID, 10)))

	c.logger.WithValue("input", input).Debug("UpdateItem called")

	return c.database.UpdateItem(ctx, input)
}

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

	return c.database.DeleteItem(ctx, itemID, userID)
}

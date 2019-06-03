package dbclient

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.ItemDataManager = (*Client)(nil)

// GetItem fetches an item from the postgres querier
func (c *Client) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "GetItem")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(itemID, 10)))
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("GetItem called")

	return c.querier.GetItem(ctx, itemID, userID)
}

// GetItemCount fetches the count of items from the postgres querier that meet a particular filter
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
		filter = models.DefaultQueryFilter()
	}
	filter.SetPage(filter.Page)

	return c.querier.GetItemCount(ctx, filter, userID)
}

// GetAllItemsCount fetches the count of items from the postgres querier that meet a particular filter
func (c *Client) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetAllItemsCount")
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	return c.querier.GetAllItemsCount(ctx)
}

// GetItems fetches a list of items from the postgres querier that meet a particular filter
func (c *Client) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	ctx, span := trace.StartSpan(ctx, "GetItems")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	logger := c.logger.WithValues(map[string]interface{}{
		"filter":  filter,
		"user_id": userID,
	})
	logger.Debug("GetItems called")

	if filter == nil {
		logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter()
	}
	filter.SetPage(filter.Page)

	itemList, err := c.querier.GetItems(ctx, filter, userID)

	logger.WithValues(map[string]interface{}{
		"item_count": itemList.TotalCount,
		"item_list":  itemList.Items,
		"err":        err,
	}).Debug("returning from GetItems")

	return itemList, err
}

// CreateItem creates an item in a postgres querier
func (c *Client) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "CreateItem")
	defer span.End()

	c.logger.WithValue("input", input).Debug("CreateItem called")

	return c.querier.CreateItem(ctx, input)
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (c *Client) UpdateItem(ctx context.Context, input *models.Item) error {
	ctx, span := trace.StartSpan(ctx, "UpdateItem")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(input.ID, 10)))

	c.logger.WithValue("input", input).Debug("UpdateItem called")

	return c.querier.UpdateItem(ctx, input)
}

// DeleteItem deletes an item from the querier by its ID
func (c *Client) DeleteItem(ctx context.Context, itemID uint64, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "DeleteItem")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(itemID, 10)))
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("DeleteItem called")

	return c.querier.DeleteItem(ctx, itemID, userID)
}

package dbclient

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.ItemDataManager = (*Client)(nil)

func attachItemIDToSpan(span *trace.Span, itemID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("item_id", strconv.FormatUint(itemID, 10)))
	}
}

// GetItem fetches an item from the postgres querier
func (c *Client) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "GetItem")
	defer span.End()

	attachUserIDToSpan(span, userID)
	attachItemIDToSpan(span, itemID)

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

	if filter == nil {
		filter = models.DefaultQueryFilter()
	}

	attachUserIDToSpan(span, userID)
	attachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetItemCount called")

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

	if filter == nil {
		filter = models.DefaultQueryFilter()
	}

	attachUserIDToSpan(span, userID)
	attachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetItems called")

	itemList, err := c.querier.GetItems(ctx, filter, userID)

	return itemList, err
}

// GetAllItemsForUser fetches a list of items from the postgres querier that meet a particular filter
func (c *Client) GetAllItemsForUser(ctx context.Context, userID uint64) ([]models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllItemsForUser")
	defer span.End()

	attachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("GetAllItemsForUser called")

	itemList, err := c.querier.GetAllItemsForUser(ctx, userID)

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

	attachItemIDToSpan(span, input.ID)
	c.logger.WithValue("item_id", input.ID).Debug("UpdateItem called")

	return c.querier.UpdateItem(ctx, input)
}

// DeleteItem deletes an item from the querier by its ID
func (c *Client) DeleteItem(ctx context.Context, itemID, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "DeleteItem")
	defer span.End()

	attachUserIDToSpan(span, userID)
	attachItemIDToSpan(span, itemID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("DeleteItem called")

	return c.querier.DeleteItem(ctx, itemID, userID)
}

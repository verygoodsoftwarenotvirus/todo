package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.ItemDataManager = (*Client)(nil)

// ItemExists fetches whether or not an item exists from the database
func (c *Client) ItemExists(ctx context.Context, itemID, userID uint64) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "ItemExists")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachItemIDToSpan(span, itemID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("ItemExists called")

	return c.querier.ItemExists(ctx, itemID, userID)
}

// GetItem fetches an item from the database
func (c *Client) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "GetItem")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachItemIDToSpan(span, itemID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("GetItem called")

	return c.querier.GetItem(ctx, itemID, userID)
}

// GetAllItemsCount fetches the count of items from the database that meet a particular filter
func (c *Client) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetAllItemsCount")
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	return c.querier.GetAllItemsCount(ctx)
}

// GetItems fetches a list of items from the database that meet a particular filter
func (c *Client) GetItems(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.ItemList, error) {
	ctx, span := trace.StartSpan(ctx, "GetItems")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetItems called")

	itemList, err := c.querier.GetItems(ctx, userID, filter)

	return itemList, err
}

// CreateItem creates an item in the database
func (c *Client) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	ctx, span := trace.StartSpan(ctx, "CreateItem")
	defer span.End()

	c.logger.WithValue("input", input).Debug("CreateItem called")

	return c.querier.CreateItem(ctx, input)
}

// UpdateItem updates a particular item. Note that UpdateItem expects the
// provided input to have a valid ID.
func (c *Client) UpdateItem(ctx context.Context, updated *models.Item) error {
	ctx, span := trace.StartSpan(ctx, "UpdateItem")
	defer span.End()

	tracing.AttachItemIDToSpan(span, updated.ID)
	c.logger.WithValue("item_id", updated.ID).Debug("UpdateItem called")

	return c.querier.UpdateItem(ctx, updated)
}

// ArchiveItem archives an item from the database by its ID
func (c *Client) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "ArchiveItem")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachItemIDToSpan(span, itemID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("ArchiveItem called")

	return c.querier.ArchiveItem(ctx, itemID, userID)
}

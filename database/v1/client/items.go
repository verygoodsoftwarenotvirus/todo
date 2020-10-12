package dbclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var _ models.ItemDataManager = (*Client)(nil)

// ItemExists fetches whether or not an item exists from the database.
func (c *Client) ItemExists(ctx context.Context, itemID, userID uint64) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "ItemExists")
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("ItemExists called")

	return c.querier.ItemExists(ctx, itemID, userID)
}

// GetItem fetches an item from the database.
func (c *Client) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	ctx, span := tracing.StartSpan(ctx, "GetItem")
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("GetItem called")

	return c.querier.GetItem(ctx, itemID, userID)
}

// GetAllItemsCount fetches the count of items from the database that meet a particular filter.
func (c *Client) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAllItemsCount")
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	return c.querier.GetAllItemsCount(ctx)
}

// GetAllItems fetches a list of all items in the database.
func (c *Client) GetAllItems(ctx context.Context, results chan []models.Item) error {
	ctx, span := tracing.StartSpan(ctx, "GetAllItems")
	defer span.End()

	c.logger.Debug("GetAllItems called")

	return c.querier.GetAllItems(ctx, results)
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (c *Client) GetItems(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.ItemList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetItems")
	defer span.End()

	tracing.AttachFilterToSpan(span, filter)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		"user_id": userID,
	}).Debug("GetItems called")

	itemList, err := c.querier.GetItems(ctx, userID, filter)

	return itemList, err
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (c *Client) GetItemsForAdmin(ctx context.Context, filter *models.QueryFilter) (*models.ItemList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetItemsForAdmin")
	defer span.End()

	tracing.AttachFilterToSpan(span, filter)

	c.logger.Debug("GetItemsForAdmin called")

	itemList, err := c.querier.GetItemsForAdmin(ctx, filter)

	return itemList, err
}

// GetItemsWithIDs fetches items from the database within a given set of IDs.
func (c *Client) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]models.Item, error) {
	ctx, span := tracing.StartSpan(ctx, "GetItemsWithIDs")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		"user_id":  userID,
		"limit":    limit,
		"id_count": len(ids),
	}).Debug("GetItemsWithIDs called")

	itemList, err := c.querier.GetItemsWithIDs(ctx, userID, limit, ids)

	return itemList, err
}

// GetItemsWithIDsForAdmin fetches items from the database within a given set of IDs.
func (c *Client) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]models.Item, error) {
	ctx, span := tracing.StartSpan(ctx, "GetItemsWithIDsForAdmin")
	defer span.End()

	c.logger.WithValues(map[string]interface{}{
		"limit":    limit,
		"id_count": len(ids),
		"ids":      ids,
	}).Debug(fmt.Sprintf("GetItemsWithIDsForAdmin called: %v", ids))

	itemList, err := c.querier.GetItemsWithIDsForAdmin(ctx, limit, ids)

	return itemList, err
}

// CreateItem creates an item in the database.
func (c *Client) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	ctx, span := tracing.StartSpan(ctx, "CreateItem")
	defer span.End()

	c.logger.WithValue("input", input).Debug("CreateItem called")

	return c.querier.CreateItem(ctx, input)
}

// UpdateItem updates a particular item. Note that UpdateItem expects the
// provided input to have a valid ID.
func (c *Client) UpdateItem(ctx context.Context, updated *models.Item) error {
	ctx, span := tracing.StartSpan(ctx, "UpdateItem")
	defer span.End()

	tracing.AttachItemIDToSpan(span, updated.ID)
	c.logger.WithValue("item_id", updated.ID).Debug("UpdateItem called")

	return c.querier.UpdateItem(ctx, updated)
}

// ArchiveItem archives an item from the database by its ID.
func (c *Client) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx, "ArchiveItem")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachItemIDToSpan(span, itemID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"user_id": userID,
	}).Debug("ArchiveItem called")

	return c.querier.ArchiveItem(ctx, itemID, userID)
}

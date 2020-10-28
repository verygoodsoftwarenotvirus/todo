package dbclient

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var _ models.AuditLogEntryDataManager = (*Client)(nil)

// GetItem fetches an item from the database.
func (c *Client) GetAuditLogEntry(ctx context.Context, entryID uint64) (*models.AuditLogEntry, error) {
	ctx, span := tracing.StartSpan(ctx, "GetItem")
	defer span.End()

	tracing.AttachAuditLogEntryIDToSpan(span, entryID)

	c.logger.WithValues(map[string]interface{}{
		"item_id": entryID,
	}).Debug("GetItem called")

	return c.querier.GetAuditLogEntry(ctx, entryID)
}

// GetAllItemsCount fetches the count of items from the database that meet a particular filter.
func (c *Client) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAllItemsCount")
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	return c.querier.GetAllItemsCount(ctx)
}

// GetAllItems fetches a list of all items in the database.
func (c *Client) GetAllAuditLogEntries(ctx context.Context, results chan []models.AuditLogEntry) error {
	ctx, span := tracing.StartSpan(ctx, "GetAllItems")
	defer span.End()

	c.logger.Debug("GetAllItems called")

	return c.querier.GetAllAuditLogEntries(ctx, results)
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *models.QueryFilter) (*models.AuditLogEntryList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetItems")
	defer span.End()

	tracing.AttachFilterToSpan(span, filter)

	c.logger.Debug("GetItems called")

	itemList, err := c.querier.GetAuditLogEntries(ctx, filter)

	return itemList, err
}

// CreateItem creates an item in the database.
func (c *Client) CreateAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) (*models.AuditLogEntry, error) {
	ctx, span := tracing.StartSpan(ctx, "CreateItem")
	defer span.End()

	c.logger.WithValue("input", input).Debug("CreateItem called")

	return c.querier.CreateAuditLogEntry(ctx, input)
}

package httpclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// ItemExists retrieves whether or not an item exists.
func (c *Client) ItemExists(ctx context.Context, itemID uint64) (exists bool, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return false, ErrInvalidIDProvided
	}

	req, err := c.BuildItemExistsRequest(ctx, itemID)
	if err != nil {
		return false, fmt.Errorf("building request: %w", err)
	}

	return c.checkExistence(ctx, req)
}

// GetItem retrieves an item.
func (c *Client) GetItem(ctx context.Context, itemID uint64) (item *types.Item, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.BuildGetItemRequest(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &item); retrieveErr != nil {
		return nil, retrieveErr
	}

	return item, nil
}

// SearchItems searches for a list of items.
func (c *Client) SearchItems(ctx context.Context, query string, limit uint8) (items []*types.Item, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if query == "" {
		return nil, ErrEmptyQueryProvided
	}

	if limit == 0 {
		limit = 20
	}

	req, err := c.BuildSearchItemsRequest(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &items); retrieveErr != nil {
		return nil, retrieveErr
	}

	return items, nil
}

// GetItems retrieves a list of items.
func (c *Client) GetItems(ctx context.Context, filter *types.QueryFilter) (items *types.ItemList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetItemsRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &items); retrieveErr != nil {
		return nil, retrieveErr
	}

	return items, nil
}

// CreateItem creates an item.
func (c *Client) CreateItem(ctx context.Context, input *types.ItemCreationInput) (item *types.Item, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildCreateItemRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &item)

	return item, err
}

// UpdateItem updates an item.
func (c *Client) UpdateItem(ctx context.Context, item *types.Item) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if item == nil {
		return ErrNilInputProvided
	}

	req, err := c.BuildUpdateItemRequest(ctx, item)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &item)
}

// ArchiveItem archives an item.
func (c *Client) ArchiveItem(ctx context.Context, itemID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.BuildArchiveItemRequest(ctx, itemID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// GetAuditLogForItem retrieves a list of audit log entries pertaining to an item.
func (c *Client) GetAuditLogForItem(ctx context.Context, itemID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.BuildGetAuditLogForItemRequest(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

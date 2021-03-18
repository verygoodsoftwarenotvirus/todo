package httpclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// ItemExists retrieves whether or not an item exists.
func (c *Client) ItemExists(ctx context.Context, itemID uint64) (bool, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return false, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildItemExistsRequest(ctx, itemID)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return false, fmt.Errorf("building request: %w", err)
	}

	exists, err := c.checkExistence(ctx, req)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return false, fmt.Errorf("checking existence for item #%d: %w", itemID, err)
	}

	return exists, nil
}

// GetItem retrieves an item.
func (c *Client) GetItem(ctx context.Context, itemID uint64) (item *types.Item, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetItemRequest(ctx, itemID)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &item); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
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

	req, err := c.requestBuilder.BuildSearchItemsRequest(ctx, query, limit)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &items); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, retrieveErr
	}

	return items, nil
}

// GetItems retrieves a list of items.
func (c *Client) GetItems(ctx context.Context, filter *types.QueryFilter) (items *types.ItemList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.requestBuilder.BuildGetItemsRequest(ctx, filter)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &items); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
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
		tracing.AttachErrorToSpan(span, validationErr)
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.requestBuilder.BuildCreateItemRequest(ctx, input)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if createErr := c.executeRequest(ctx, req, &item); createErr != nil {
		tracing.AttachErrorToSpan(span, createErr)
		return nil, fmt.Errorf("creating item: %w", createErr)
	}

	return item, nil
}

// UpdateItem updates an item.
func (c *Client) UpdateItem(ctx context.Context, item *types.Item) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if item == nil {
		return ErrNilInputProvided
	}

	req, err := c.requestBuilder.BuildUpdateItemRequest(ctx, item)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building request: %w", err)
	}

	if updateErr := c.executeRequest(ctx, req, &item); updateErr != nil {
		tracing.AttachErrorToSpan(span, updateErr)
		return fmt.Errorf("updating item #%d: %w", item.ID, updateErr)
	}

	return nil
}

// ArchiveItem archives an item.
func (c *Client) ArchiveItem(ctx context.Context, itemID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildArchiveItemRequest(ctx, itemID)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building request: %w", err)
	}

	if archiveErr := c.executeRequest(ctx, req, nil); archiveErr != nil {
		tracing.AttachErrorToSpan(span, archiveErr)
		return fmt.Errorf("archiving item #%d: %w", itemID, archiveErr)
	}

	return nil
}

// GetAuditLogForItem retrieves a list of audit log entries pertaining to an item.
func (c *Client) GetAuditLogForItem(ctx context.Context, itemID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetAuditLogForItemRequest(ctx, itemID)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, fmt.Errorf("fetching audit log entries for item #%d: %w", itemID, retrieveErr)
	}

	return entries, nil
}

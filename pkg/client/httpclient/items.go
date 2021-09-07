package httpclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// ItemExists retrieves whether an item exists.
func (c *Client) ItemExists(ctx context.Context, itemID string) (bool, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	if itemID == "" {
		return false, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	req, err := c.requestBuilder.BuildItemExistsRequest(ctx, itemID)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "building item existence request")
	}

	exists, err := c.responseIsOK(ctx, req)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "checking existence for item %s", itemID)
	}

	return exists, nil
}

// GetItem gets an item.
func (c *Client) GetItem(ctx context.Context, itemID string) (*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	if itemID == "" {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	req, err := c.requestBuilder.BuildGetItemRequest(ctx, itemID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building get item request")
	}

	var item *types.Item
	if err = c.fetchAndUnmarshal(ctx, req, &item); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving item")
	}

	return item, nil
}

// SearchItems searches through a list of items.
func (c *Client) SearchItems(ctx context.Context, query string, limit uint8) ([]*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	if query == "" {
		return nil, ErrEmptyQueryProvided
	}

	if limit == 0 {
		limit = 20
	}

	logger = logger.WithValue(keys.SearchQueryKey, query).WithValue(keys.FilterLimitKey, limit)

	req, err := c.requestBuilder.BuildSearchItemsRequest(ctx, query, limit)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building search for items request")
	}

	var items []*types.Item
	if err = c.fetchAndUnmarshal(ctx, req, &items); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving items")
	}

	return items, nil
}

// GetItems retrieves a list of items.
func (c *Client) GetItems(ctx context.Context, filter *types.QueryFilter) (*types.ItemList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.loggerWithFilter(filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetItemsRequest(ctx, filter)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building items list request")
	}

	var items *types.ItemList
	if err = c.fetchAndUnmarshal(ctx, req, &items); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving items")
	}

	return items, nil
}

// CreateItem creates an item.
func (c *Client) CreateItem(ctx context.Context, input *types.ItemCreationInput) (*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildCreateItemRequest(ctx, input)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building create item request")
	}

	var item *types.Item
	if err = c.fetchAndUnmarshal(ctx, req, &item); err != nil {
		return nil, observability.PrepareError(err, logger, span, "creating item")
	}

	return item, nil
}

// UpdateItem updates an item.
func (c *Client) UpdateItem(ctx context.Context, item *types.Item) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	if item == nil {
		return ErrNilInputProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, item.ID)
	tracing.AttachItemIDToSpan(span, item.ID)

	req, err := c.requestBuilder.BuildUpdateItemRequest(ctx, item)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building update item request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, &item); err != nil {
		return observability.PrepareError(err, logger, span, "updating item %s", item.ID)
	}

	return nil
}

// ArchiveItem archives an item.
func (c *Client) ArchiveItem(ctx context.Context, itemID string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	if itemID == "" {
		return ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	req, err := c.requestBuilder.BuildArchiveItemRequest(ctx, itemID)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building archive item request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "archiving item %s", itemID)
	}

	return nil
}

// GetAuditLogForItem retrieves a list of audit log entries pertaining to an item.
func (c *Client) GetAuditLogForItem(ctx context.Context, itemID string) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	if itemID == "" {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	req, err := c.requestBuilder.BuildGetAuditLogForItemRequest(ctx, itemID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building get audit log entries for item request")
	}

	var entries []*types.AuditLogEntry
	if err = c.fetchAndUnmarshal(ctx, req, &entries); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving plan")
	}

	return entries, nil
}

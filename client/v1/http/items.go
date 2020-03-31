package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

const (
	itemsBasePath = "items"
)

// BuildItemExistsRequest builds an HTTP request for checking the existence of an item
func (c *V1Client) BuildItemExistsRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := trace.StartSpan(ctx, "BuildItemExistsRequest")
	defer span.End()

	uri := c.BuildURL(
		nil,
		itemsBasePath,
		strconv.FormatUint(itemID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodHead, uri, nil)
}

// ItemExists retrieves whether or not an item exists
func (c *V1Client) ItemExists(ctx context.Context, itemID uint64) (exists bool, err error) {
	ctx, span := trace.StartSpan(ctx, "ItemExists")
	defer span.End()

	req, err := c.BuildItemExistsRequest(ctx, itemID)
	if err != nil {
		return false, fmt.Errorf("building request: %w", err)
	}

	return c.checkExistence(ctx, req)
}

// BuildGetItemRequest builds an HTTP request for fetching an item
func (c *V1Client) BuildGetItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := trace.StartSpan(ctx, "BuildGetItemRequest")
	defer span.End()

	uri := c.BuildURL(
		nil,
		itemsBasePath,
		strconv.FormatUint(itemID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetItem retrieves an item
func (c *V1Client) GetItem(ctx context.Context, itemID uint64) (item *models.Item, err error) {
	ctx, span := trace.StartSpan(ctx, "GetItem")
	defer span.End()

	req, err := c.BuildGetItemRequest(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &item); retrieveErr != nil {
		return nil, retrieveErr
	}

	return item, nil
}

// BuildGetItemsRequest builds an HTTP request for fetching items
func (c *V1Client) BuildGetItemsRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	ctx, span := trace.StartSpan(ctx, "BuildGetItemsRequest")
	defer span.End()

	uri := c.BuildURL(
		filter.ToValues(),
		itemsBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetItems retrieves a list of items
func (c *V1Client) GetItems(ctx context.Context, filter *models.QueryFilter) (items *models.ItemList, err error) {
	ctx, span := trace.StartSpan(ctx, "GetItems")
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

// BuildCreateItemRequest builds an HTTP request for creating an item
func (c *V1Client) BuildCreateItemRequest(ctx context.Context, input *models.ItemCreationInput) (*http.Request, error) {
	ctx, span := trace.StartSpan(ctx, "BuildCreateItemRequest")
	defer span.End()

	uri := c.BuildURL(
		nil,
		itemsBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// CreateItem creates an item
func (c *V1Client) CreateItem(ctx context.Context, input *models.ItemCreationInput) (item *models.Item, err error) {
	ctx, span := trace.StartSpan(ctx, "CreateItem")
	defer span.End()

	req, err := c.BuildCreateItemRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &item)
	return item, err
}

// BuildUpdateItemRequest builds an HTTP request for updating an item
func (c *V1Client) BuildUpdateItemRequest(ctx context.Context, item *models.Item) (*http.Request, error) {
	ctx, span := trace.StartSpan(ctx, "BuildUpdateItemRequest")
	defer span.End()

	uri := c.BuildURL(
		nil,
		itemsBasePath,
		strconv.FormatUint(item.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPut, uri, item)
}

// UpdateItem updates an item
func (c *V1Client) UpdateItem(ctx context.Context, item *models.Item) error {
	ctx, span := trace.StartSpan(ctx, "UpdateItem")
	defer span.End()

	req, err := c.BuildUpdateItemRequest(ctx, item)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &item)
}

// BuildArchiveItemRequest builds an HTTP request for updating an item
func (c *V1Client) BuildArchiveItemRequest(ctx context.Context, itemID uint64) (*http.Request, error) {
	ctx, span := trace.StartSpan(ctx, "BuildArchiveItemRequest")
	defer span.End()

	uri := c.BuildURL(
		nil,
		itemsBasePath,
		strconv.FormatUint(itemID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveItem archives an item
func (c *V1Client) ArchiveItem(ctx context.Context, itemID uint64) error {
	ctx, span := trace.StartSpan(ctx, "ArchiveItem")
	defer span.End()

	req, err := c.BuildArchiveItemRequest(ctx, itemID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

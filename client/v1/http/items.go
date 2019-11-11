package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	itemsBasePath = "items"
)

// BuildGetItemRequest builds an HTTP request for fetching an item
func (c *V1Client) BuildGetItemRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetItem retrieves an item
func (c *V1Client) GetItem(ctx context.Context, id uint64) (item *models.Item, err error) {
	req, err := c.BuildGetItemRequest(ctx, id)
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
	uri := c.BuildURL(filter.ToValues(), itemsBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetItems retrieves a list of items
func (c *V1Client) GetItems(ctx context.Context, filter *models.QueryFilter) (items *models.ItemList, err error) {
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
func (c *V1Client) BuildCreateItemRequest(ctx context.Context, body *models.ItemCreationInput) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath)

	return c.buildDataRequest(http.MethodPost, uri, body)
}

// CreateItem creates an item
func (c *V1Client) CreateItem(ctx context.Context, input *models.ItemCreationInput) (item *models.Item, err error) {
	req, err := c.BuildCreateItemRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &item)
	return item, err
}

// BuildUpdateItemRequest builds an HTTP request for updating an item
func (c *V1Client) BuildUpdateItemRequest(ctx context.Context, updated *models.Item) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(updated.ID, 10))

	return c.buildDataRequest(http.MethodPut, uri, updated)
}

// UpdateItem updates an item
func (c *V1Client) UpdateItem(ctx context.Context, updated *models.Item) error {
	req, err := c.BuildUpdateItemRequest(ctx, updated)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &updated)
}

// BuildArchiveItemRequest builds an HTTP request for updating an item
func (c *V1Client) BuildArchiveItemRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodDelete, uri, nil)
}

// ArchiveItem archives an item
func (c *V1Client) ArchiveItem(ctx context.Context, id uint64) error {
	req, err := c.BuildArchiveItemRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

package client

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	itemsBasePath = "items"
)

// BuildGetItemRequest builds an http Request for fetching an item
func (c *V1Client) BuildGetItemRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetItem retrieves an item
func (c *V1Client) GetItem(ctx context.Context, id uint64) (item *models.Item, err error) {
	req, err := c.BuildGetItemRequest(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	if err = c.retrieve(ctx, req, &item); err != nil {
		return nil, err
	}

	return item, nil
}

// BuildGetItemsRequest builds an http Request for fetching items
func (c *V1Client) BuildGetItemsRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	uri := c.BuildURL(filter.ToValues(), itemsBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetItems retrieves a list of items
func (c *V1Client) GetItems(ctx context.Context, filter *models.QueryFilter) (items *models.ItemList, err error) {
	req, err := c.BuildGetItemsRequest(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	if err = c.retrieve(ctx, req, &items); err != nil  {
		return nil, err
	}

	return items, nil
}

// BuildCreateItemRequest builds an http Request for creating an item
func (c *V1Client) BuildCreateItemRequest(ctx context.Context, body *models.ItemInput) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath)

	return c.buildDataRequest(http.MethodPost, uri, body)
}

// CreateItem creates an item
func (c *V1Client) CreateItem(ctx context.Context, input *models.ItemInput) (item *models.Item, err error) {
	req, err := c.BuildCreateItemRequest(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.makeRequest(ctx, req, &item)
	return item, err
}

// BuildUpdateItemRequest builds an http Request for updating an item
func (c *V1Client) BuildUpdateItemRequest(ctx context.Context, updated *models.Item) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(updated.ID, 10))

	return c.buildDataRequest(http.MethodPut, uri, updated)
}

// UpdateItem updates an item
func (c *V1Client) UpdateItem(ctx context.Context, updated *models.Item) error {
	req, err := c.BuildUpdateItemRequest(ctx, updated)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	return c.makeRequest(ctx, req, &updated)
}

// BuildDeleteItemRequest builds an http Request for updating an item
func (c *V1Client) BuildDeleteItemRequest(ctx context.Context, id uint64) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))

	return http.NewRequest(http.MethodDelete, uri, nil)
}

// DeleteItem deletes an item
func (c *V1Client) DeleteItem(ctx context.Context, id uint64) error {
	req, err := c.BuildDeleteItemRequest(ctx, id)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	return c.makeRequest(ctx, req, nil)
}

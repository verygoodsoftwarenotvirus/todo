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

// GetItem gets an item
func (c *V1Client) GetItem(ctx context.Context, id uint64) (item *models.Item, err error) {
	logger := c.logger.WithValue("id", id)
	logger.Debug("GetItem called")

	req, err := c.BuildGetItemRequest(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &item)
	return item, err
}

// BuildGetItemsRequest builds an http Request for fetching items
func (c *V1Client) BuildGetItemsRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	uri := c.BuildURL(filter.ToValues(), itemsBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetItems gets a list of items
func (c *V1Client) GetItems(ctx context.Context, filter *models.QueryFilter) (items *models.ItemList, err error) {
	logger := c.logger.WithValue("filter", filter)
	logger.Debug("GetItems called")

	req, err := c.BuildGetItemsRequest(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &items)
	return items, err
}

// BuildCreateItemRequest builds an http Request for creating an item
func (c *V1Client) BuildCreateItemRequest(ctx context.Context, body *models.ItemInput) (*http.Request, error) {
	uri := c.BuildURL(nil, itemsBasePath)

	return c.buildDataRequest(http.MethodPost, uri, body)
}

// CreateItem creates an item
func (c *V1Client) CreateItem(ctx context.Context, input *models.ItemInput) (item *models.Item, err error) {
	logger := c.logger.WithValues(map[string]interface{}{
		"input_name":    input.Name,
		"input_details": input.Details,
	})
	logger.Debug("CreateItem called")

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
	logger := c.logger.WithValue("id", updated.ID)
	logger.Debug("UpdateItem called")

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
	c.logger.WithValue("id", id).Debug("DeleteItem called")

	req, err := c.BuildDeleteItemRequest(ctx, id)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	return c.makeRequest(ctx, req, nil)
}

// func (c *V1Client) buildItemsFeed(conn *websocket.Conn, itemChan chan models.Item) {
// 	defer conn.Close()
// 	for {
// 		_, message, err := conn.ReadMessage()
// 		if err != nil {
// 			c.logger.Errorf("error: %v", err)
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				c.logger.Println("something is supposed to happen here?")
// 			}
// 			break
// 		}
// 		c.logger.Debug("message read from connection")

// 		e := &models.Event{}

// 		if err := json.NewDecoder(bytes.NewReader(message)).Decode(e); err != nil {
// 			c.logger.Errorf("error decoding item: %v", err)
// 			break
// 		}

// 		if _, ok := e.Data.(models.Item); !ok {
// 			continue
// 		}

// 		item := e.Data.(models.Item)
// 		c.logger.Debug("writing item %d to channel", item.ID)
// 		itemChan <- item
// 	}
// }

// func (c *V1Client) NewItemsFeed() (<-chan models.Item, error) {
// 	itemChan := make(chan models.Item)
// 	fq := &FeedQuery{
// 		DataTypes: []string{"item"},
// 		Events:    []string{"create"},
// 		Topics:    []string{"*"},
// 	}
// 	u := c.buildURL(fq.Values(), "event_feed")
// 	conn, err := c.DialWebsocket(fq)
// 	if err != nil {
// 		c.logger.Debug("encountered error dialing %q: %v", u.String(), err)
// 		return nil, err
// 	}

// 	if err != nil {
// 		return nil, err
// 	}
// 	go c.buildItemsFeed(conn, itemChan)
// 	return itemChan, nil
// }

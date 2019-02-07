package client

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const (
	itemsBasePath = "items"
)

// GetItem gets an item
func (c *V1Client) GetItem(ctx context.Context, id uint64) (item *models.Item, err error) {
	logger := c.logger.WithValue("id", id)
	logger.Debug("GetItem called")

	var span opentracing.Span // If I don't set this value here, then ctx gets over
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetItem")
	span.SetTag("itemID", id)
	defer span.Finish()

	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))
	err = c.get(ctx, uri, &item)
	return item, err
}

// GetItemCount an item
func (c *V1Client) GetItemCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	logger := c.logger.WithValue("filter", filter)
	logger.Debug("GetItemCount called")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetItemCount")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	x := models.CountResponse{}
	uri := c.BuildURL(filter.ToValues(), itemsBasePath, "count")
	err := c.get(ctx, uri, &x)
	return x.Count, err
}

// GetItems gets a list of items
func (c *V1Client) GetItems(ctx context.Context, filter *models.QueryFilter) (items *models.ItemList, err error) {
	logger := c.logger.WithValue("filter", filter)
	logger.Debug("GetItems called")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetItems")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(filter.ToValues(), itemsBasePath)
	err = c.get(ctx, uri, &items)
	return items, err
}

// CreateItem creates an item
func (c *V1Client) CreateItem(ctx context.Context, input *models.ItemInput) (item *models.Item, err error) {
	logger := c.logger.WithValues(map[string]interface{}{
		"input_name":    input.Name,
		"input_details": input.Details,
	})
	logger.Debug("CreateItem called")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateItem")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, itemsBasePath)
	err = c.post(ctx, uri, input, &item)
	return item, err
}

// UpdateItem updates an item
func (c *V1Client) UpdateItem(ctx context.Context, updated *models.Item) error {
	logger := c.logger.WithValue("id", updated.ID)
	logger.Debug("UpdateItem called")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "UpdateItem")
	span.SetTag("itemID", updated.ID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(updated.ID, 10))
	err := c.put(ctx, uri, updated, &updated)
	return err
}

// DeleteItem deletes an item
func (c *V1Client) DeleteItem(ctx context.Context, id uint64) error {
	logger := c.logger.WithValue("id", id)
	logger.Debug("DeleteItem called")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "DeleteItem")
	span.SetTag("itemID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))
	err := c.delete(ctx, uri)
	return err
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

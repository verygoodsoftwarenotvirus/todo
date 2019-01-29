package client

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const itemsBasePath = "items"

// GetItem gets an item
func (c *V1Client) GetItem(ctx context.Context, id uint64) (item *models.Item, err error) {
	span := c.tracer.StartSpan("GetItem")
	span.SetTag("itemID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))
	return item, c.get(ctx, uri, &item)
}

// GetItemCount an item
func (c *V1Client) GetItemCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	span := c.tracer.StartSpan("GetItemCount")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	x := models.CountResponse{}
	uri := c.BuildURL(filter.ToValues(), itemsBasePath, "count")
	return x.Count, c.get(ctx, uri, &x)
}

// GetItems gets a list of items
func (c *V1Client) GetItems(ctx context.Context, filter *models.QueryFilter) (items *models.ItemList, err error) {
	span := c.tracer.StartSpan("GetItems")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(filter.ToValues(), itemsBasePath)
	return items, c.get(ctx, uri, &items)
}

// CreateItem creates an item
func (c *V1Client) CreateItem(ctx context.Context, input *models.ItemInput) (item *models.Item, err error) {
	span := c.tracer.StartSpan("CreateItem")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, itemsBasePath)
	return item, c.post(ctx, uri, input, &item)
}

// UpdateItem updates an item
func (c *V1Client) UpdateItem(ctx context.Context, updated *models.Item) (err error) {
	span := c.tracer.StartSpan("UpdateItem")
	span.SetTag("itemID", updated.ID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(updated.ID, 10))
	return c.put(ctx, uri, updated, &updated)
}

// DeleteItem deletes an item
func (c *V1Client) DeleteItem(ctx context.Context, id uint64) error {
	span := c.tracer.StartSpan("DeleteItem")
	span.SetTag("itemID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10))
	return c.delete(ctx, uri)
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
// 		c.logger.Debugln("message read from connection")

// 		e := &models.Event{}

// 		if err := json.NewDecoder(bytes.NewReader(message)).Decode(e); err != nil {
// 			c.logger.Errorf("error decoding item: %v", err)
// 			break
// 		}

// 		if _, ok := e.Data.(models.Item); !ok {
// 			continue
// 		}

// 		item := e.Data.(models.Item)
// 		c.logger.Debugf("writing item %d to channel", item.ID)
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
// 		c.logger.Debugf("encountered error dialing %q: %v", u.String(), err)
// 		return nil, err
// 	}

// 	if err != nil {
// 		return nil, err
// 	}
// 	go c.buildItemsFeed(conn, itemChan)
// 	return itemChan, nil
// }

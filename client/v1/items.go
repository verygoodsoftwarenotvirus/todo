package client

import (
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const itemsBasePath = "items"

// GetItem gets an item
func (c *V1Client) GetItem(id uint64) (item *models.Item, err error) {
	return item, c.get(c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10)), &item)
}

// GetItemCount an item
func (c *V1Client) GetItemCount(filter *models.QueryFilter) (uint64, error) {
	x := models.CountResponse{}
	return x.Count, c.get(c.BuildURL(filter.ToValues(), itemsBasePath, "count"), &x)
}

// GetItems gets a list of items
func (c *V1Client) GetItems(filter *models.QueryFilter) (items *models.ItemList, err error) {
	return items, c.get(c.BuildURL(filter.ToValues(), itemsBasePath), &items)
}

// CreateItem creates an item
func (c *V1Client) CreateItem(input *models.ItemInput) (item *models.Item, err error) {
	return item, c.post(c.BuildURL(nil, itemsBasePath), input, &item)
}

// UpdateItem updates an item
func (c *V1Client) UpdateItem(updated *models.Item) (err error) {
	return c.put(c.BuildURL(nil, itemsBasePath, strconv.FormatUint(updated.ID, 10)), updated, &updated)
}

// DeleteItem deletes an item
func (c *V1Client) DeleteItem(id uint64) error {
	return c.delete(c.BuildURL(nil, itemsBasePath, strconv.FormatUint(id, 10)))
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

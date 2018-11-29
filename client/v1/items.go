package client

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/gorilla/websocket"
)

const itemsBasePath = "items"

func (c *V1Client) GetItem(id uint) (item *models.Item, err error) {
	p := fmt.Sprintf("%s/%d", itemsBasePath, id)
	u := c.BuildURL(nil, p)
	item = &models.Item{}

	err = c.get(u, &item)

	return
}

func (c *V1Client) GetItems(filter *models.QueryFilter) (items []models.Item, err error) {
	var u string
	if filter == nil {
		u = c.BuildURL(nil, itemsBasePath)
	} else {
		u = c.BuildURL(filter.ToValues(), itemsBasePath)
	}

	items = []models.Item{}
	err = c.get(u, &items)

	return
}

func (c *V1Client) CreateItem(input *models.ItemInput) (*models.Item, error) {
	u := c.BuildURL(nil, itemsBasePath)
	item := &models.Item{}

	err := c.post(u, input, item)

	return item, err
}

func (c *V1Client) UpdateItem(updated *models.Item) (err error) {
	p := fmt.Sprintf("%s/%d", itemsBasePath, updated.ID)
	u := c.BuildURL(nil, p)

	return c.put(u, updated, &models.Item{})
}

func (c *V1Client) DeleteItem(id uint) error {
	p := fmt.Sprintf("%s/%d", itemsBasePath, id)
	u := c.BuildURL(nil, p)

	return c.delete(u)
}

func (c *V1Client) buildItemsFeed(conn *websocket.Conn, itemChan chan models.Item) {
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			c.logger.Errorf("error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Println("something is supposed to happen here?")
			}
			break
		}
		c.logger.Debugln("message read from connection")

		e := &models.Event{}

		if err := json.NewDecoder(bytes.NewReader(message)).Decode(e); err != nil {
			c.logger.Errorf("error decoding item: %v", err)
			break
		}

		if _, ok := e.Data.(models.Item); !ok {
			continue
		}

		item := e.Data.(models.Item)
		c.logger.Debugf("writing item %d to channel", item.ID)
		itemChan <- item
	}
}

func (c *V1Client) NewItemsFeed() (<-chan models.Item, error) {
	itemChan := make(chan models.Item)
	fq := &FeedQuery{
		DataTypes: []string{"item"},
		Events:    []string{"create"},
		Topics:    []string{"*"},
	}
	u := c.buildURL(fq.Values(), "event_feed")
	conn, err := c.DialWebsocket(fq)
	if err != nil {
		c.logger.Debugf("encountered error dialing %q: %v", u.String(), err)
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	go c.buildItemsFeed(conn, itemChan)
	return itemChan, nil
}

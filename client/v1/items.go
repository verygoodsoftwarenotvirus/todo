package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/gorilla/websocket"
)

const itemsBasePath = "items"

func (c *V1Client) GetItemChannel() {
	// p := fmt.Sprintf("%s/%d", itemsBasePath, id)
	// u := c.BuildURL(nil, p)
	// item = &models.Item{}

	// err = c.get(u, &item)

	// return
}

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
		u = c.BuildURL(filter.ToMap(), itemsBasePath)
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

		item := models.Item{}
		if err := json.NewDecoder(bytes.NewReader(message)).Decode(&item); err != nil {
			c.logger.Errorf("error decoding item: %v", err)
			break
		}
		c.logger.Debugf("writing item %d to channel", item.ID)
		itemChan <- item
	}
}

func (c *V1Client) ItemsFeed() (<-chan models.Item, error) {
	itemChan := make(chan models.Item)
	if !c.IsUp() {
		c.logger.Debugln("returning early from ItemsFeed because the service is down")
		return nil, errors.New("service is down")
	}

	u := c.buildURL(nil, "items", "feed")
	u.Scheme = "wss"
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c.logger.Debugf("connecting to websocket at %q", u.String())
	conn, res, err := dialer.Dial(u.String(), nil)
	if err != nil {
		c.logger.Debugf("encountered error dialing %q: %v", u.String(), err)
		return nil, err
	}

	if res.StatusCode < http.StatusBadRequest {
		go c.buildItemsFeed(conn, itemChan)
	} else {
		return nil, fmt.Errorf("encountered status code: %d when trying to reach websocket", res.StatusCode)
	}

	return itemChan, nil
}

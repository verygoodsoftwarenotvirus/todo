package client

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
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

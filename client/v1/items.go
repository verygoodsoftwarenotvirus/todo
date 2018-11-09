package client

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

func (c *V1Client) GetItem(id uint) (*models.Item, error) {
	return nil, nil
}

func (c *V1Client) GetItems(filter *models.QueryFilter) ([]models.Item, error) {
	return nil, nil
}

func (c *V1Client) CreateItem(input *models.ItemInput) (*models.Item, error) {
	return nil, nil
}

func (c *V1Client) UpdateItem(updated *models.Item) error {
	return nil
}

func (c *V1Client) DeleteItem(id uint) error {
	return nil
}

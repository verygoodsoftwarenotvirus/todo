package client

import (
	v1 "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

type TodoClient interface {
	GetItem(id uint) (*models.Item, error)
	GetItems(filter *models.QueryFilter) ([]models.Item, error)
	CreateItem(input *models.ItemInput) (*models.Item, error)
	UpdateItem(updated *models.Item) error
	DeleteItem(id uint) error
}

func NewClient(address, authToken string, debug bool) (TodoClient, error) {
	cfg := &v1.Config{
		Debug:     debug,
		Address:   address,
		AuthToken: authToken,
	}
	return v1.NewClient(cfg)
}

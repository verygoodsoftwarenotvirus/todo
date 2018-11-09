package client

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	v1 "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"
)

type Client interface {
	GetItem(id uint) (*models.Item, error)
	GetItems(filter *models.QueryFilter) ([]models.Item, error)
	CreateItem(input *models.ItemInput) error
	UpdateItem(updated *models.Item) error
	DeleteItem(id uint) error
}

func NewClient(storeURL string, username string, password string) (Client, error) {
	return v1.NewClient(storeURL, username, password, http.DefaultClient)
}

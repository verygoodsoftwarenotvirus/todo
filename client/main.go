package client

import (
	"net/http"

	v1 "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

// TodoClient defines a Todo service client
type TodoClient interface {
	GetItem(id uint64) (*models.Item, error)
	GetItems(filter *models.QueryFilter) (*models.ItemList, error)
	CreateItem(input *models.ItemInput) (*models.Item, error)
	UpdateItem(updated *models.Item) error
	DeleteItem(id uint64) error
}

// NewClient builds a new TodoClient
func NewClient(address, clientID, clientSecret string, logger *logrus.Logger, client *http.Client, debug bool) (TodoClient, error) {
	return v1.NewClient(address, clientID, clientSecret, logger, client, debug)
}

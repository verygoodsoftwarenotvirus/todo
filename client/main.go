package client

import (
	"context"
	"net/http"

	v1 "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// TodoClient defines a Todo service client
type TodoClient interface {
	GetItem(ctx context.Context, id uint64) (*models.Item, error)
	GetItems(ctx context.Context, filter *models.QueryFilter) (*models.ItemList, error)
	CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error)
	UpdateItem(ctx context.Context, updated *models.Item) error
	DeleteItem(ctx context.Context, id uint64) error
}

// NewClient builds a new TodoClient
func NewClient(
	address,
	clientID,
	clientSecret string,
	logger *logrus.Logger,
	newLogger logging.Logger,
	client *http.Client,
	tracer opentracing.Tracer,
	debug bool,
) (TodoClient, error) {
	return v1.NewClient(address, clientID, clientSecret, logger, newLogger, client, tracer, debug)
}

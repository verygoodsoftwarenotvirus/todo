package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/events/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const MiddlewareCtxKey models.ContextKey = "item_input"

type (
	ItemsService struct {
		logger   *logrus.Logger
		db       database.Database
		upgrader websocket.Upgrader
		eventHub *events.EventHub
		// cachedItems []models.Item
	}

	ItemsServiceConfig struct {
		Logger   *logrus.Logger
		DB       database.Database
		EventHub *events.EventHub
	}
)

func NewItemsService(cfg ItemsServiceConfig) *ItemsService {
	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
	}

	return &ItemsService{
		logger:   cfg.Logger,
		db:       cfg.DB,
		upgrader: websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		eventHub: cfg.EventHub,
		// itemHub:  newItemHub(),
	}
}

package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const MiddlewareCtxKey models.ContextKey = "item_input"

type (
	ItemsService struct {
		logger   *logrus.Logger
		db       database.Database
		upgrader websocket.Upgrader
		// cachedItems []models.Item
	}

	ItemsServiceConfig struct {
		Logger   *logrus.Logger
		Database database.Database
	}
)

func NewItemsService(cfg ItemsServiceConfig) *ItemsService {
	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
	}

	return &ItemsService{
		logger:   cfg.Logger,
		db:       cfg.Database,
		upgrader: websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}
}

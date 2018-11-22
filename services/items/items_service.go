package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	// "gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ItemsService struct {
	logger   *logrus.Logger
	db       database.Database
	upgrader websocket.Upgrader
	itemHub  *ItemHub
	// cachedItems []models.Item
}

type ItemsServiceConfig struct {
	Logger *logrus.Logger
	DB     database.Database
}

func NewItemsService(cfg ItemsServiceConfig) *ItemsService {
	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
	}

	return &ItemsService{
		logger:   cfg.Logger,
		db:       cfg.DB,
		upgrader: websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		itemHub:  newItemHub(),
	}
}

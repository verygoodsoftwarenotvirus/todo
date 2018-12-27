package items

import (
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const MiddlewareCtxKey models.ContextKey = "item_input"

type (
	ItemsService struct {
		logger        *logrus.Logger
		db            database.Database
		upgrader      websocket.Upgrader
		userIDFetcher func(*http.Request) uint64
		// cachedItems []models.Item
	}

	ItemsServiceConfig struct {
		Logger        *logrus.Logger
		Database      database.Database
		UserIDFetcher func(*http.Request) uint64
	}
)

func NewItemsService(cfg ItemsServiceConfig) (*ItemsService, error) {
	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
	}

	if cfg.UserIDFetcher == nil {
		return nil, errors.New("UserIDFetcher cannot be nil")
	}

	return &ItemsService{
		logger:   cfg.Logger,
		db:       cfg.Database,
		userIDFetcher: cfg.UserIDFetcher,
		upgrader: websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}, nil
}

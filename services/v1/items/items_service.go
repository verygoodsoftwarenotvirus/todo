package items

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const MiddlewareCtxKey models.ContextKey = "item_input"

type (
	// ItemsService handles TODO List items
	ItemsService struct {
		logger        *logrus.Logger
		db            database.Database
		upgrader      websocket.Upgrader
		userIDFetcher func(*http.Request) uint64
	}
)

// UserIDFetcher is a function that fetches user IDs
type UserIDFetcher func(*http.Request) uint64

// ProvideItemsService builds a new ItemsService
func ProvideItemsService(logger *logrus.Logger, db database.Database, userIDFetcher UserIDFetcher) *ItemsService {

	return &ItemsService{
		logger:        logger,
		db:            db,
		userIDFetcher: userIDFetcher,
		upgrader:      websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}
}

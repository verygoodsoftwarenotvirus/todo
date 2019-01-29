package items

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// MiddlewareCtxKey is a string alias we can use for referring to item input data in contexts
	MiddlewareCtxKey models.ContextKey = "item_input"
)

type (
	// Service handles TODO List items
	Service struct {
		logger        *logrus.Logger
		db            database.Database
		upgrader      websocket.Upgrader
		userIDFetcher func(*http.Request) uint64
	}
)

// UserIDFetcher is a function that fetches user IDs
type UserIDFetcher func(*http.Request) uint64

// ProvideItemsService builds a new ItemsService
func ProvideItemsService(logger *logrus.Logger, db database.Database, userIDFetcher UserIDFetcher) *Service {

	return &Service{
		logger:        logger,
		db:            db,
		userIDFetcher: userIDFetcher,
		upgrader:      websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}
}

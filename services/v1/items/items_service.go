package items

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"github.com/opentracing/opentracing-go"
)

const (
	// MiddlewareCtxKey is a string alias we can use for referring to item input data in contexts
	MiddlewareCtxKey models.ContextKey = "item_input"
)

type (
	// ServiceTracer is an arbitrary type alias we're using for dependency injection
	ServiceTracer opentracing.Tracer

	// Service handles TODO List items
	Service struct {
		logger        logging.Logger
		db            database.Database
		upgrader      websocket.Upgrader
		tracer        opentracing.Tracer
		userIDFetcher UserIDFetcher
		itemIDFetcher ItemIDFetcher
	}
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideItemsService,
		ProvideItemsServiceTracer,
	)
)

// ProvideItemsServiceTracer provides a UserServiceTracer from an tracer building function
func ProvideItemsServiceTracer() (ServiceTracer, error) {
	return tracing.ProvideTracer("todo-server-items-service")
}

// UserIDFetcher is a function that fetches user IDs
type UserIDFetcher func(*http.Request) uint64

// ItemIDFetcher is a function that fetches item IDs
type ItemIDFetcher func(*http.Request) uint64

// ProvideItemsService builds a new ItemsService
func ProvideItemsService(logger logging.Logger, db database.Database, userIDFetcher UserIDFetcher, itemIDFetcher ItemIDFetcher, tracer ServiceTracer) *Service {
	svc := &Service{
		logger:        logger,
		db:            db,
		tracer:        tracer,
		userIDFetcher: userIDFetcher,
		itemIDFetcher: itemIDFetcher,
		upgrader:      websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}
	return svc
}

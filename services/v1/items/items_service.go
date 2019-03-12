package items

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
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

	serviceName = "items_service"
)

type (
	// Tracer is an arbitrary type alias we're using for dependency injection
	Tracer opentracing.Tracer

	// Service handles TODO List items
	Service struct {
		logger        logging.Logger
		db            database.Database
		upgrader      websocket.Upgrader
		tracer        opentracing.Tracer
		userIDFetcher UserIDFetcher
		itemIDFetcher ItemIDFetcher
		encoder       encoding.ResponseEncoder
	}
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideItemsService,
	)
)

// UserIDFetcher is a function that fetches user IDs
type UserIDFetcher func(*http.Request) uint64

// ItemIDFetcher is a function that fetches item IDs
type ItemIDFetcher func(*http.Request) uint64

// ProvideItemsService builds a new ItemsService
func ProvideItemsService(
	logger logging.Logger,
	db database.Database,
	userIDFetcher UserIDFetcher,
	itemIDFetcher ItemIDFetcher,
	encoder encoding.ResponseEncoder,
) *Service {
	svc := &Service{
		logger:        logger.WithName(serviceName),
		db:            db,
		tracer:        tracing.ProvideTracer(serviceName),
		encoder:       encoder,
		userIDFetcher: userIDFetcher,
		itemIDFetcher: itemIDFetcher,
		upgrader:      websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}
	return svc
}

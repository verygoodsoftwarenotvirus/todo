package items

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

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
		itemDatabase  models.ItemDataManager
		tracer        opentracing.Tracer
		userIDFetcher UserIDFetcher
		itemIDFetcher ItemIDFetcher
		encoder       encoding.ServerEncoderDecoder
	}
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
	encoder encoding.ServerEncoderDecoder,
) *Service {
	svc := &Service{
		logger:        logger.WithName(serviceName),
		itemDatabase:  db,
		tracer:        tracing.ProvideTracer(serviceName),
		encoder:       encoder,
		userIDFetcher: userIDFetcher,
		itemIDFetcher: itemIDFetcher,
	}
	return svc
}

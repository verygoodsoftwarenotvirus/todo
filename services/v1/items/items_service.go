package items

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	// MiddlewareCtxKey is a string alias we can use for referring to item input data in contexts
	MiddlewareCtxKey models.ContextKey   = "item_input"
	counterName      metrics.CounterName = "items"
	serviceName                          = "items_service"
)

type (
	// Service handles TODO List items
	Service struct {
		logger        logging.Logger
		itemCounter   metrics.UnitCounter
		itemDatabase  models.ItemDataManager
		userIDFetcher UserIDFetcher
		itemIDFetcher ItemIDFetcher
		encoder       encoding.EncoderDecoder
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
	encoder encoding.EncoderDecoder,
	itemCounterProvider metrics.UnitCounterProvider,
) (*Service, error) {
	itemCounter, err := itemCounterProvider(counterName, "the number of items managed by the items service")
	if err != nil {
		return nil, errors.Wrap(err, "error initializing counter")
	}

	svc := &Service{
		logger:        logger.WithName(serviceName),
		itemDatabase:  db,
		encoder:       encoder,
		itemCounter:   itemCounter,
		userIDFetcher: userIDFetcher,
		itemIDFetcher: itemIDFetcher,
	}

	ctx := context.Background()
	itemCount, err := svc.itemDatabase.GetAllItemsCount(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "setting current item count")
	}
	svc.itemCounter.IncrementBy(ctx, itemCount)

	return svc, nil

}

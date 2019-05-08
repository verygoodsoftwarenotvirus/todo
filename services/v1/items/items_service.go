package items

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/pkg/errors"
)

const (
	// MiddlewareCtxKey is a string alias we can use for referring to item input data in contexts
	MiddlewareCtxKey models.ContextKey   = "item_input"
	counterName      metrics.CounterName = "items"
	topicName        string              = "items"
	serviceName      string              = "items_service"
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
		newsman       *newsman.Newsman
	}

	// UserIDFetcher is a function that fetches user IDs
	UserIDFetcher func(*http.Request) uint64

	// ItemIDFetcher is a function that fetches item IDs
	ItemIDFetcher func(*http.Request) uint64
)

// ProvideItemsService builds a new ItemsService
func ProvideItemsService(
	logger logging.Logger,
	db database.Database,
	userIDFetcher UserIDFetcher,
	itemIDFetcher ItemIDFetcher,
	encoder encoding.EncoderDecoder,
	itemCounterProvider metrics.UnitCounterProvider,
	newsman *newsman.Newsman,
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
		newsman:       newsman,
	}

	ctx := context.Background()
	itemCount, err := svc.itemDatabase.GetAllItemsCount(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "setting current item count")
	}
	svc.itemCounter.IncrementBy(ctx, itemCount)

	return svc, nil
}

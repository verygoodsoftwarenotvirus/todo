package items

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	// CreateMiddlewareCtxKey is a string alias we can use for referring to item input data in contexts.
	CreateMiddlewareCtxKey models.ContextKey = "item_create_input"
	// UpdateMiddlewareCtxKey is a string alias we can use for referring to item update data in contexts.
	UpdateMiddlewareCtxKey models.ContextKey = "item_update_input"

	counterName        metrics.CounterName = "items"
	counterDescription string              = "the number of items managed by the items service"
	topicName          string              = "items"
	serviceName        string              = "items_service"
)

var (
	_ models.ItemDataServer = (*Service)(nil)
)

type (
	// Service handles to-do list items
	Service struct {
		logger          logging.Logger
		itemDataManager models.ItemDataManager
		itemIDFetcher   ItemIDFetcher
		userIDFetcher   UserIDFetcher
		itemCounter     metrics.UnitCounter
		encoderDecoder  encoding.EncoderDecoder
		reporter        newsman.Reporter
	}

	// UserIDFetcher is a function that fetches user IDs.
	UserIDFetcher func(*http.Request) uint64

	// ItemIDFetcher is a function that fetches item IDs.
	ItemIDFetcher func(*http.Request) uint64
)

// ProvideItemsService builds a new ItemsService.
func ProvideItemsService(
	logger logging.Logger,
	itemDataManager models.ItemDataManager,
	itemIDFetcher ItemIDFetcher,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
	itemCounterProvider metrics.UnitCounterProvider,
	reporter newsman.Reporter,
) (*Service, error) {
	itemCounter, err := itemCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &Service{
		logger:          logger.WithName(serviceName),
		itemIDFetcher:   itemIDFetcher,
		userIDFetcher:   userIDFetcher,
		itemDataManager: itemDataManager,
		encoderDecoder:  encoder,
		itemCounter:     itemCounter,
		reporter:        reporter,
	}

	return svc, nil
}

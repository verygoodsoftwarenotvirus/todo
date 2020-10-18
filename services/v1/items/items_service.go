package items

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	// createMiddlewareCtxKey is a string alias we can use for referring to item input data in contexts.
	createMiddlewareCtxKey models.ContextKey = "item_create_input"
	// updateMiddlewareCtxKey is a string alias we can use for referring to item update data in contexts.
	updateMiddlewareCtxKey models.ContextKey = "item_update_input"

	counterName        metrics.CounterName = "items"
	counterDescription string              = "the number of items managed by the items service"
	topicName          string              = "items"
	serviceName        string              = "items_service"
)

var (
	_ models.ItemDataServer = (*Service)(nil)
)

type (
	// SearchIndex is a type alias for dependency injection's sake
	SearchIndex search.IndexManager

	// Service handles to-do list items
	Service struct {
		logger             logging.Logger
		itemDataManager    models.ItemDataManager
		itemIDFetcher      ItemIDFetcher
		sessionInfoFetcher SessionInfoFetcher
		itemCounter        metrics.UnitCounter
		encoderDecoder     encoding.EncoderDecoder
		reporter           newsman.Reporter
		search             SearchIndex
	}

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*models.SessionInfo, error)

	// ItemIDFetcher is a function that fetches item IDs.
	ItemIDFetcher func(*http.Request) uint64
)

// ProvideItemsService builds a new ItemsService.
func ProvideItemsService(
	logger logging.Logger,
	itemDataManager models.ItemDataManager,
	itemIDFetcher ItemIDFetcher,
	sessionInfoFetcher SessionInfoFetcher,
	encoder encoding.EncoderDecoder,
	itemCounterProvider metrics.UnitCounterProvider,
	reporter newsman.Reporter,
	searchIndexManager SearchIndex,
) (*Service, error) {
	itemCounter, err := itemCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &Service{
		logger:             logger.WithName(serviceName),
		itemIDFetcher:      itemIDFetcher,
		sessionInfoFetcher: sessionInfoFetcher,
		itemDataManager:    itemDataManager,
		encoderDecoder:     encoder,
		itemCounter:        itemCounter,
		reporter:           reporter,
		search:             searchIndexManager,
	}

	return svc, nil
}

// ProvideItemsServiceSearchIndex provides a search index for the service
func ProvideItemsServiceSearchIndex(
	searchSettings config.SearchSettings,
	indexProvider search.IndexManagerProvider,
	logger logging.Logger,
) (SearchIndex, error) {
	logger.WithValue("index_path", searchSettings.ItemsIndexPath).Debug("setting up items search index")

	searchIndex, indexInitErr := indexProvider(searchSettings.ItemsIndexPath, models.ItemsSearchIndexName, logger)
	if indexInitErr != nil {
		logger.Error(indexInitErr, "setting up items search index")
		return nil, indexInitErr
	}

	return searchIndex, nil
}

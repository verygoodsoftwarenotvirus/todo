package items

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	// createMiddlewareCtxKey is a string alias we can use for referring to item input data in contexts.
	createMiddlewareCtxKey types.ContextKey = "item_create_input"
	// updateMiddlewareCtxKey is a string alias we can use for referring to item update data in contexts.
	updateMiddlewareCtxKey types.ContextKey = "item_update_input"

	counterName        metrics.CounterName = "items"
	counterDescription string              = "the number of items managed by the items service"
	serviceName        string              = "items_service"
)

var _ types.ItemDataServer = (*Service)(nil)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

	// Service handles to-do list items.
	Service struct {
		logger             logging.Logger
		itemDataManager    types.ItemDataManager
		auditLog           types.ItemAuditManager
		itemIDFetcher      ItemIDFetcher
		sessionInfoFetcher SessionInfoFetcher
		itemCounter        metrics.UnitCounter
		encoderDecoder     encoding.EncoderDecoder
		search             SearchIndex
	}

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)

	// ItemIDFetcher is a function that fetches item IDs.
	ItemIDFetcher func(*http.Request) uint64
)

// ProvideItemsService builds a new ItemsService.
func ProvideItemsService(
	logger logging.Logger,
	itemDataManager types.ItemDataManager,
	auditLog types.ItemAuditManager,
	itemIDFetcher ItemIDFetcher,
	sessionInfoFetcher SessionInfoFetcher,
	encoder encoding.EncoderDecoder,
	itemCounterProvider metrics.UnitCounterProvider,
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
		auditLog:           auditLog,
		encoderDecoder:     encoder,
		itemCounter:        itemCounter,
		search:             searchIndexManager,
	}

	return svc, nil
}

// ProvideItemsServiceSearchIndex provides a search index for the service.
func ProvideItemsServiceSearchIndex(
	searchSettings config.SearchSettings,
	indexProvider search.IndexManagerProvider,
	logger logging.Logger,
) (SearchIndex, error) {
	logger.WithValue("index_path", searchSettings.ItemsIndexPath).Debug("setting up items search index")

	searchIndex, indexInitErr := indexProvider(searchSettings.ItemsIndexPath, types.ItemsSearchIndexName, logger)
	if indexInitErr != nil {
		logger.Error(indexInitErr, "setting up items search index")
		return nil, indexInitErr
	}

	return searchIndex, nil
}

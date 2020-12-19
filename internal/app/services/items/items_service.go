package items

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	counterName        metrics.CounterName = "items"
	counterDescription string              = "the number of items managed by the items service"
	serviceName        string              = "items_service"
)

var _ types.ItemDataService = (*service)(nil)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

	// service handles to-do list items.
	service struct {
		logger             logging.Logger
		itemDataManager    types.ItemDataManager
		auditLog           types.ItemAuditManager
		itemIDFetcher      func(*http.Request) uint64
		sessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)
		itemCounter        metrics.UnitCounter
		encoderDecoder     encoding.EncoderDecoder
		tracer             tracing.Tracer
		search             SearchIndex
	}

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)

	// ItemIDFetcher is a function that fetches item IDs.
	ItemIDFetcher func(*http.Request) uint64
)

// ProvideService builds a new ItemsService.
func ProvideService(
	logger logging.Logger,
	itemDataManager types.ItemDataManager,
	auditLog types.ItemAuditManager,
	encoder encoding.EncoderDecoder,
	itemCounterProvider metrics.UnitCounterProvider,
	searchSettings search.Config,
	indexProvider search.IndexManagerProvider,
) (types.ItemDataService, error) {
	itemCounter, err := itemCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	logger.WithValue("index_path", searchSettings.ItemsIndexPath).Debug("setting up items search index")

	searchIndexManager, indexInitErr := indexProvider(searchSettings.ItemsIndexPath, types.ItemsSearchIndexName, logger)
	if indexInitErr != nil {
		logger.Error(indexInitErr, "setting up items search index")
		return nil, indexInitErr
	}

	svc := &service{
		logger:             logger.WithName(serviceName),
		itemIDFetcher:      routeparams.BuildRouteParamIDFetcher(logger, ItemIDURIParamKey, "item"),
		sessionInfoFetcher: routeparams.SessionInfoFetcherFromRequestContext,
		itemDataManager:    itemDataManager,
		auditLog:           auditLog,
		encoderDecoder:     encoder,
		itemCounter:        itemCounter,
		search:             searchIndexManager,
		tracer:             tracing.NewTracer(serviceName),
	}

	return svc, nil
}

package items

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
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
		logger                    logging.Logger
		itemDataManager           types.ItemDataManager
		itemIDFetcher             func(*http.Request) uint64
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		itemCounter               metrics.UnitCounter
		encoderDecoder            encoding.ServerEncoderDecoder
		tracer                    tracing.Tracer
		search                    SearchIndex
	}
)

// ProvideService builds a new ItemsService.
func ProvideService(
	logger logging.Logger,
	itemDataManager types.ItemDataManager,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	searchSettings search.Config,
	indexProvider search.IndexManagerProvider,
	routeParamManager routing.RouteParamManager,
) (types.ItemDataService, error) {
	logger.WithValue("index_path", searchSettings.ItemsIndexPath).Debug("setting up items search index")

	searchIndexManager, indexInitErr := indexProvider(searchSettings.ItemsIndexPath, types.ItemsSearchIndexName, logger)
	if indexInitErr != nil {
		logger.Error(indexInitErr, "setting up items search index")
		return nil, indexInitErr
	}

	svc := &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		itemIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, ItemIDURIParamKey, "item"),
		sessionContextDataFetcher: routeParamManager.FetchContextFromRequest,
		itemDataManager:           itemDataManager,
		encoderDecoder:            encoder,
		itemCounter:               metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		search:                    searchIndexManager,
		tracer:                    tracing.NewTracer(serviceName),
	}

	return svc, nil
}

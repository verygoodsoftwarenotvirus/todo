package items

import (
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"log"
	"net/http"
	"time"
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

	// service handles items.
	service struct {
		logger                    logging.Logger
		itemDataManager           types.ItemDataManager
		itemIDFetcher             func(*http.Request) uint64
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		itemCounter               metrics.UnitCounter
		encoderDecoder            encoding.ServerEncoderDecoder
		tracer                    tracing.Tracer
		pendingWritesProducer     events.Producer
		search                    SearchIndex
	}
)

// ProvideService builds a new ItemsService.
func ProvideService(
	logger logging.Logger,
	cfg Config,
	itemDataManager types.ItemDataManager,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	searchIndexProvider search.IndexManagerProvider,
	routeParamManager routing.RouteParamManager,
	producerProvider events.ProducerProvider,
) (types.ItemDataService, error) {
	searchIndexManager, err := searchIndexProvider(search.IndexPath(cfg.SearchIndexPath), "items", logger)
	if err != nil {
		return nil, fmt.Errorf("setting up search index: %w", err)
	}

	// TODO: put topic in config
	pendingWritesProducer, err := producerProvider.ProviderProducer("pending_writes")
	if err != nil {
		return nil, fmt.Errorf("setting up event producer: %w", err)
	}

	svc := &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		itemIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, ItemIDURIParamKey, serviceName),
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		itemDataManager:           itemDataManager,
		pendingWritesProducer:     pendingWritesProducer,
		encoderDecoder:            encoder,
		itemCounter:               metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		search:                    searchIndexManager,
		tracer:                    tracing.NewTracer(serviceName),
	}

	go func() {
		for range time.Tick(time.Second) {
			if err = pendingWritesProducer.Publish([]byte(`{"things": "stuff"}`)); err != nil {
				log.Fatal(err)
			}
		}
	}()

	return svc, nil
}

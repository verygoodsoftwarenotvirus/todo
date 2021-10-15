package items

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	serviceName string = "items_service"
)

var _ types.ItemDataService = (*service)(nil)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

	// service handles items.
	service struct {
		logger                    logging.Logger
		itemDataManager           types.ItemDataManager
		itemIDFetcher             func(*http.Request) string
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		preWritesPublisher        publishers.Publisher
		preUpdatesPublisher       publishers.Publisher
		preArchivesPublisher      publishers.Publisher
		encoderDecoder            encoding.ServerEncoderDecoder
		tracer                    tracing.Tracer
		search                    SearchIndex
		async                     bool
	}
)

// ProvideService builds a new ItemsService.
func ProvideService(
	ctx context.Context,
	logger logging.Logger,
	cfg *Config,
	itemDataManager types.ItemDataManager,
	encoder encoding.ServerEncoderDecoder,
	searchIndexProvider search.IndexManagerProvider,
	routeParamManager routing.RouteParamManager,
	publisherProvider publishers.PublisherProvider,
) (types.ItemDataService, error) {
	client := &http.Client{Transport: tracing.BuildTracedHTTPTransport(time.Second)}
	searchIndexManager, err := searchIndexProvider(ctx, logger, client, search.IndexPath(cfg.SearchIndexPath), "items", "name", "details")
	if err != nil {
		return nil, fmt.Errorf("setting up search index: %w", err)
	}

	preWritesPublisher, err := publisherProvider.ProviderPublisher(cfg.PreWritesTopicName)
	if err != nil {
		return nil, fmt.Errorf("setting up queue provider: %w", err)
	}

	preUpdatesPublisher, err := publisherProvider.ProviderPublisher(cfg.PreUpdatesTopicName)
	if err != nil {
		return nil, fmt.Errorf("setting up queue provider: %w", err)
	}

	preArchivesPublisher, err := publisherProvider.ProviderPublisher(cfg.PreArchivesTopicName)
	if err != nil {
		return nil, fmt.Errorf("setting up queue provider: %w", err)
	}

	svc := &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		itemIDFetcher:             routeParamManager.BuildRouteParamStringIDFetcher(ItemIDURIParamKey),
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		itemDataManager:           itemDataManager,
		preWritesPublisher:        preWritesPublisher,
		preUpdatesPublisher:       preUpdatesPublisher,
		preArchivesPublisher:      preArchivesPublisher,
		encoderDecoder:            encoder,
		async:                     cfg.Async,
		search:                    searchIndexManager,
		tracer:                    tracing.NewTracer(serviceName),
	}

	return svc, nil
}

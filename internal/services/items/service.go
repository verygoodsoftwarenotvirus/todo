package items

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
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

	// Config configures the service.
	Config struct {
		Logger          logging.Config `json:"logging" mapstructure:"logging" toml:"logging,omitempty"`
		SearchIndexPath string         `json:"search_index_path" mapstructure:"search_index_path" toml:"search_index_path,omitempty"`
	}

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

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.SearchIndexPath, validation.Required),
	)
}

// ProvideService builds a new ItemsService.
func ProvideService(
	logger logging.Logger,
	cfg Config,
	itemDataManager types.ItemDataManager,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	indexProvider search.IndexManagerProvider,
	routeParamManager routing.RouteParamManager,
) (types.ItemDataService, error) {
	logger.WithValue("index_path", cfg.SearchIndexPath).Debug("setting up items search index")

	searchIndexManager, indexInitErr := indexProvider(search.IndexPath(cfg.SearchIndexPath), types.ItemsSearchIndexName, logger)
	if indexInitErr != nil {
		logger.Error(indexInitErr, "setting up items search index")
		return nil, indexInitErr
	}

	svc := &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		itemIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, ItemIDURIParamKey, "item"),
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		itemDataManager:           itemDataManager,
		encoderDecoder:            encoder,
		itemCounter:               metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		search:                    searchIndexManager,
		tracer:                    tracing.NewTracer(serviceName),
	}

	return svc, nil
}

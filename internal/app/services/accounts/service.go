package accounts

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	counterName        metrics.CounterName = "accounts"
	counterDescription string              = "the number of accounts managed by the accounts service"
	serviceName        string              = "accounts_service"
)

var _ types.AccountDataService = (*service)(nil)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

	// service handles to-do list accounts.
	service struct {
		logger                       logging.Logger
		accountDataManager           types.AccountDataManager
		accountMembershipDataManager types.AccountUserMembershipDataManager
		accountIDFetcher             func(*http.Request) uint64
		userIDFetcher                func(*http.Request) uint64
		requestContextFetcher        func(*http.Request) (*types.RequestContext, error)
		accountCounter               metrics.UnitCounter
		encoderDecoder               encoding.HTTPResponseEncoder
		tracer                       tracing.Tracer
	}
)

// ProvideService builds a new AccountsService.
func ProvideService(
	logger logging.Logger,
	accountDataManager types.AccountDataManager,
	accountMembershipDataManager types.AccountUserMembershipDataManager,
	encoder encoding.HTTPResponseEncoder,
	counterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) types.AccountDataService {
	return &service{
		logger:                       logging.EnsureLogger(logger).WithName(serviceName),
		accountIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, AccountIDURIParamKey, "account"),
		userIDFetcher:                routeParamManager.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		requestContextFetcher:        routeParamManager.FetchContextFromRequest,
		accountDataManager:           accountDataManager,
		accountMembershipDataManager: accountMembershipDataManager,
		encoderDecoder:               encoder,
		accountCounter:               metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		tracer:                       tracing.NewTracer(serviceName),
	}
}

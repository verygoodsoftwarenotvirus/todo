package accounts

import (
	"net/http"

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
		sessionContextDataFetcher    func(*http.Request) (*types.SessionContextData, error)
		accountCounter               metrics.UnitCounter
		encoderDecoder               encoding.ServerEncoderDecoder
		tracer                       tracing.Tracer
	}
)

// ProvideService builds a new AccountsService.
func ProvideService(
	logger logging.Logger,
	accountDataManager types.AccountDataManager,
	accountMembershipDataManager types.AccountUserMembershipDataManager,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) types.AccountDataService {
	return &service{
		logger:                       logging.EnsureLogger(logger).WithName(serviceName),
		accountIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, AccountIDURIParamKey, "account"),
		userIDFetcher:                routeParamManager.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		sessionContextDataFetcher:    authservice.FetchContextFromRequest,
		accountDataManager:           accountDataManager,
		accountMembershipDataManager: accountMembershipDataManager,
		encoderDecoder:               encoder,
		accountCounter:               metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		tracer:                       tracing.NewTracer(serviceName),
	}
}

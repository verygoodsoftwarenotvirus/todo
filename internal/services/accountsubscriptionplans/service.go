package accountsubscriptionplans

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	counterName        metrics.CounterName = "account_subscription_plans"
	counterDescription string              = "the number of account subscription plans managed by the account subscription plans service"
	serviceName        string              = "account_subscription_plans_service"
)

var _ types.AccountSubscriptionPlanDataService = (*service)(nil)

type (
	// service handles to-do list account subscription plans.
	service struct {
		logger                             logging.Logger
		accountSubscriptionPlanDataManager types.AccountSubscriptionPlanDataManager
		accountSubscriptionPlanIDFetcher   func(*http.Request) uint64
		sessionContextDataFetcher          func(*http.Request) (*types.SessionContextData, error)
		planCounter                        metrics.UnitCounter
		encoderDecoder                     encoding.ServerEncoderDecoder
		tracer                             tracing.Tracer
	}
)

// ProvideService builds a new PlansService.
func ProvideService(
	logger logging.Logger,
	accountSubscriptionPlanDataManager types.AccountSubscriptionPlanDataManager,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) types.AccountSubscriptionPlanDataService {
	return &service{
		logger:                             logging.EnsureLogger(logger).WithName(serviceName),
		accountSubscriptionPlanIDFetcher:   routeParamManager.BuildRouteParamIDFetcher(logger, AccountSubscriptionPlanIDURIParamKey, "account subscription plan"),
		sessionContextDataFetcher:          routeParamManager.FetchContextFromRequest,
		accountSubscriptionPlanDataManager: accountSubscriptionPlanDataManager,
		encoderDecoder:                     encoder,
		planCounter:                        metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		tracer:                             tracing.NewTracer(serviceName),
	}
}

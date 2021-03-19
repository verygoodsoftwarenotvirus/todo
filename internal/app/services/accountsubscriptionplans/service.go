package accountsubscriptionplans

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	counterName        metrics.CounterName = "accountsubscriptionplans"
	counterDescription string              = "the number of accountsubscriptionplans managed by the accountsubscriptionplans service"
	serviceName        string              = "plans_service"
)

var _ types.AccountSubscriptionPlanDataService = (*service)(nil)

type (
	// service handles to-do list account subscription plans.
	service struct {
		logger                logging.Logger
		planDataManager       types.AccountSubscriptionPlanDataManager
		planIDFetcher         func(*http.Request) uint64
		requestContextFetcher func(*http.Request) (*types.RequestContext, error)
		planCounter           metrics.UnitCounter
		encoderDecoder        encoding.ServerEncoderDecoder
		tracer                tracing.Tracer
	}
)

// ProvideService builds a new PlansService.
func ProvideService(
	logger logging.Logger,
	planDataManager types.AccountSubscriptionPlanDataManager,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) types.AccountSubscriptionPlanDataService {
	return &service{
		logger:                logging.EnsureLogger(logger).WithName(serviceName),
		planIDFetcher:         routeParamManager.BuildRouteParamIDFetcher(logger, AccountSubscriptionPlanIDURIParamKey, "account subscription plan"),
		requestContextFetcher: routeParamManager.FetchContextFromRequest,
		planDataManager:       planDataManager,
		encoderDecoder:        encoder,
		planCounter:           metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		tracer:                tracing.NewTracer(serviceName),
	}
}

package accountsubscriptionplans

import (
	"fmt"
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
		logger             logging.Logger
		planDataManager    types.AccountSubscriptionPlanDataManager
		planIDFetcher      func(*http.Request) uint64
		sessionInfoFetcher func(*http.Request) (*types.RequestContext, error)
		planCounter        metrics.UnitCounter
		encoderDecoder     encoding.HTTPResponseEncoder
		tracer             tracing.Tracer
	}
)

// ProvideService builds a new PlansService.
func ProvideService(
	logger logging.Logger,
	planDataManager types.AccountSubscriptionPlanDataManager,
	encoder encoding.HTTPResponseEncoder,
	planCounterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) (types.AccountSubscriptionPlanDataService, error) {
	planCounter, err := planCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("initializing counter: %w", err)
	}

	svc := &service{
		logger:             logging.EnsureLogger(logger).WithName(serviceName),
		planIDFetcher:      routeParamManager.BuildRouteParamIDFetcher(logger, AccountSubscriptionPlanIDURIParamKey, "account subscription plan"),
		sessionInfoFetcher: routeParamManager.FetchContextFromRequest,
		planDataManager:    planDataManager,
		encoderDecoder:     encoder,
		planCounter:        planCounter,
		tracer:             tracing.NewTracer(serviceName),
	}

	return svc, nil
}

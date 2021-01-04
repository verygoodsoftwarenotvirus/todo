package plans

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	counterName        metrics.CounterName = "plans"
	counterDescription string              = "the number of plans managed by the plans service"
	serviceName        string              = "plans_service"
)

var _ types.AccountSubscriptionPlanDataService = (*service)(nil)

type (
	// service handles to-do list plans.
	service struct {
		logger             logging.Logger
		planDataManager    types.AccountSubscriptionPlanDataManager
		auditLog           types.AccountSubscriptionPlanAuditManager
		planIDFetcher      func(*http.Request) uint64
		sessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)
		planCounter        metrics.UnitCounter
		encoderDecoder     encoding.EncoderDecoder
		tracer             tracing.Tracer
	}
)

// ProvideService builds a new PlansService.
func ProvideService(
	logger logging.Logger,
	planDataManager types.AccountSubscriptionPlanDataManager,
	auditLog types.AccountSubscriptionPlanAuditManager,
	encoder encoding.EncoderDecoder,
	planCounterProvider metrics.UnitCounterProvider,
) (types.AccountSubscriptionPlanDataService, error) {
	planCounter, err := planCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &service{
		logger:             logger.WithName(serviceName),
		planIDFetcher:      routeparams.BuildRouteParamIDFetcher(logger, PlanIDURIParamKey, "plan"),
		sessionInfoFetcher: routeparams.SessionInfoFetcherFromRequestContext,
		planDataManager:    planDataManager,
		auditLog:           auditLog,
		encoderDecoder:     encoder,
		planCounter:        planCounter,
		tracer:             tracing.NewTracer(serviceName),
	}

	return svc, nil
}

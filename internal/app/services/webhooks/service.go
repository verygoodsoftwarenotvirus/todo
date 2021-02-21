package webhooks

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
	// createMiddlewareCtxKey is a string alias we can use for referring to webhook input data in contexts.
	createMiddlewareCtxKey types.ContextKey = "webhook_create_input"
	// updateMiddlewareCtxKey is a string alias we can use for referring to webhook input data in contexts.
	updateMiddlewareCtxKey types.ContextKey = "webhook_update_input"

	counterName        metrics.CounterName = "webhooks"
	counterDescription string              = "the number of webhooks managed by the webhooks service"
	serviceName        string              = "webhooks_service"
)

var (
	_ types.WebhookDataService = (*service)(nil)
)

type (
	// service handles webhooks.
	service struct {
		logger             logging.Logger
		webhookCounter     metrics.UnitCounter
		webhookDataManager types.WebhookDataManager
		sessionInfoFetcher func(*http.Request) (*types.RequestContext, error)
		webhookIDFetcher   func(*http.Request) uint64
		encoderDecoder     encoding.HTTPResponseEncoder
		tracer             tracing.Tracer
	}
)

// ProvideWebhooksService builds a new WebhooksService.
func ProvideWebhooksService(
	logger logging.Logger,
	webhookDataManager types.WebhookDataManager,
	encoder encoding.HTTPResponseEncoder,
	webhookCounterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) (types.WebhookDataService, error) {
	webhookCounter, err := webhookCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("initializing counter: %w", err)
	}

	svc := &service{
		logger:             logging.EnsureLogger(logger).WithName(serviceName),
		webhookDataManager: webhookDataManager,
		encoderDecoder:     encoder,
		webhookCounter:     webhookCounter,
		sessionInfoFetcher: routeParamManager.FetchContextFromRequest,
		webhookIDFetcher:   routeParamManager.BuildRouteParamIDFetcher(logger, WebhookIDURIParamKey, "webhook"),
		tracer:             tracing.NewTracer(serviceName),
	}

	return svc, nil
}

package webhooks

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
		logger                    logging.Logger
		webhookCounter            metrics.UnitCounter
		webhookDataManager        types.WebhookDataManager
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		webhookIDFetcher          func(*http.Request) uint64
		encoderDecoder            encoding.ServerEncoderDecoder
		tracer                    tracing.Tracer
	}
)

// ProvideWebhooksService builds a new WebhooksService.
func ProvideWebhooksService(
	logger logging.Logger,
	webhookDataManager types.WebhookDataManager,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) types.WebhookDataService {
	return &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		webhookDataManager:        webhookDataManager,
		encoderDecoder:            encoder,
		webhookCounter:            metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		sessionContextDataFetcher: routeParamManager.FetchContextFromRequest,
		webhookIDFetcher:          routeParamManager.BuildRouteParamIDFetcher(logger, WebhookIDURIParamKey, "webhook"),
		tracer:                    tracing.NewTracer(serviceName),
	}
}
package webhooks

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
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
		auditLog           types.WebhookAuditManager
		sessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)
		webhookIDFetcher   func(*http.Request) uint64
		encoderDecoder     encoding.EncoderDecoder
	}
)

// ProvideWebhooksService builds a new WebhooksService.
func ProvideWebhooksService(
	logger logging.Logger,
	webhookDataManager types.WebhookDataManager,
	auditLog types.WebhookAuditManager,
	encoder encoding.EncoderDecoder,
	webhookCounterProvider metrics.UnitCounterProvider,
) (types.WebhookDataService, error) {
	webhookCounter, err := webhookCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &service{
		logger:             logger.WithName(serviceName),
		webhookDataManager: webhookDataManager,
		auditLog:           auditLog,
		encoderDecoder:     encoder,
		webhookCounter:     webhookCounter,
		sessionInfoFetcher: routeparams.SessionInfoFetcherFromRequestContext,
		webhookIDFetcher:   routeparams.BuildRouteParamIDFetcher(logger, WebhookIDURIParamKey, "webhook"),
	}

	return svc, nil
}

package webhooks

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	// createMiddlewareCtxKey is a string alias we can use for referring to webhook input data in contexts.
	createMiddlewareCtxKey models.ContextKey = "webhook_create_input"
	// updateMiddlewareCtxKey is a string alias we can use for referring to webhook input data in contexts.
	updateMiddlewareCtxKey models.ContextKey = "webhook_update_input"

	counterName        metrics.CounterName = "webhooks"
	counterDescription string              = "the number of webhooks managed by the webhooks service"
	serviceName        string              = "webhooks_service"
)

var (
	_ models.WebhookDataServer = (*Service)(nil)
)

type (
	// Service handles TODO ListHandler webhooks.
	Service struct {
		logger             logging.Logger
		webhookCounter     metrics.UnitCounter
		webhookDataManager models.WebhookDataManager
		auditLog           models.WebhookAuditManager
		userIDFetcher      UserIDFetcher
		webhookIDFetcher   WebhookIDFetcher
		encoderDecoder     encoding.EncoderDecoder
	}

	// UserIDFetcher is a function that fetches user IDs.
	UserIDFetcher func(*http.Request) uint64

	// WebhookIDFetcher is a function that fetches webhook IDs.
	WebhookIDFetcher func(*http.Request) uint64
)

// ProvideWebhooksService builds a new WebhooksService.
func ProvideWebhooksService(
	logger logging.Logger,
	webhookDataManager models.WebhookDataManager,
	auditLog models.WebhookAuditManager,
	userIDFetcher UserIDFetcher,
	webhookIDFetcher WebhookIDFetcher,
	encoder encoding.EncoderDecoder,
	webhookCounterProvider metrics.UnitCounterProvider,
) (*Service, error) {
	webhookCounter, err := webhookCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &Service{
		logger:             logger.WithName(serviceName),
		webhookDataManager: webhookDataManager,
		auditLog:           auditLog,
		encoderDecoder:     encoder,
		webhookCounter:     webhookCounter,
		userIDFetcher:      userIDFetcher,
		webhookIDFetcher:   webhookIDFetcher,
	}

	return svc, nil
}

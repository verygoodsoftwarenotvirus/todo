package webhooks

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

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
		sessionInfoFetcher SessionInfoFetcher
		webhookIDFetcher   WebhookIDFetcher
		encoderDecoder     encoding.EncoderDecoder
	}

	// UserIDFetcher is a function that fetches user IDs.
	UserIDFetcher func(*http.Request) uint64

	// WebhookIDFetcher is a function that fetches webhook IDs.
	WebhookIDFetcher func(*http.Request) uint64

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*models.SessionInfo, error)
)

// ProvideWebhooksService builds a new WebhooksService.
func ProvideWebhooksService(
	logger logging.Logger,
	webhookDataManager models.WebhookDataManager,
	auditLog models.WebhookAuditManager,
	sessionInfoFetcher SessionInfoFetcher,
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
		sessionInfoFetcher: sessionInfoFetcher,
		webhookIDFetcher:   webhookIDFetcher,
	}

	return svc, nil
}

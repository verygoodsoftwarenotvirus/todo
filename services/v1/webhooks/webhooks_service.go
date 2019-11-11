package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	// CreateMiddlewareCtxKey is a string alias we can use for referring to webhook input data in contexts
	CreateMiddlewareCtxKey models.ContextKey = "webhook_create_input"
	// UpdateMiddlewareCtxKey is a string alias we can use for referring to webhook input data in contexts
	UpdateMiddlewareCtxKey models.ContextKey = "webhook_update_input"

	counterName metrics.CounterName = "webhooks"
	topicName   string              = "webhooks"
	serviceName string              = "webhooks_service"
)

var (
	_ models.WebhookDataServer = (*Service)(nil)
)

type (
	eventManager interface {
		newsman.Reporter

		TuneIn(newsman.Listener)
	}

	// Service handles TODO ListHandler webhooks
	Service struct {
		logger           logging.Logger
		webhookCounter   metrics.UnitCounter
		webhookDatabase  models.WebhookDataManager
		userIDFetcher    UserIDFetcher
		webhookIDFetcher WebhookIDFetcher
		encoderDecoder   encoding.EncoderDecoder
		eventManager     eventManager
	}

	// UserIDFetcher is a function that fetches user IDs
	UserIDFetcher func(*http.Request) uint64

	// WebhookIDFetcher is a function that fetches webhook IDs
	WebhookIDFetcher func(*http.Request) uint64
)

// ProvideWebhooksService builds a new WebhooksService
func ProvideWebhooksService(
	ctx context.Context,
	logger logging.Logger,
	webhookDatabase models.WebhookDataManager,
	userIDFetcher UserIDFetcher,
	webhookIDFetcher WebhookIDFetcher,
	encoder encoding.EncoderDecoder,
	webhookCounterProvider metrics.UnitCounterProvider,
	em *newsman.Newsman,
) (*Service, error) {
	webhookCounter, err := webhookCounterProvider(counterName, "the number of webhooks managed by the webhooks service")
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &Service{
		logger:           logger.WithName(serviceName),
		webhookDatabase:  webhookDatabase,
		encoderDecoder:   encoder,
		webhookCounter:   webhookCounter,
		userIDFetcher:    userIDFetcher,
		webhookIDFetcher: webhookIDFetcher,
		eventManager:     em,
	}

	webhookCount, err := svc.webhookDatabase.GetAllWebhooksCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("setting current webhook count: %w", err)
	}
	svc.webhookCounter.IncrementBy(ctx, webhookCount)

	return svc, nil
}

package webhooks

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/pkg/errors"
)

const (
	// MiddlewareCtxKey is a string alias we can use for referring to webhook input data in contexts
	MiddlewareCtxKey models.ContextKey   = "webhook_input"
	counterName      metrics.CounterName = "webhooks"
	topicName        string              = "webhooks"
	serviceName      string              = "webhooks_service"
)

type (
	eventManager interface {
		newsman.Reporter

		TuneIn(newsman.Listener)
	}

	// Service handles TODO List webhooks
	Service struct {
		logger           logging.Logger
		webhookCounter   metrics.UnitCounter
		webhookDatabase  models.WebhookDataManager
		userIDFetcher    UserIDFetcher
		webhookIDFetcher WebhookIDFetcher
		encoderDecoder   encoding.EncoderDecoder
		newsman          eventManager
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
	newsman *newsman.Newsman,
) (*Service, error) {
	webhookCounter, err := webhookCounterProvider(counterName, "the number of webhooks managed by the webhooks service")
	if err != nil {
		return nil, errors.Wrap(err, "error initializing counter")
	}

	svc := &Service{
		logger:           logger.WithName(serviceName),
		webhookDatabase:  webhookDatabase,
		encoderDecoder:   encoder,
		webhookCounter:   webhookCounter,
		userIDFetcher:    userIDFetcher,
		webhookIDFetcher: webhookIDFetcher,
		newsman:          newsman,
	}

	webhookCount, err := svc.webhookDatabase.GetAllWebhooksCount(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "setting current webhook count")
	}
	svc.webhookCounter.IncrementBy(ctx, webhookCount)

	return svc, nil

}

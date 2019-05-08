package webhooks

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

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
	// Service handles TODO List webhooks
	Service struct {
		logger           logging.Logger
		webhookCounter   metrics.UnitCounter
		webhookDatabase  models.WebhookDataManager
		userIDFetcher    UserIDFetcher
		webhookIDFetcher WebhookIDFetcher
		encoder          encoding.EncoderDecoder
		newsman          *newsman.Newsman
	}

	// UserIDFetcher is a function that fetches user IDs
	UserIDFetcher func(*http.Request) uint64

	// WebhookIDFetcher is a function that fetches webhook IDs
	WebhookIDFetcher func(*http.Request) uint64
)

// ProvideWebhooksService builds a new WebhooksService
func ProvideWebhooksService(
	logger logging.Logger,
	db database.Database,
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
		webhookDatabase:  db,
		encoder:          encoder,
		webhookCounter:   webhookCounter,
		userIDFetcher:    userIDFetcher,
		webhookIDFetcher: webhookIDFetcher,
		newsman:          newsman,
	}

	ctx := context.Background()
	webhookCount, err := svc.webhookDatabase.GetAllWebhooksCount(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "setting current webhook count")
	}
	svc.webhookCounter.IncrementBy(ctx, webhookCount)

	return svc, nil

}

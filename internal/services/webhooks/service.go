package webhooks

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	serviceName string = "webhooks_service"
)

var (
	_ types.WebhookDataService = (*service)(nil)
)

type (
	// service handles webhooks.
	service struct {
		logger                    logging.Logger
		webhookDataManager        types.WebhookDataManager
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		webhookIDFetcher          func(*http.Request) string
		encoderDecoder            encoding.ServerEncoderDecoder
		preWritesProducer         publishers.Publisher
		preArchivesProducer       publishers.Publisher
		tracer                    tracing.Tracer
	}
)

// ProvideWebhooksService builds a new WebhooksService.
func ProvideWebhooksService(
	logger logging.Logger,
	cfg *Config,
	webhookDataManager types.WebhookDataManager,
	encoder encoding.ServerEncoderDecoder,
	routeParamManager routing.RouteParamManager,
	producerProvider publishers.PublisherProvider,
) (types.WebhookDataService, error) {
	preWritesProducer, err := producerProvider.ProviderPublisher(cfg.PreWritesTopicName)
	if err != nil {
		return nil, fmt.Errorf("setting up event producer: %w", err)
	}

	preArchivesProducer, err := producerProvider.ProviderPublisher(cfg.PreArchivesTopicName)
	if err != nil {
		return nil, fmt.Errorf("setting up event producer: %w", err)
	}

	s := &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		webhookDataManager:        webhookDataManager,
		encoderDecoder:            encoder,
		preWritesProducer:         preWritesProducer,
		preArchivesProducer:       preArchivesProducer,
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		webhookIDFetcher:          routeParamManager.BuildRouteParamStringIDFetcher(WebhookIDURIParamKey),
		tracer:                    tracing.NewTracer(serviceName),
	}

	return s, nil
}

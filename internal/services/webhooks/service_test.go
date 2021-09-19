package webhooks

import (
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService() *service {
	return &service{
		logger:             logging.NewNoopLogger(),
		webhookDataManager: &mocktypes.WebhookDataManager{},
		webhookIDFetcher:   func(req *http.Request) string { return "" },
		encoderDecoder:     mockencoding.NewMockEncoderDecoder(),
		tracer:             tracing.NewTracer("test"),
	}
}

func TestProvideWebhooksService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamStringIDFetcher",
			WebhookIDURIParamKey,
		).Return(func(*http.Request) string { return "" })

		cfg := &Config{}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return(&publishers.MockProducer{}, nil)

		actual, err := ProvideWebhooksService(
			logging.NewNoopLogger(),
			cfg,
			&mocktypes.WebhookDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			rpm,
			pp,
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}

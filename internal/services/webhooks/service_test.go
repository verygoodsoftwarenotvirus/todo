package webhooks

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
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

		cfg := &Config{
			PreWritesTopicName:   "pre-writes",
			PreArchivesTopicName: "pre-archives",
		}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return(&publishers.MockProducer{}, nil)
		pp.On("ProviderPublisher", cfg.PreArchivesTopicName).Return(&publishers.MockProducer{}, nil)

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

		mock.AssertExpectationsForObjects(t, rpm, pp)
	})

	T.Run("with error providing pre-writes publisher", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			PreWritesTopicName:   "pre-writes",
			PreArchivesTopicName: "pre-archives",
		}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return((*publishers.MockProducer)(nil), errors.New("blah"))

		actual, err := ProvideWebhooksService(
			logging.NewNoopLogger(),
			cfg,
			&mocktypes.WebhookDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			nil,
			pp,
		)

		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, pp)
	})

	T.Run("with error providing pre-archives publisher", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			PreWritesTopicName:   "pre-writes",
			PreArchivesTopicName: "pre-archives",
		}

		pp := &publishers.MockProducerProvider{}
		pp.On("ProviderPublisher", cfg.PreWritesTopicName).Return(&publishers.MockProducer{}, nil)
		pp.On("ProviderPublisher", cfg.PreArchivesTopicName).Return((*publishers.MockProducer)(nil), errors.New("blah"))

		actual, err := ProvideWebhooksService(
			logging.NewNoopLogger(),
			cfg,
			&mocktypes.WebhookDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			nil,
			pp,
		)

		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, pp)
	})
}

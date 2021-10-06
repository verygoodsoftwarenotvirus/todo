package websockets

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mock2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/consumers/mock"

	"github.com/stretchr/testify/mock"

	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

func buildTestService() *service {
	return &service{
		cookieName:     "testing",
		logger:         logging.NewNoopLogger(),
		encoderDecoder: mockencoding.NewMockEncoderDecoder(),
		tracer:         tracing.NewTracer("test"),
		connections:    map[string][]websocketConnection{},
	}
}

func TestProvideService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		authCfg := &authservice.Config{}
		logger := logging.NewNoopLogger()
		encoder := encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON)

		consumer := &mock2.Consumer{}
		consumer.On("Consume", chan bool(nil), chan error(nil))

		consumerProvider := &mock2.ConsumerProvider{}
		consumerProvider.On(
			"ProviderConsumer",
			testutils.ContextMatcher,
			dataChangesTopicName,
			mock.Anything,
		).Return(consumer, nil)

		actual, err := ProvideService(
			ctx,
			authCfg,
			logger,
			encoder,
			consumerProvider,
		)

		require.NoError(t, err)
		require.NotNil(t, actual)

		mock.AssertExpectationsForObjects(t, consumerProvider)
	})

	T.Run("with consumer provider error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		authCfg := &authservice.Config{}
		logger := logging.NewNoopLogger()
		encoder := encoding.ProvideServerEncoderDecoder(logger, encoding.ContentTypeJSON)

		consumerProvider := &mock2.ConsumerProvider{}
		consumerProvider.On(
			"ProviderConsumer",
			testutils.ContextMatcher,
			dataChangesTopicName,
			mock.Anything,
		).Return(&mock2.Consumer{}, errors.New("blah"))

		actual, err := ProvideService(
			ctx,
			authCfg,
			logger,
			encoder,
			consumerProvider,
		)

		require.Error(t, err)
		require.Nil(t, actual)
	})
}

func Test_buildWebsocketErrorFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		encoder := encoding.ProvideServerEncoderDecoder(nil, encoding.ContentTypeJSON)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		buildWebsocketErrorFunc(encoder)(res, req, 200, errors.New("blah"))
	})
}

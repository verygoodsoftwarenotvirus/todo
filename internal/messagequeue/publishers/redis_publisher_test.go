package publishers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

type mockMessagePublisher struct {
	mock.Mock
}

func (m *mockMessagePublisher) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return m.Called(ctx, channel, message).Get(0).(*redis.IntCmd)
}

func Test_redisPublisher_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()

		provider := ProvideRedisPublisherProvider(logger, t.Name())
		require.NotNil(t, provider)

		a, err := provider.ProviderPublisher(t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*redisPublisher)
		require.True(t, ok)

		ctx := context.Background()
		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		mmp := &mockMessagePublisher{}
		mmp.On(
			"Publish",
			testutils.ContextMatcher,
			actual.topic,
			[]byte(fmt.Sprintf(`{"name":%q}%s`, t.Name(), string(byte(10)))),
		).Return(&redis.IntCmd{})

		actual.publisher = mmp

		err = actual.Publish(ctx, inputData)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmp)
	})

	T.Run("with error encoding value", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()

		provider := ProvideRedisPublisherProvider(logger, t.Name())
		require.NotNil(t, provider)

		a, err := provider.ProviderPublisher(t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*redisPublisher)
		require.True(t, ok)

		ctx := context.Background()
		inputData := &struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		err = actual.Publish(ctx, inputData)
		assert.Error(t, err)
	})
}

func TestProvideRedisPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()

		actual := ProvideRedisPublisherProvider(logger, t.Name())
		assert.NotNil(t, actual)
	})
}

func Test_publisherProvider_ProviderPublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()

		provider := ProvideRedisPublisherProvider(logger, t.Name())
		require.NotNil(t, provider)

		actual, err := provider.ProviderPublisher(t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with cache hit", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()

		provider := ProvideRedisPublisherProvider(logger, t.Name())
		require.NotNil(t, provider)

		actual, err := provider.ProviderPublisher(t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		actual, err = provider.ProviderPublisher(t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

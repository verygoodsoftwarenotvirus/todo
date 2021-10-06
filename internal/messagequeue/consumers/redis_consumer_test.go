package consumers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

func Test_provideRedisConsumer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		actual := provideRedisConsumer(ctx, logger, redis.NewClient(&redis.Options{}), t.Name(), nil)
		assert.NotNil(t, actual)
	})
}

type mockChannelProvider struct {
	mock.Mock
}

func (m *mockChannelProvider) Channel(options ...redis.ChannelOption) <-chan *redis.Message {
	return m.Called(options).Get(0).(<-chan *redis.Message)
}

func convertChan(c chan *redis.Message) <-chan *redis.Message {
	return c
}

func Test_redisConsumer_Consume(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		hf := func(context.Context, []byte) error {
			return nil
		}

		actual := provideRedisConsumer(ctx, logger, redis.NewClient(&redis.Options{}), t.Name(), hf)
		require.NotNil(t, actual)

		returnChan := make(chan *redis.Message)
		mockSub := &mockChannelProvider{}
		mockSub.On("Channel", []redis.ChannelOption(nil)).Return(convertChan(returnChan))

		actual.subscription = mockSub
		stopChan := make(chan bool)
		errorsChan := make(chan error)

		go actual.Consume(stopChan, errorsChan)

		returnChan <- &redis.Message{}

		<-time.After(time.Second)
		stopChan <- true

		mock.AssertExpectationsForObjects(t, mockSub)
	})

	T.Run("with error handling message", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		anticipatedError := errors.New("blah")

		hf := func(context.Context, []byte) error {
			return anticipatedError
		}

		actual := provideRedisConsumer(ctx, logger, redis.NewClient(&redis.Options{}), t.Name(), hf)
		require.NotNil(t, actual)

		returnChan := make(chan *redis.Message)
		mockSub := &mockChannelProvider{}
		mockSub.On("Channel", []redis.ChannelOption(nil)).Return(convertChan(returnChan))

		actual.subscription = mockSub
		stopChan := make(chan bool)
		errorsChan := make(chan error)

		go actual.Consume(stopChan, errorsChan)

		returnChan <- &redis.Message{}

		err := <-errorsChan
		assert.Error(t, err)
		assert.Error(t, anticipatedError, err)

		stopChan <- true

		mock.AssertExpectationsForObjects(t, mockSub)
	})

	T.Run("with nil returnChan", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		actual := provideRedisConsumer(ctx, logger, redis.NewClient(&redis.Options{}), t.Name(), nil)
		require.NotNil(t, actual)

		returnChan := make(<-chan *redis.Message)
		mockSub := &mockChannelProvider{}
		mockSub.On("Channel", []redis.ChannelOption(nil)).Return(returnChan)

		actual.subscription = mockSub
		errorsChan := make(chan error)

		go actual.Consume(nil, errorsChan)

		<-time.After(time.Second)

		mock.AssertExpectationsForObjects(t, mockSub)
	})
}

func TestProvideRedisConsumerProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		exampleAddress := t.Name()

		actual := ProvideRedisConsumerProvider(logger, exampleAddress)
		assert.NotNil(t, actual)
	})
}

func Test_consumerProvider_ProviderConsumer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		exampleAddress := t.Name()

		conPro := ProvideRedisConsumerProvider(logger, exampleAddress)
		require.NotNil(t, conPro)

		ctx := context.Background()

		actual, err := conPro.ProviderConsumer(ctx, t.Name(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("hitting cache", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		exampleAddress := t.Name()

		conPro := ProvideRedisConsumerProvider(logger, exampleAddress)
		require.NotNil(t, conPro)

		ctx := context.Background()

		actual, err := conPro.ProviderConsumer(ctx, t.Name(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		actual, err = conPro.ProviderConsumer(ctx, t.Name(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

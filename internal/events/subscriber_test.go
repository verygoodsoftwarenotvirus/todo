package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/mempubsub"
)

func TestProvideSubscriber(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			Enabled:                true,
			Provider:               ProviderMemory,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          "mem://whatever",
			AckDeadline:            time.Second,
		}

		topic := mempubsub.NewTopic()
		subscription := mempubsub.NewSubscription(topic, cfg.AckDeadline)

		s, err := ProvideSubscriber(logger, subscription, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, s)
	})

	T.Run("with nil subscription", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{}

		s, err := ProvideSubscriber(logger, nil, cfg)
		assert.Error(t, err)
		assert.Nil(t, s)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		topic := mempubsub.NewTopic()
		subscription := mempubsub.NewSubscription(topic, time.Second)

		s, err := ProvideSubscriber(logger, subscription, nil)
		assert.Error(t, err)
		assert.Nil(t, s)
	})

	T.Run("with disabled config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			Enabled: false,
		}

		topic := mempubsub.NewTopic()
		subscription := mempubsub.NewSubscription(topic, cfg.AckDeadline)

		s, err := ProvideSubscriber(logger, subscription, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, s)
	})
}

func Test_subscriber_HandleEvents(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Enabled:                true,
			Provider:               ProviderMemory,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          "mem://whatever",
			AckDeadline:            time.Second,
		}

		topic := mempubsub.NewTopic()
		subscription := mempubsub.NewSubscription(topic, cfg.AckDeadline)

		s, err := ProvideSubscriber(logger, subscription, cfg)
		require.NoError(t, err)
		require.NotNil(t, s)

		p, err := ProvidePublisher(ctx, logger, cfg)
		require.NoError(t, err)
		require.NotNil(t, p)

		stopChan := make(chan bool, 1)

		var (
			calledHat sync.Mutex
			called    bool
		)
		deadline := time.After(time.Second)
		pollTicker := time.NewTicker(time.Second / 5)

		go func() {
			for {
				select {
				case <-deadline:
					stopChan <- true
				case <-pollTicker.C:
					require.NoError(t, topic.Send(ctx, &pubsub.Message{Body: []byte("{}")}))
				}
			}
		}()

		go s.HandleEvents(time.Second/10, stopChan, func(body []byte) {
			calledHat.Lock()
			defer calledHat.Unlock()

			called = true
		})

		time.Sleep(2 * time.Second)

		calledHat.Lock()
		defer calledHat.Unlock()
		assert.True(t, called)
	})
}

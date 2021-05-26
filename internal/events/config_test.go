package events

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderAWSSQS,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.NoError(t, err)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               t.Name(),
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.Error(t, err)
	})
}

func TestProvidePublishTopic(T *testing.T) {
	T.Parallel()

	T.Run("standard_memory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderMemory,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvidePublishTopic(ctx, cfg)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("standard_aws", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderAWSSQS,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvidePublishTopic(ctx, cfg)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("standard_kafka", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderKafka,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvidePublishTopic(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("standard_rabbitmq", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderRabbitMQ,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvidePublishTopic(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("standard_nats", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderNATS,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvidePublishTopic(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestProvideSubscription(T *testing.T) {
	T.Parallel()

	T.Run("standard_gcp", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderGoogleCloudPubSub,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvideSubscription(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("standard_aws", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderAWSSQS,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvideSubscription(ctx, cfg)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("standard_kafka", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderKafka,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvideSubscription(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("standard_rabbitmq", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderRabbitMQ,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvideSubscription(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("standard_nats", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderNATS,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvideSubscription(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("standard_azure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:               ProviderAzureServiceBus,
			Topic:                  t.Name(),
			SubscriptionIdentifier: t.Name(),
			ConnectionURL:          t.Name(),
			Enabled:                true,
		}

		actual, err := ProvideSubscription(ctx, cfg)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

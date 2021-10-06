package config

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/assert"
)

func Test_cleanString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, cleanString(t.Name()))
	})
}

func TestProvideConsumerProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()
		cfg := &Config{
			Provider: ProviderRedis,
		}

		provider, err := ProvideConsumerProvider(logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()
		cfg := &Config{}

		provider, err := ProvideConsumerProvider(logger, cfg)
		assert.Error(t, err)
		assert.Nil(t, provider)
	})
}

func TestProvidePublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()
		cfg := &Config{
			Provider: ProviderRedis,
		}

		provider, err := ProvidePublisherProvider(logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewZerologLogger()
		cfg := &Config{}

		provider, err := ProvidePublisherProvider(logger, cfg)
		assert.Error(t, err)
		assert.Nil(t, provider)
	})
}

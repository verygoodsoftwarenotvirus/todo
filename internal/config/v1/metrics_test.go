package config

import (
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig_ProvideInstrumentationHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c := &ServerConfig{
			Metrics: MetricsSettings{
				RuntimeMetricsCollectionInterval: time.Second,
				MetricsProvider:                  DefaultMetricsProvider,
			},
		}

		ih, err := c.ProvideInstrumentationHandler(noop.ProvideNoopLogger())
		assert.NoError(t, err)
		assert.NotNil(t, ih)
	})

	T.Run("with empty config", func(t *testing.T) {
		c := &ServerConfig{
			Metrics: MetricsSettings{
				RuntimeMetricsCollectionInterval: time.Second},
		}

		ih, err := c.ProvideInstrumentationHandler(noop.ProvideNoopLogger())
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidMetricsProvider)
		assert.Nil(t, ih)
	})
}

func TestServerConfig_ProvideTracing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c := &ServerConfig{
			Metrics: MetricsSettings{
				TracingProvider: DefaultTracingProvider,
			},
		}

		assert.NoError(t, c.ProvideTracing(noop.ProvideNoopLogger()))
	})

	T.Run("with empty config", func(t *testing.T) {
		c := &ServerConfig{
			Metrics: MetricsSettings{},
		}

		err := c.ProvideTracing(noop.ProvideNoopLogger())
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidTracingProvider)
	})
}

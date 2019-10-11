package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
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
}

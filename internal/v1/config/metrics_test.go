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

		assert.NotNil(t, c.ProvideInstrumentationHandler(noop.ProvideNoopLogger()))
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

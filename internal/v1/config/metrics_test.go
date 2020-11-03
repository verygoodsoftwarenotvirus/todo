package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestServerConfig_ProvideInstrumentationHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		c := &ServerConfig{
			Metrics: MetricsSettings{
				RuntimeMetricsCollectionInterval: time.Second,
				MetricsProvider:                  DefaultMetricsProvider,
			},
		}

		assert.NotNil(t, c.ProvideInstrumentationHandler(noop.NewLogger()))
	})
}

func TestServerConfig_ProvideTracing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		c := &ServerConfig{
			Metrics: MetricsSettings{
				TracingProvider: DefaultTracingProvider,
			},
		}

		assert.NoError(t, c.ProvideTracing(noop.NewLogger()))
	})
}

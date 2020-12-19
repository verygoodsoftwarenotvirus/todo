package observability

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestConfig_ProvideInstrumentationHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			RuntimeMetricsCollectionInterval: time.Second,
			MetricsProvider:                  DefaultMetricsProvider,
		}

		assert.NotNil(t, c.ProvideInstrumentationHandler(noop.NewLogger()))
	})
}

func TestConfig_ProvideTracing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			TracingProvider: Jaeger,
		}

		assert.NoError(t, c.InitializeTracer(noop.NewLogger()))
	})
}

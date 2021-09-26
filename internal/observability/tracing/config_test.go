package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

func TestConfig_Initialize(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Jaeger: &JaegerConfig{
				CollectorEndpoint: t.Name(),
				ServiceName:       t.Name(),
			},
			Provider: Jaeger,
		}

		actual, err := cfg.Initialize(logging.NewNoopLogger())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("without provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		actual, err := cfg.Initialize(logging.NewNoopLogger())
		assert.Nil(t, actual)
		assert.NoError(t, err)
	})
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Jaeger: &JaegerConfig{
				CollectorEndpoint: t.Name(),
				ServiceName:       t.Name(),
			},
			Provider: Jaeger,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestJaegerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &JaegerConfig{
			CollectorEndpoint: t.Name(),
			ServiceName:       t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

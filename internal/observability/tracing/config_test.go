package tracing

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/assert"
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

		actual, err := cfg.Initialize(logging.NewNonOperationalLogger())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("without provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		actual, err := cfg.Initialize(logging.NewNonOperationalLogger())
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

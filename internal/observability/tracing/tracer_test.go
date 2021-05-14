package tracing

import (
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/assert"
)

func Test_tracingErrorHandler_Handle(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		errorHandler{logger: logging.NewNonOperationalLogger()}.Handle(errors.New("blah"))
	})
}

func TestConfig_SetupJaeger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Jaeger: &JaegerConfig{
				CollectorEndpoint: "blah blah blah",
				ServiceName:       t.Name(),
			},
			Provider:                  Jaeger,
			SpanCollectionProbability: 0,
		}

		actual, err := cfg.SetupJaeger()
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with empty collector endpoint", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Jaeger: &JaegerConfig{
				CollectorEndpoint: "",
				ServiceName:       t.Name(),
			},
			Provider:                  Jaeger,
			SpanCollectionProbability: 0,
		}

		actual, err := cfg.SetupJaeger()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

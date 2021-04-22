package tracing

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"github.com/stretchr/testify/assert"
)

func TestNewInstrumentedSQLTracer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewInstrumentedSQLTracer(t.Name()))
	})
}

func Test_instrumentedSQLTracerWrapper_GetSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		w := NewInstrumentedSQLTracer(t.Name())

		assert.NotNil(t, w.GetSpan(ctx))
	})
}

func TestNewInstrumentedSQLLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewInstrumentedSQLLogger(logging.NewNonOperationalLogger()))
	})
}

func Test_instrumentedSQLLoggerWrapper_Log(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		w := NewInstrumentedSQLLogger(logging.NewNonOperationalLogger())

		w.Log(ctx, t.Name())
	})
}

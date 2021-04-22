package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTracer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewTracer(t.Name()))
	})
}

func Test_otelSpanManager_StartSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		NewTracer(t.Name()).StartSpan(context.Background())
	})
}

func Test_otelSpanManager_StartCustomSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		NewTracer(t.Name()).StartCustomSpan(ctx, t.Name())
	})
}

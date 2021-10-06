package workers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

func TestProvideDataChangesWorker(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := ProvideDataChangesWorker(logging.NewZerologLogger())
		assert.NotNil(t, actual)
	})
}

func TestDataChangesWorker_HandleMessage(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := ProvideDataChangesWorker(logging.NewZerologLogger())
		assert.NotNil(t, actual)

		ctx := context.Background()
		assert.NoError(t, actual.HandleMessage(ctx, []byte("{}")))
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()

		actual := ProvideDataChangesWorker(logging.NewZerologLogger())
		assert.NotNil(t, actual)

		ctx := context.Background()
		assert.Error(t, actual.HandleMessage(ctx, []byte("} bad JSON lol")))
	})
}

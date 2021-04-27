package observability

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		descriptionFmt, descriptionArgs := "things and %s", "stuff"
		err := errors.New("blah")
		logger := logging.NewNonOperationalLogger()
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareError(err, logger, span, descriptionFmt, descriptionArgs))
	})
}

func TestAcknowledgeError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		descriptionFmt, descriptionArgs := "things and %s", "stuff"
		err := errors.New("blah")
		logger := logging.NewNonOperationalLogger()
		_, span := tracing.StartSpan(ctx)

		AcknowledgeError(err, logger, span, descriptionFmt, descriptionArgs)
	})
}

func TestNoteEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		descriptionFmt, descriptionArgs := "things and %s", "stuff"
		logger := logging.NewNonOperationalLogger()
		_, span := tracing.StartSpan(ctx)

		NoteEvent(logger, span, descriptionFmt, descriptionArgs)
	})
}

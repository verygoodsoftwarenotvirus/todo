package converters

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestConvertAuditLogEntryCreationInputToEntry(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleInput := fakes.BuildFakeAuditLogEntryCreationInput()
		actual := ConvertAuditLogEntryCreationInputToEntry(exampleInput)

		assert.Equal(t, exampleInput.EventType, actual.EventType)
		assert.Equal(t, exampleInput.Context, actual.Context)
	})
}

package converters

import (
	"testing"

	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fake"

	"github.com/stretchr/testify/assert"
)

func TestConvertAuditLogEntryCreationInputToEntry(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleInput := fakemodels.BuildFakeAuditLogEntryCreationInput()
		actual := ConvertAuditLogEntryCreationInputToEntry(exampleInput)

		assert.Equal(t, exampleInput.EventType, actual.EventType)
		assert.Equal(t, exampleInput.Context, actual.Context)
	})

}

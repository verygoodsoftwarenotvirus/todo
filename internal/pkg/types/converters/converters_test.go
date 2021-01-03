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

func TestConvertItemToItemUpdateInput(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		expected := fakes.BuildFakeItem()
		actual := ConvertItemToItemUpdateInput(expected)

		assert.Equal(t, expected.Name, actual.Name, "expected BucketName to equal %q, but encountered %q instead", expected.Name, actual.Name)
		assert.Equal(t, expected.Details, actual.Details, "expected Details to equal %q, but encountered %q instead", expected.Name, actual.Name)
	})
}

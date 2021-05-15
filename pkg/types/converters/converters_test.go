package converters

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestConvertAuditLogEntryCreationInputToEntry(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleInput := fakes.BuildFakeAuditLogEntryCreationInput()
		actual := ConvertAuditLogEntryCreationInputToEntry(exampleInput)

		assert.Equal(t, exampleInput.EventType, actual.EventType)
		assert.Equal(t, exampleInput.Context, actual.Context)
	})
}

func TestConvertAccountToAccountUpdateInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleInput := fakes.BuildFakeAccount()
		actual := ConvertAccountToAccountUpdateInput(exampleInput)

		assert.Equal(t, exampleInput.Name, actual.Name)
		assert.Equal(t, exampleInput.BelongsToUser, actual.BelongsToUser)
	})
}

func TestConvertItemToItemUpdateInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := fakes.BuildFakeItem()
		actual := ConvertItemToItemUpdateInput(expected)

		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.Details, actual.Details)
	})
}

func TestConvertAccountSubscriptionPlanToPlanUpdateInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := fakes.BuildFakeAccountSubscriptionPlan()
		actual := ConvertAccountSubscriptionPlanToPlanUpdateInput(expected)

		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.Description, actual.Description)
		assert.Equal(t, expected.Price, actual.Price)
		assert.Equal(t, expected.Period, actual.Period)
	})
}

package types

import (
	"context"
	"testing"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/stretchr/testify/assert"
)

func TestItem_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &Item{}

		updated := &ItemUpdateInput{
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		expected := []*FieldChangeSummary{
			{
				FieldName: "Name",
				OldValue:  x.Name,
				NewValue:  updated.Name,
			},
			{
				FieldName: "Details",
				OldValue:  x.Details,
				NewValue:  updated.Details,
			},
		}
		actual := x.Update(updated)
		assert.Equal(t, expected, actual, "expected and actual diff reports vary")

		assert.Equal(t, updated.Name, x.Name)
		assert.Equal(t, updated.Details, x.Details)
	})
}

func TestItemCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ItemCreationInput{
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		actual := x.ValidateWithContext(context.Background())
		assert.Nil(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ItemCreationInput{}

		actual := x.ValidateWithContext(context.Background())
		assert.Error(t, actual)
	})
}

func TestItemUpdateInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ItemUpdateInput{
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		actual := x.ValidateWithContext(context.Background())
		assert.Nil(t, actual)
	})

	T.Run("with empty strings", func(t *testing.T) {
		t.Parallel()

		x := &ItemUpdateInput{}

		actual := x.ValidateWithContext(context.Background())
		assert.Error(t, actual)
	})
}

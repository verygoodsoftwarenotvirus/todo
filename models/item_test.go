package models

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/stretchr/testify/assert"
)

func TestItem_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		i := &Item{}

		updated := &ItemUpdateInput{
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		expected := []FieldChangeSummary{
			{
				FieldName: "Name",
				OldValue:  i.Name,
				NewValue:  updated.Name,
			},
			{
				FieldName: "Details",
				OldValue:  i.Details,
				NewValue:  updated.Details,
			},
		}
		actual := i.Update(updated)
		assert.Equal(t, expected, actual, "expected and actual diff reports vary")

		assert.Equal(t, updated.Name, i.Name)
		assert.Equal(t, updated.Details, i.Details)
	})
}

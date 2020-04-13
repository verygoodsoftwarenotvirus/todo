package models

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/stretchr/testify/assert"
)

func TestItem_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		i := &Item{}

		expected := &ItemUpdateInput{
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		i.Update(expected)
		assert.Equal(t, expected.Name, i.Name)
		assert.Equal(t, expected.Details, i.Details)
	})
}

func TestItem_ToUpdateInput(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		item := &Item{
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		expected := &ItemUpdateInput{
			Name:    item.Name,
			Details: item.Details,
		}
		actual := item.ToUpdateInput()

		assert.Equal(t, expected, actual)
	})
}

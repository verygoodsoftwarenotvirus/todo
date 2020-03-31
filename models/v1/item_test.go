package models

import (
	"testing"

	fake "github.com/brianvoe/gofakeit"
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
		i := &Item{
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		expected := &ItemUpdateInput{
			Name:    i.Name,
			Details: i.Details,
		}
		actual := i.ToUpdateInput()

		assert.Equal(t, expected, actual)
	})
}

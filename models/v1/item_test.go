package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItem_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		i := &Item{}

		expected := &ItemUpdateInput{
			Name:    "expected name",
			Details: "expected details",
		}

		i.Update(expected)
		assert.Equal(t, expected.Name, i.Name)
		assert.Equal(t, expected.Details, i.Details)
	})
}

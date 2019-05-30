package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestItem_Update(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		i := &Item{
			Name:    "name",
			Details: "deets",
		}

		expected := &ItemInput{
			Name:    "expected name",
			Details: "expected details",
		}

		i.Update(expected)
		assert.Equal(t, expected.Name, i.Name)
		assert.Equal(t, expected.Details, i.Details)
	})
}

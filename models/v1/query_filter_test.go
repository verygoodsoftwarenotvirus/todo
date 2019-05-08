package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArbitrary(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		assert.True(t, true)
	})
}

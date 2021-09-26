package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildURLVarChunk(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := "/{things:stuff}"
		actual := buildURLVarChunk("things", "stuff")

		assert.Equal(t, expected, actual)
	})
}

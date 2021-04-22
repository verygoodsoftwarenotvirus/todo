package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCallerName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, GetCallerName())
	})
}

package httpserver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvideHTTPServer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := provideHTTPServer(8888)

		assert.NotNil(t, x)
	})
}

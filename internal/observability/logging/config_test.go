package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvideLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Provider: ProviderZerolog,
		}

		assert.NotNil(t, ProvideLogger(cfg))
	})

	T.Run("no provider", func(t *testing.T) {
		t.Parallel()

		cfg := Config{}

		assert.NotNil(t, ProvideLogger(cfg))
	})
}

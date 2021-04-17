package frontend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			StaticFilesDirectory: "blah",
		}

		assert.NoError(t, cfg.Validate(context.Background()))
	})

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		assert.Error(t, cfg.Validate(context.Background()))
	})
}

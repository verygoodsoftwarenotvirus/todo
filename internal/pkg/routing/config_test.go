package routing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:       ChiProvider,
			AllowedOrigins: nil,
			AllowedHeaders: nil,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

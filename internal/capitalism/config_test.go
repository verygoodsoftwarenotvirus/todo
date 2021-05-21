package capitalism

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &StripeConfig{
			APIKey: "blah",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing API key", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &StripeConfig{
			APIKey: "",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

package encoding

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
			ContentType: contentTypeJSON,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

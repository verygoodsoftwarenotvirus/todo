package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		ctx := context.Background()

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

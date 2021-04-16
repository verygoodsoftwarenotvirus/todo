package audit

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		ctx := context.Background()

		assert.NoError(t, cfg.Validate(ctx))
	})
}

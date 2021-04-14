package httpserver

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			StartupDeadline: time.Second,
			HTTPPort:        8080,
			Debug:           true,
		}

		assert.NoError(t, cfg.Validate(ctx))
	})
}

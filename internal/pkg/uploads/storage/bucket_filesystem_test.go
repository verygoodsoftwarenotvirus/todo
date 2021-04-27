package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &FilesystemConfig{
			RootDirectory: t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3Config_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &S3Config{
			BucketName: t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

package uploads

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/storage"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Storage: storage.Config{
				FilesystemConfig: &storage.FilesystemConfig{RootDirectory: "/blah"},
				AzureConfig: &storage.AzureConfig{
					BucketName: "blahs",
					Retrying:   &storage.AzureRetryConfig{},
				},
				GCSConfig: &storage.GCSConfig{
					ServiceAccountKeyFilepath: "/blah/blah",
					BucketName:                "blah",
					Scopes:                    nil,
				},
				S3Config:          &storage.S3Config{BucketName: "blahs"},
				BucketName:        "blahs",
				UploadFilenameKey: "blahs",
				Provider:          "blahs",
			},
			Debug: false,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

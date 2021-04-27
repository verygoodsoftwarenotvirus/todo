package storage

import (
	"context"
	"net/http"
	"os"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			BucketName:       t.Name(),
			Provider:         FilesystemProvider,
			FilesystemConfig: &FilesystemConfig{RootDirectory: t.Name()},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestNewUploadManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		l := logging.NewNonOperationalLogger()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   MemoryProvider,
		}
		rpm := &mockrouting.RouteParamManager{}
		rpm.On("BuildRouteParamStringIDFetcher", cfg.UploadFilenameKey).Return(func(*http.Request) string { return t.Name() })

		x, err := NewUploadManager(ctx, l, cfg, rpm)
		assert.NotNil(t, x)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, rpm)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		l := logging.NewNonOperationalLogger()
		rpm := &mockrouting.RouteParamManager{}

		x, err := NewUploadManager(ctx, l, nil, rpm)
		assert.Nil(t, x)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, rpm)
	})

	T.Run("with invalid config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		l := logging.NewNonOperationalLogger()
		cfg := &Config{}
		rpm := &mockrouting.RouteParamManager{}
		rpm.On("BuildRouteParamStringIDFetcher", cfg.UploadFilenameKey).Return(func(*http.Request) string { return t.Name() })

		x, err := NewUploadManager(ctx, l, cfg, rpm)
		assert.Nil(t, x)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}

func TestUploader_selectBucket(T *testing.T) {
	T.Parallel()

	T.Run("azure happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider: AzureProvider,
			AzureConfig: &AzureConfig{
				AccountName: "blah",
				BucketName:  "blahs",
				Retrying:    &AzureRetryConfig{},
			},
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("azure with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider:    AzureProvider,
			AzureConfig: nil,
		}

		assert.Error(t, u.selectBucket(ctx, cfg))
	})

	T.Run("gcs with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider:  GCSProvider,
			GCSConfig: nil,
		}

		assert.Error(t, u.selectBucket(ctx, cfg))
	})

	T.Run("s3 happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider: S3Provider,
			S3Config: &S3Config{
				BucketName: t.Name(),
			},
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("s3 with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider: S3Provider,
			S3Config: nil,
		}

		assert.Error(t, u.selectBucket(ctx, cfg))
	})

	T.Run("memory provider", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider: MemoryProvider,
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("filesystem happy path", func(t *testing.T) {
		t.Parallel()

		tempDir := os.TempDir()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider: FilesystemProvider,
			FilesystemConfig: &FilesystemConfig{
				RootDirectory: tempDir,
			},
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("filesystem with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		u := &Uploader{}
		cfg := &Config{
			Provider:         FilesystemProvider,
			FilesystemConfig: nil,
		}

		assert.Error(t, u.selectBucket(ctx, cfg))
	})
}

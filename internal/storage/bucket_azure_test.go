package storage

import (
	"context"
	"encoding/base64"
	"math"
	"testing"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

func TestAzureConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &AzureConfig{
			AuthMethod:                   azureTokenAuthMethod,
			AccountName:                  t.Name(),
			BucketName:                   t.Name(),
			TokenCredentialsInitialToken: t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestAzureConfig_authMethodIsSharedKey(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.True(t, (&AzureConfig{AuthMethod: azureSharedKeyAuthMethod4}).authMethodIsSharedKey())
	})
}

func TestAzureRetryConfig_buildRetryOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, (&AzureRetryConfig{}).buildRetryOptions())
	})
}

func Test_provideAzureBucket(T *testing.T) {
	T.Parallel()

	T.Run("with anonymous credential", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &AzureConfig{
			AccountName: t.Name(),
			BucketName:  t.Name(),
		}

		x, err := provideAzureBucket(ctx, cfg, logging.NewNoopLogger())
		assert.NoError(t, err)
		assert.NotNil(t, x)
	})

	T.Run("with shared key", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &AzureConfig{
			AuthMethod:          azureSharedKeyAuthMethod4,
			AccountName:         t.Name(),
			BucketName:          t.Name(),
			SharedKeyAccountKey: base64.StdEncoding.EncodeToString([]byte(t.Name())),
		}

		x, err := provideAzureBucket(ctx, cfg, logging.NewNoopLogger())
		assert.NoError(t, err)
		assert.NotNil(t, x)
	})

	T.Run("with shared key and not shared key", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &AzureConfig{
			AuthMethod:          azureSharedKeyAuthMethod4,
			AccountName:         t.Name(),
			BucketName:          t.Name(),
			SharedKeyAccountKey: "",
		}

		x, err := provideAzureBucket(ctx, cfg, logging.NewNoopLogger())
		assert.Error(t, err)
		assert.Nil(t, x)
	})

	T.Run("with shared key and invalid key", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &AzureConfig{
			AuthMethod:          azureSharedKeyAuthMethod4,
			AccountName:         t.Name(),
			BucketName:          t.Name(),
			SharedKeyAccountKey: "        lol not valid base64       ",
		}

		x, err := provideAzureBucket(ctx, cfg, logging.NewNoopLogger())
		assert.Error(t, err)
		assert.Nil(t, x)
	})

	T.Run("with token", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &AzureConfig{
			AuthMethod:                   azureTokenAuthMethod,
			BucketName:                   t.Name(),
			AccountName:                  t.Name(),
			TokenCredentialsInitialToken: t.Name(),
		}

		x, err := provideAzureBucket(ctx, cfg, logging.NewNoopLogger())
		assert.NoError(t, err)
		assert.NotNil(t, x)
	})

	T.Run("with token auth and no token", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &AzureConfig{
			AuthMethod:                   azureTokenAuthMethod,
			BucketName:                   t.Name(),
			AccountName:                  t.Name(),
			TokenCredentialsInitialToken: "",
		}

		x, err := provideAzureBucket(ctx, cfg, logging.NewNoopLogger())
		assert.Error(t, err)
		assert.Nil(t, x)
	})
}

func Test_buildPipelineLogFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := buildPipelineLogFunc(logging.NewNoopLogger())

		for _, level := range []pipeline.LogLevel{
			pipeline.LogNone,
			pipeline.LogError,
			pipeline.LogWarning,
			pipeline.LogInfo,
			pipeline.LogDebug,
			pipeline.LogLevel(math.MaxUint32),
		} {
			x(level, t.Name())
		}
	})
}

func Test_buildPipelineOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, buildPipelineOptions(logging.NewNoopLogger(), &AzureRetryConfig{}))
	})
}

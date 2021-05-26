package secrets

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"

	"github.com/stretchr/testify/require"
)

func buildExampleKey(ctx context.Context, t *testing.T) string {
	t.Helper()

	rawBytes, err := random.GenerateRawBytes(ctx, expectedLocalKeyLength)
	require.NoError(t, err)

	return base64.URLEncoding.EncodeToString(rawBytes)
}

func TestProvideSecretKeeper(T *testing.T) {
	T.Parallel()

	T.Run("standard_local", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		cfg := &Config{
			Provider: ProviderLocal,
			Key:      buildExampleKey(ctx, t),
		}

		actual, err := ProvideSecretKeeper(ctx, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("standard_aws", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		cfg := &Config{
			Provider: ProviderAWS,
			Key:      buildExampleKey(ctx, t),
		}

		actual, err := ProvideSecretKeeper(ctx, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("standard_vault", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		cfg := &Config{
			Provider: ProviderHashicorpVault,
			Key:      buildExampleKey(ctx, t),
		}

		actual, err := ProvideSecretKeeper(ctx, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

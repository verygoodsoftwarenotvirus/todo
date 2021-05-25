package secrets

import (
	"context"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"

	"github.com/stretchr/testify/require"
)

type example struct {
	Name string
}

func buildTestSecretManager(t *testing.T) SecretManager {
	t.Helper()

	ctx := context.Background()
	logger := logging.NewNoopLogger()

	b, err := random.GenerateRawBytes(ctx, expectedLocalKeyLength)
	require.NoError(t, err)
	require.NotNil(t, b)

	cfg := &Config{
		Provider: ProviderLocal,
		Key:      base64.URLEncoding.EncodeToString(b),
	}

	k, err := ProvideSecretKeeper(ctx, cfg)
	require.NotNil(t, k)
	require.NoError(t, err)

	sm, err := ProvideSecretManager(logger, k)
	require.NotNil(t, sm)
	require.NoError(t, err)

	return sm
}

func TestProvideSecretManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		b, err := random.GenerateRawBytes(ctx, expectedLocalKeyLength)

		cfg := &Config{
			Provider: ProviderLocal,
			Key:      b,
		}

		k, err := ProvideSecretKeeper(ctx, cfg)
		require.NotNil(t, k)
		require.NoError(t, err)
	})
}

func Test_secretManager_Encrypt(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		cfg := &Config{
			Provider: ProviderLocal,
			Key:      "j6wdsvcb6E6uh0M36scNwz3MQ-IwE8Uja-BQm2hs5u0=",
		}

		k, err := ProvideSecretKeeper(ctx, cfg)
		require.NotNil(t, k)
		require.NoError(t, err)

		sm, err := ProvideSecretManager(logger, k)
		require.NotNil(t, sm)
		require.NoError(t, err)

		exampleInput := &example{Name: t.Name()}

		encrypted, err := sm.Encrypt(ctx, exampleInput)
		require.NotEmpty(t, encrypted)
		require.NoError(t, err)

		expected := "yuTt+xqENPZuOxs+H+zzu6axC56jkvss448PAyrvWV+OFmKVZj+89H9iqsQaPbvWT3VvotfucsIz3KNjWjnEZv/4hQhW4tIyuffVjnDZsxy5bT/y7EQ="
		actual := base64.URLEncoding.EncodeToString(encrypted)

		assert.Equal(t, expected, actual)
	})
}

func Test_secretManager_Decrypt(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		sm := buildTestSecretManager(t)

		expected := &example{Name: t.Name()}

		encrypted, err := sm.Encrypt(ctx, expected)
		require.NotEmpty(t, encrypted)
		require.NoError(t, err)

		var actual *example
		err = sm.Decrypt(ctx, encrypted, &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

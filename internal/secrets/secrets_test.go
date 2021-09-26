package secrets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/secrets"
	"gocloud.dev/secrets/localsecrets"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
)

type example struct {
	Name string
}

func buildTestSecretKeeper(ctx context.Context, t *testing.T) *secrets.Keeper {
	t.Helper()

	b, err := random.GenerateRawBytes(ctx, expectedLocalKeyLength)
	require.NoError(t, err)
	require.NotNil(t, b)

	key := base64.URLEncoding.EncodeToString(b)
	cfg := &Config{
		Provider: ProviderLocal,
		Key:      key,
	}

	k, err := ProvideSecretKeeper(ctx, cfg)
	require.NotNil(t, k)
	require.NoError(t, err)

	return k
}

func buildTestSecretManager(t *testing.T) SecretManager {
	t.Helper()

	ctx := context.Background()
	logger := logging.NewNoopLogger()

	k := buildTestSecretKeeper(ctx, t)
	require.NotNil(t, k)

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

		k := buildTestSecretKeeper(ctx, t)

		sm, err := ProvideSecretManager(nil, k)
		require.NoError(t, err)
		require.NotNil(t, sm)
	})

	T.Run("with nil keeper", func(t *testing.T) {
		t.Parallel()

		k, err := ProvideSecretManager(nil, nil)
		require.Nil(t, k)
		require.Error(t, err)
	})
}

type broken struct {
	Thing json.Number `json:"thing"`
}

func Test_secretManager_Encrypt(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		rawKey, err := localsecrets.NewRandomKey()
		require.NoError(t, err)

		key := []byte{}
		for i := range rawKey {
			key = append(key, rawKey[i])
		}

		cfg := &Config{
			Provider: ProviderLocal,
			Key:      base64.URLEncoding.EncodeToString(key),
		}

		k, err := ProvideSecretKeeper(ctx, cfg)
		require.NotNil(t, k)
		require.NoError(t, err)

		sm, err := ProvideSecretManager(logger, k)
		require.NotNil(t, sm)
		require.NoError(t, err)

		exampleInput := &example{Name: t.Name()}

		actual, err := sm.Encrypt(ctx, exampleInput)
		require.NoError(t, err)
		assert.NotEmpty(t, actual)
	})

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		rawKey, err := localsecrets.NewRandomKey()
		require.NoError(t, err)

		key := []byte{}
		for i := range rawKey {
			key = append(key, rawKey[i])
		}

		cfg := &Config{
			Provider: ProviderLocal,
			Key:      base64.URLEncoding.EncodeToString(key),
		}

		k, err := ProvideSecretKeeper(ctx, cfg)
		require.NotNil(t, k)
		require.NoError(t, err)

		sm, err := ProvideSecretManager(logger, k)
		require.NotNil(t, sm)
		require.NoError(t, err)

		exampleInput := &broken{Thing: json.Number(t.Name())}

		actual, err := sm.Encrypt(ctx, exampleInput)
		require.Error(t, err)
		assert.Empty(t, actual)
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

	T.Run("with invalid value", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		sm := buildTestSecretManager(t)

		var actual *example
		err := sm.Decrypt(ctx, " this isn't a real string lol ", &actual)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with inability to decrypt", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		sm := buildTestSecretManager(t)

		expected := &example{Name: t.Name()}

		encrypted, err := sm.Encrypt(ctx, expected)
		require.NotEmpty(t, encrypted)
		require.NoError(t, err)

		sm.(*secretManager).keeper = buildTestSecretKeeper(ctx, t)

		var actual *example
		err = sm.Decrypt(ctx, encrypted, &actual)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid JSON value", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		sm := buildTestSecretManager(t)

		encrypted, err := sm.(*secretManager).keeper.Encrypt(ctx, []byte(` this isn't a real JSON string lol `))
		require.NoError(t, err)
		encoded := base64.URLEncoding.EncodeToString(encrypted)

		var actual *example
		err = sm.Decrypt(ctx, encoded, &actual)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

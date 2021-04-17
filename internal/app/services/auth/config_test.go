package auth

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/random"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		localKey, err := random.GenerateRawBytes(ctx, pasetoKeyRequiredLength)
		require.NoError(t, err)

		cfg := &Config{
			PASETO: PASETOConfig{
				Issuer:       "issuer",
				LocalModeKey: localKey,
			},
			Cookies: CookieConfig{
				Name:     "name",
				Domain:   "domain",
				Lifetime: time.Second,
			},
			Debug:                 false,
			EnableUserSignup:      false,
			MinimumUsernameLength: 123,
			MinimumPasswordLength: 123,
		}

		assert.NoError(t, cfg.Validate(ctx))
	})
}

func TestCookieConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &CookieConfig{
			Name:     "name",
			Domain:   "domain",
			Lifetime: time.Second,
		}
		ctx := context.Background()

		assert.NoError(t, cfg.Validate(ctx))
	})
}

func TestPASETOConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		localKey, err := random.GenerateRawBytes(ctx, pasetoKeyRequiredLength)
		require.NoError(t, err)

		cfg := &PASETOConfig{
			Issuer:       "issuer",
			LocalModeKey: localKey,
		}

		assert.NoError(t, cfg.Validate(ctx))
	})
}

package secrets

import (
	"context"
	"errors"
	"fmt"

	"gocloud.dev/secrets"
	"gocloud.dev/secrets/localsecrets"
)

const (
	// ProviderLocal is the thing to use to indicate you want the Local provider for secret management.
	ProviderLocal = "local"

	expectedLocalKeyLength = 32
)

var (
	errInvalidProvider = errors.New("invalid provider")
	errNilConfig       = errors.New("nil config provided")
)

// Config is how we configure the secret manager.
type Config struct {
	_ struct{}

	Provider string
	Key      string
}

// ProvideSecretKeeper provides a new secret keeper.
func ProvideSecretKeeper(_ context.Context, cfg *Config) (*secrets.Keeper, error) {
	if cfg == nil {
		return nil, errNilConfig
	}

	switch cfg.Provider {
	case ProviderLocal:
		key, err := localsecrets.Base64Key(cfg.Key)
		if err != nil {
			return nil, fmt.Errorf("doing: %w", err)
		}

		keeper := localsecrets.NewKeeper(key)

		return keeper, nil
	default:
		return nil, errInvalidProvider
	}
}

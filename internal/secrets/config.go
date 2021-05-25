package secrets

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/api"

	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/secrets"
	"gocloud.dev/secrets/awskms"
	"gocloud.dev/secrets/azurekeyvault"
	"gocloud.dev/secrets/gcpkms"
	"gocloud.dev/secrets/hashivault"
	"gocloud.dev/secrets/localsecrets"
)

const (
	ProviderLocal          = "local"
	ProviderGCP            = "gcp_kms"
	ProviderAWS            = "aws_kms"
	ProviderAzureKeyVault  = "azure_keyvault"
	ProviderHashicorpVault = "vault"

	expectedLocalKeyLength = 32
)

var (
	errInvalidProvider = errors.New("invalid provider")
	errNilConfig       = errors.New("nil config provided")
)

// Config is how we configure the secret manager.
type Config struct {
	Provider string
	Key      string
}

// ProvideSecretKeeper provides a new secret keeper.
func ProvideSecretKeeper(ctx context.Context, cfg *Config) (*secrets.Keeper, error) {
	if cfg == nil {
		return nil, errNilConfig
	}

	switch cfg.Provider {
	case ProviderGCP:
		// Get a client to use with the KMS API.
		client, _, err := gcpkms.Dial(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("connecting to GCP KMS: %w", err)
		}

		keeper := gcpkms.OpenKeeper(client, cfg.Key, nil)

		return keeper, nil
	case ProviderAWS:
		sess, err := session.NewSession(nil)
		if err != nil {
			return nil, fmt.Errorf("doing: %w", err)
		}

		// Get a client to use with the KMS API.
		client, err := awskms.Dial(sess)
		if err != nil {
			return nil, fmt.Errorf("doing: %w", err)
		}

		// Construct a *secrets.Keeper.
		keeper := awskms.OpenKeeper(client, cfg.Key, nil)

		return keeper, nil
	case ProviderAzureKeyVault:

		client, err := azurekeyvault.Dial()
		if err != nil {
			return nil, fmt.Errorf("doing: %w", err)
		}

		// Construct a *secrets.Keeper.
		keeper, err := azurekeyvault.OpenKeeper(client, cfg.Key, nil)
		if err != nil {
			return nil, fmt.Errorf("doing: %w", err)
		}

		return keeper, nil
	case ProviderHashicorpVault:
		// Get a client to use with the Vault API.
		client, err := hashivault.Dial(ctx, &hashivault.Config{
			Token: "",
			APIConfig: api.Config{
				Address: "",
			},
		})
		if err != nil {
			return nil, fmt.Errorf("doing: %w", err)
		}

		// Construct a *secrets.Keeper.
		keeper := hashivault.OpenKeeper(client, cfg.Key, nil)

		return keeper, nil
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

package gocloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
)

const (
	// AzureProvider indicates we'd like to use the azure adapter for blob.
	AzureProvider = "azure"
)

type (
	// AzureRetryConfig configures storage retries.
	AzureRetryConfig struct {
		MaxTries                    int32         `json:"max_tries" mapstructure:"max_tries" toml:"max_tries,omitempty"`
		TryTimeout                  time.Duration `json:"try_timeout" mapstructure:"try_timeout" toml:"try_timeout,omitempty"`
		RetryDelay                  time.Duration `json:"retry_delay" mapstructure:"retry_delay" toml:"retry_delay,omitempty"`
		MaxRetryDelay               time.Duration `json:"max_retry_delay" mapstructure:"max_retry_delay" toml:"max_retry_delay,omitempty"`
		RetryReadsFromSecondaryHost string        `json:"retry_reads_from_secondary_host" mapstructure:"retry_reads_from_secondary_host" toml:"retry_reads_from_secondary_host,omitempty"`
	}

	// AzureConfig configures an azure instance of an UploadManager.
	AzureConfig struct {
		AuthMethod                   string            `json:"auth_method" mapstructure:"auth_method" toml:"auth_method,omitempty"`
		AccountName                  string            `json:"account_name" mapstructure:"account_name" toml:"account_name,omitempty"`
		ContainerName                string            `json:"container_name" mapstructure:"container_name" toml:"container_name,omitempty"`
		Retrying                     *AzureRetryConfig `json:"retrying" mapstructure:"retrying" toml:"retrying,omitempty"`
		TokenCredentialsInitialToken string            `json:"token_creds_initial_token" mapstructure:"token_creds_initial_token" toml:"token_creds_initial_token,omitempty"`
		SharedKeyAccountKey          string            `json:"shared_key_account_key" mapstructure:"shared_key_aaccount_key" toml:"shared_key_account_key,omitempty"`
	}
)

func (cfg *AzureRetryConfig) buildRetryOptions() azblob.RetryOptions {
	return azblob.RetryOptions{
		Policy:                      azblob.RetryPolicyExponential,
		MaxTries:                    cfg.MaxTries,
		TryTimeout:                  cfg.TryTimeout,
		RetryDelay:                  cfg.RetryDelay,
		MaxRetryDelay:               cfg.MaxRetryDelay,
		RetryReadsFromSecondaryHost: cfg.RetryReadsFromSecondaryHost,
	}
}

func (c *AzureConfig) authMethodIsSharedKey() bool {
	return c.AuthMethod == azureSharedKeyAuthMethod1 ||
		c.AuthMethod == azureSharedKeyAuthMethod2 ||
		c.AuthMethod == azureSharedKeyAuthMethod3 ||
		c.AuthMethod == azureSharedKeyAuthMethod4
}

// Validate validates the AzureConfig.
func (c *AzureConfig) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.AuthMethod, validation.Required),
		validation.Field(&c.AccountName, validation.Required),
		validation.Field(&c.ContainerName, validation.Required),
		validation.Field(&c.Retrying, validation.When(c.Retrying != nil, validation.Required)),
		validation.Field(&c.SharedKeyAccountKey, validation.When(c.authMethodIsSharedKey(), validation.Required).Else(validation.Nil)),
		validation.Field(&c.TokenCredentialsInitialToken, validation.When(c.AuthMethod == azureTokenAuthMethod, validation.Required).Else(validation.Nil)),
	)
}

const (
	azureSharedKeyAuthMethod1 = "sharedkey"
	azureSharedKeyAuthMethod2 = "shared-key"
	azureSharedKeyAuthMethod3 = "shared_key"
	azureSharedKeyAuthMethod4 = "shared"
	azureTokenAuthMethod      = "token"
)

func provideAzureBucket(ctx context.Context, cfg *AzureConfig, logger logging.Logger) (*blob.Bucket, error) {
	var (
		cred   azblob.Credential
		bucket *blob.Bucket
		err    error
	)

	switch strings.TrimSpace(strings.ToLower(cfg.AuthMethod)) {
	case azureSharedKeyAuthMethod1, azureSharedKeyAuthMethod2, azureSharedKeyAuthMethod3, azureSharedKeyAuthMethod4:
		if cfg.SharedKeyAccountKey == "" {
			return nil, ErrInvalidConfiguration
		}

		if cred, err = azblob.NewSharedKeyCredential(
			cfg.AccountName,
			cfg.SharedKeyAccountKey,
		); err != nil {
			return nil, fmt.Errorf("error reading shared key credential: %w", err)
		}
	case azureTokenAuthMethod:
		if cfg.TokenCredentialsInitialToken == "" {
			return nil, ErrInvalidConfiguration
		}

		cred = azblob.NewTokenCredential(cfg.TokenCredentialsInitialToken, nil)
	default:
		cred = azblob.NewAnonymousCredential()
	}

	if bucket, err = azureblob.OpenBucket(
		ctx,
		azureblob.NewPipeline(cred, buildPipelineOptions(logger, cfg.Retrying)),
		azureblob.AccountName(cfg.AccountName),
		cfg.ContainerName,
		nil,
	); err != nil {
		return nil, fmt.Errorf("error initializing azure bucket: %w", err)
	}

	return bucket, nil
}

func buildPipelineOptions(logger logging.Logger, retrying *AzureRetryConfig) azblob.PipelineOptions {
	options := azblob.PipelineOptions{
		Log: pipeline.LogOptions{
			Log: func(level pipeline.LogLevel, message string) {
				switch level {
				case pipeline.LogNone:
					// shouldn't happen, but do nothing just in case
				case pipeline.LogPanic, pipeline.LogFatal:
					logger.Fatal(errors.New(message))
				case pipeline.LogError:
					logger.Error(errors.New("message"), "azure pipeline error")
				case pipeline.LogWarning:
					logger.Debug(message)
				case pipeline.LogInfo:
					logger.Info(message)
				case pipeline.LogDebug:
					logger.Debug(message)
				default:
					logger.Debug(message)
				}
			},
			ShouldLog: func(level pipeline.LogLevel) bool {
				return level != pipeline.LogNone
			},
		},
	}

	if retrying != nil {
		options.Retry = retrying.buildRetryOptions()
	}

	return options
}

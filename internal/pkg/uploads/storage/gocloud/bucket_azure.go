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
		MaxTries                    int32
		TryTimeout                  time.Duration
		RetryDelay                  time.Duration
		MaxRetryDelay               time.Duration
		RetryReadsFromSecondaryHost string
	}

	// AzureTokenCreds configures using a token to authenticate.
	AzureTokenCreds struct {
		InitialToken string
	}

	// AzureSharedKeyConfig configures using a shared key to authenticate.
	AzureSharedKeyConfig struct {
		AccountName string
		AccountKey  string
	}

	// AzureUserDelegationConfig configures using a user delegation key to authenticate.
	AzureUserDelegationConfig struct {
		AccountName               string
		UserDelegationKeyFilepath string
	}

	// AzureConfig configures an azure instance of an UploadManager.
	AzureConfig struct {
		AuthMethod           string
		AccountName          string
		ContainerName        string
		Retrying             *AzureRetryConfig
		TokenCreds           *AzureTokenCreds
		SharedKeyConfig      *AzureSharedKeyConfig
		UserDelegationConfig *AzureUserDelegationConfig
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

// Validate validates the AzureTokenCreds.
func (c *AzureTokenCreds) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.InitialToken, validation.Required),
	)
}

// Validate validates the AzureSharedKeyConfig.
func (c *AzureSharedKeyConfig) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.AccountName, validation.Required),
		validation.Field(&c.AccountKey, validation.Required),
	)
}

// Validate validates the AzureUserDelegationConfig.
func (c *AzureUserDelegationConfig) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c, validation.Required),
	)
}

// Validate validates the AzureConfig.
func (c *AzureConfig) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.AuthMethod, validation.Required),
		validation.Field(&c.AccountName, validation.Required),
		validation.Field(&c.ContainerName, validation.Required),
		validation.Field(&c.Retrying, validation.When(c.Retrying != nil, validation.Required)),
		validation.Field(&c.SharedKeyConfig, validation.When(c.AuthMethod == azureSharedKeyAuthMethod1, validation.Required).Else(validation.Nil)),
		validation.Field(&c.SharedKeyConfig, validation.When(c.AuthMethod == azureSharedKeyAuthMethod2, validation.Required).Else(validation.Nil)),
		validation.Field(&c.SharedKeyConfig, validation.When(c.AuthMethod == azureSharedKeyAuthMethod3, validation.Required).Else(validation.Nil)),
		validation.Field(&c.SharedKeyConfig, validation.When(c.AuthMethod == azureSharedKeyAuthMethod4, validation.Required).Else(validation.Nil)),
		validation.Field(&c.TokenCreds, validation.When(c.AuthMethod == azureTokenAuthMethod, validation.Required).Else(validation.Nil)),
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
		if cfg.SharedKeyConfig == nil {
			return nil, ErrInvalidConfiguration
		}

		if cred, err = azblob.NewSharedKeyCredential(
			cfg.SharedKeyConfig.AccountName,
			cfg.SharedKeyConfig.AccountKey,
		); err != nil {
			return nil, fmt.Errorf("error reading shared key credential: %w", err)
		}
	case azureTokenAuthMethod:
		if cfg.TokenCreds == nil {
			return nil, ErrInvalidConfiguration
		}

		cred = azblob.NewTokenCredential(cfg.TokenCreds.InitialToken, nil)
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

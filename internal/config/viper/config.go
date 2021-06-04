package viper

import (
	"context"
	"errors"
	"math"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	"github.com/spf13/viper"
)

const (
	maxPASETOLifetime = 10 * time.Minute
)

var (
	errNilInput = errors.New("nil input provided")
)

// BuildViperConfig is a constructor function that initializes a viper config.
func BuildViperConfig() *viper.Viper {
	cfg := viper.New()

	// meta stuff.
	cfg.SetDefault(ConfigKeyMetaRunMode, config.DefaultRunMode)
	cfg.SetDefault(ConfigKeyServerStartupDeadline, config.DefaultStartupDeadline)

	// encoding stuff.
	cfg.SetDefault(ConfigKeyEncodingContentType, "application/json")

	// auth stuff.
	cfg.SetDefault(ConfigKeyAuthCookieDomain, authservice.DefaultCookieDomain)
	cfg.SetDefault(ConfigKeyAuthCookieLifetime, authservice.DefaultCookieLifetime)
	cfg.SetDefault(ConfigKeyAuthEnableUserSignup, true)

	// database stuff
	cfg.SetDefault(ConfigKeyDatabaseRunMigrations, true)
	cfg.SetDefault(ConfigKeyAuthMinimumUsernameLength, 4)
	cfg.SetDefault(ConfigKeyAuthMinimumPasswordLength, 8)

	// metrics stuff.
	cfg.SetDefault(ConfigKeyDatabaseMetricsCollectionInterval, metrics.DefaultMetricsCollectionInterval)
	cfg.SetDefault(ConfigKeyMetricsRuntimeCollectionInterval, dbconfig.DefaultMetricsCollectionInterval)

	// tracing stuff.
	cfg.SetDefault(ConfigKeyObservabilityTracingSpanCollectionProbability, 1)

	// audit log stuff.
	cfg.SetDefault(ConfigKeyAuditLogEnabled, true)

	// search stuff
	cfg.SetDefault(ConfigKeySearchProvider, search.BleveProvider)

	// webhooks stuff.
	cfg.SetDefault(ConfigKeyWebhooksEnabled, false)

	// server stuff.
	cfg.SetDefault(ConfigKeyServerHTTPPort, 80)

	return cfg
}

// FromConfig returns a viper instance from a config struct.
func FromConfig(input *config.InstanceConfig) (*viper.Viper, error) {
	if input == nil {
		return nil, errNilInput
	}

	ctx := context.Background()

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, err
	}

	cfg := BuildViperConfig()

	cfg.Set(ConfigKeyMetaDebug, input.Meta.Debug)
	cfg.Set(ConfigKeyMetaRunMode, string(input.Meta.RunMode))

	cfg.Set(ConfigKeyServerStartupDeadline, input.Server.StartupDeadline)
	cfg.Set(ConfigKeyServerHTTPPort, input.Server.HTTPPort)
	cfg.Set(ConfigKeyServerDebug, input.Server.Debug)

	cfg.Set(ConfigKeyEncodingContentType, input.Encoding.ContentType)

	cfg.Set(ConfigKeyFrontendUseFakeData, input.Services.Frontend.UseFakeData)

	cfg.Set(ConfigKeyAuthDebug, input.Services.Auth.Debug)
	cfg.Set(ConfigKeyAuthEnableUserSignup, input.Services.Auth.EnableUserSignup)
	cfg.Set(ConfigKeyAuthMinimumUsernameLength, input.Services.Auth.MinimumUsernameLength)
	cfg.Set(ConfigKeyAuthMinimumPasswordLength, input.Services.Auth.MinimumPasswordLength)

	cfg.Set(ConfigKeyAuthCookieName, input.Services.Auth.Cookies.Name)
	cfg.Set(ConfigKeyAuthCookieDomain, input.Services.Auth.Cookies.Domain)
	cfg.Set(ConfigKeyAuthCookieHashKey, input.Services.Auth.Cookies.HashKey)
	cfg.Set(ConfigKeyAuthCookieSigningKey, input.Services.Auth.Cookies.SigningKey)
	cfg.Set(ConfigKeyAuthCookieLifetime, input.Services.Auth.Cookies.Lifetime)
	cfg.Set(ConfigKeyAuthSecureCookiesOnly, input.Services.Auth.Cookies.SecureOnly)

	cfg.Set(ConfigKeyCapitalismEnabled, input.Capitalism.Enabled)
	cfg.Set(ConfigKeyCapitalismProvider, input.Capitalism.Provider)
	if input.Capitalism.Stripe != nil {
		cfg.Set(ConfigKeyCapitalismStripeAPIKey, input.Capitalism.Stripe.APIKey)
		cfg.Set(ConfigKeyCapitalismStripeSuccessURL, input.Capitalism.Stripe.SuccessURL)
		cfg.Set(ConfigKeyCapitalismStripeCancelURL, input.Capitalism.Stripe.CancelURL)
		cfg.Set(ConfigKeyCapitalismStripeWebhookSecret, input.Capitalism.Stripe.WebhookSecret)
	}

	cfg.Set(ConfigKeyAuthPASETOListener, input.Services.Auth.PASETO.Issuer)
	cfg.Set(ConfigKeyAuthPASETOLifetimeKey, time.Duration(math.Min(float64(input.Services.Auth.PASETO.Lifetime), float64(maxPASETOLifetime))))
	cfg.Set(ConfigKeyAuthPASETOLocalModeKey, input.Services.Auth.PASETO.LocalModeKey)

	cfg.Set(ConfigKeyMetricsProvider, input.Observability.Metrics.Provider)

	cfg.Set(ConfigKeyObservabilityTracingProvider, input.Observability.Tracing.Provider)
	cfg.Set(ConfigKeyObservabilityTracingSpanCollectionProbability, input.Observability.Tracing.SpanCollectionProbability)

	if input.Observability.Tracing.Jaeger != nil {
		cfg.Set(ConfigKeyObservabilityTracingJaegerCollectorEndpoint, input.Observability.Tracing.Jaeger.CollectorEndpoint)
		cfg.Set(ConfigKeyObservabilityTracingJaegerServiceName, input.Observability.Tracing.Jaeger.ServiceName)
	}

	cfg.Set(ConfigKeyMetricsRuntimeCollectionInterval, input.Observability.Metrics.RuntimeMetricsCollectionInterval)
	cfg.Set(ConfigKeyDatabaseDebug, input.Database.Debug)
	cfg.Set(ConfigKeyDatabaseProvider, input.Database.Provider)
	cfg.Set(ConfigKeyDatabaseMaxPingAttempts, input.Database.MaxPingAttempts)
	cfg.Set(ConfigKeyDatabaseConnectionDetails, string(input.Database.ConnectionDetails))

	if input.Database.CreateTestUser != nil {
		cfg.Set(ConfigKeyDatabaseCreateTestUserUsername, input.Database.CreateTestUser.Username)
		cfg.Set(ConfigKeyDatabaseCreateTestUserPassword, input.Database.CreateTestUser.Password)
		cfg.Set(ConfigKeyDatabaseCreateTestUserIsServiceAdmin, input.Database.CreateTestUser.IsServiceAdmin)
		cfg.Set(ConfigKeyDatabaseCreateTestUserHashedPassword, input.Database.CreateTestUser.HashedPassword)
	}

	cfg.Set(ConfigKeyDatabaseRunMigrations, input.Database.RunMigrations)
	cfg.Set(ConfigKeyDatabaseMetricsCollectionInterval, input.Database.MetricsCollectionInterval)

	cfg.Set(ConfigKeySearchProvider, input.Search.Provider)

	cfg.Set(ConfigKeyUploaderProvider, input.Uploads.Storage.Provider)
	cfg.Set(ConfigKeyUploaderDebug, input.Uploads.Debug)

	cfg.Set(ConfigKeyUploaderBucketName, input.Uploads.Storage.BucketName)
	cfg.Set(ConfigKeyUploaderUploadFilename, input.Uploads.Storage.UploadFilenameKey)

	cfg.Set(ConfigKeyAuditLogEnabled, input.Services.AuditLog.Enabled)

	cfg.Set(ConfigKeyWebhooksEnabled, input.Services.Webhooks.Enabled)

	switch {
	case input.Uploads.Storage.AzureConfig != nil:
		cfg.Set(ConfigKeyUploaderProvider, "azure")
		cfg.Set(ConfigKeyUploaderAzureAuthMethod, input.Uploads.Storage.AzureConfig.AuthMethod)
		cfg.Set(ConfigKeyUploaderAzureAccountName, input.Uploads.Storage.AzureConfig.AccountName)
		cfg.Set(ConfigKeyUploaderAzureBucketName, input.Uploads.Storage.AzureConfig.BucketName)
		cfg.Set(ConfigKeyUploaderAzureMaxTries, input.Uploads.Storage.AzureConfig.Retrying.MaxTries)
		cfg.Set(ConfigKeyUploaderAzureTryTimeout, input.Uploads.Storage.AzureConfig.Retrying.TryTimeout)
		cfg.Set(ConfigKeyUploaderAzureRetryDelay, input.Uploads.Storage.AzureConfig.Retrying.RetryDelay)
		cfg.Set(ConfigKeyUploaderAzureMaxRetryDelay, input.Uploads.Storage.AzureConfig.Retrying.MaxRetryDelay)
		if input.Uploads.Storage.AzureConfig != nil {
			cfg.Set(ConfigKeyUploaderAzureRetryReadsFromSecondaryHost, input.Uploads.Storage.AzureConfig.Retrying.RetryReadsFromSecondaryHost)
		}
		cfg.Set(ConfigKeyUploaderAzureTokenCredentialsInitialToken, input.Uploads.Storage.AzureConfig.TokenCredentialsInitialToken)
		cfg.Set(ConfigKeyUploaderAzureSharedKeyAccountKey, input.Uploads.Storage.AzureConfig.SharedKeyAccountKey)

		fallthrough
	case input.Uploads.Storage.GCSConfig != nil:
		cfg.Set(ConfigKeyUploaderProvider, "gcs")
		cfg.Set(ConfigKeyUploaderGCSAccountKeyFilepath, input.Uploads.Storage.GCSConfig.ServiceAccountKeyFilepath)
		cfg.Set(ConfigKeyUploaderGCSScopes, input.Uploads.Storage.GCSConfig.Scopes)
		cfg.Set(ConfigKeyUploaderGCSBucketName, input.Uploads.Storage.GCSConfig.BucketName)
		cfg.Set(ConfigKeyUploaderGCSGoogleAccessID, input.Uploads.Storage.GCSConfig.BlobSettings.GoogleAccessID)

		fallthrough
	case input.Uploads.Storage.S3Config != nil:
		cfg.Set(ConfigKeyUploaderProvider, "s3")
		cfg.Set(ConfigKeyUploaderS3BucketName, input.Uploads.Storage.S3Config.BucketName)

		fallthrough
	case input.Uploads.Storage.FilesystemConfig != nil:
		cfg.Set(ConfigKeyUploaderProvider, "filesystem")
		cfg.Set(ConfigKeyUploaderFilesystemRootDirectory, input.Uploads.Storage.FilesystemConfig.RootDirectory)
	}

	cfg.Set(ConfigKeyItemsSearchIndexPath, input.Services.Items.SearchIndexPath)

	return cfg, nil
}

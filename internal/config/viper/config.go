package viper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	config "gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"

	"github.com/spf13/viper"

	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
)

const (
	maxPASETOLifetime = 10 * time.Minute
)

var (
	errNilInput                            = errors.New("nil input provided")
	errInvalidTestUserRunModeConfiguration = errors.New("requested test user in production run mode")
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
	cfg.SetDefault(ConfigKeyAuthCookieDomain, auth.DefaultCookieDomain)
	cfg.SetDefault(ConfigKeyAuthCookieLifetime, auth.DefaultCookieLifetime)
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
func FromConfig(input *config.ServerConfig) (*viper.Viper, error) {
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

	cfg.Set(ConfigKeyFrontendUseFakeData, input.Frontend.UseFakeData)

	cfg.Set(ConfigKeyAuthDebug, input.Auth.Debug)
	cfg.Set(ConfigKeyAuthEnableUserSignup, input.Auth.EnableUserSignup)
	cfg.Set(ConfigKeyAuthMinimumUsernameLength, input.Auth.MinimumUsernameLength)
	cfg.Set(ConfigKeyAuthMinimumPasswordLength, input.Auth.MinimumPasswordLength)

	cfg.Set(ConfigKeyAuthCookieName, input.Auth.Cookies.Name)
	cfg.Set(ConfigKeyAuthCookieDomain, input.Auth.Cookies.Domain)
	cfg.Set(ConfigKeyAuthCookieHashKey, input.Auth.Cookies.HashKey)
	cfg.Set(ConfigKeyAuthCookieSigningKey, input.Auth.Cookies.SigningKey)
	cfg.Set(ConfigKeyAuthCookieLifetime, input.Auth.Cookies.Lifetime)
	cfg.Set(ConfigKeyAuthSecureCookiesOnly, input.Auth.Cookies.SecureOnly)

	cfg.Set(ConfigKeyAuthPASETOListener, input.Auth.PASETO.Issuer)
	cfg.Set(ConfigKeyAuthPASETOLifetimeKey, time.Duration(math.Min(float64(input.Auth.PASETO.Lifetime), float64(maxPASETOLifetime))))
	cfg.Set(ConfigKeyAuthPASETOLocalModeKey, input.Auth.PASETO.LocalModeKey)

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
	cfg.Set(ConfigKeyItemsSearchIndexPath, string(input.Search.ItemsIndexPath))

	cfg.Set(ConfigKeyUploaderProvider, input.Uploads.Storage.Provider)
	cfg.Set(ConfigKeyUploaderDebug, input.Uploads.Debug)

	cfg.Set(ConfigKeyUploaderBucketName, input.Uploads.Storage.BucketName)
	cfg.Set(ConfigKeyUploaderUploadFilename, input.Uploads.Storage.UploadFilenameKey)

	cfg.Set(ConfigKeyAuditLogEnabled, input.AuditLog.Enabled)
	cfg.Set(ConfigKeyWebhooksEnabled, input.Webhooks.Enabled)

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

	return cfg, nil
}

// ParseConfigFile parses a configuration file.
func ParseConfigFile(ctx context.Context, logger logging.Logger, filePath string) (*config.ServerConfig, error) {
	logger = logger.WithValue("filepath", filePath)
	logger.Debug("parsing config file")

	cfg := BuildViperConfig()
	cfg.SetConfigFile(filePath)

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("trying to read the config file: %w", err)
	}

	var serverConfig *config.ServerConfig
	if err := cfg.Unmarshal(&serverConfig); err != nil {
		return nil, fmt.Errorf("trying to unmarshal the config: %w", err)
	}

	if serverConfig.Database.CreateTestUser != nil && serverConfig.Meta.RunMode == config.ProductionRunMode {
		return nil, errInvalidTestUserRunModeConfiguration
	}

	if validationErr := serverConfig.ValidateWithContext(ctx); validationErr != nil {
		return nil, validationErr
	}

	return serverConfig, nil
}
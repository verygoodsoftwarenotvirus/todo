package viper

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

	"github.com/spf13/viper"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

const (
	// ConfigKeyMetaDebug is the key viper will use to refer to the MetaSettings.debug setting.
	ConfigKeyMetaDebug = "meta.debug"
	// ConfigKeyMetaRunMode is the key viper will use to refer to the MetaSettings.RunMode setting.
	ConfigKeyMetaRunMode = "meta.run_mode"

	// ConfigKeyServerHTTPPort is the key viper will use to refer to the ServerSettings.HTTPPort setting.
	ConfigKeyServerHTTPPort = "server.http_port"
	// ConfigKeyServerStartupDeadline is the key viper will use to refer to the ServerSettings.StartupDeadline setting.
	ConfigKeyServerStartupDeadline = "server.startup_deadline"
	// ConfigKeyServerDebug is the key viper will use to refer to the ServerSettings.debug setting.
	ConfigKeyServerDebug = "server.debug"

	// ConfigKeyFrontendDebug is the key viper will use to refer to the FrontendSettings.debug setting.
	ConfigKeyFrontendDebug = "frontend.debug"
	// ConfigKeyFrontendStaticFilesDir is the key viper will use to refer to the FrontendSettings.StaticFilesDir setting.
	ConfigKeyFrontendStaticFilesDir = "frontend.static_files_directory"
	// ConfigKeyFrontendCacheStatics is the key viper will use to refer to the FrontendSettings.CacheStatics setting.
	ConfigKeyFrontendCacheStatics = "frontend.cache_static_files"

	// ConfigKeyAuthDebug is the key viper will use to refer to the AuthSettings.debug setting.
	ConfigKeyAuthDebug = "auth.debug"
	// ConfigKeyAuthCookieName is the key viper will use to refer to the AuthSettings.CookieName setting.
	ConfigKeyAuthCookieName = "auth.cookies.name"
	// ConfigKeyAuthCookieDomain is the key viper will use to refer to the AuthSettings.CookieDomain setting.
	ConfigKeyAuthCookieDomain = "auth.cookies.domain"
	// ConfigKeyAuthCookieSigningKey is the key viper will use to refer to the AuthSettings.CookieSecret setting.
	ConfigKeyAuthCookieSigningKey = "auth.cookies.signing_key"
	// ConfigKeyAuthCookieLifetime is the key viper will use to refer to the AuthSettings.CookieLifetime setting.
	ConfigKeyAuthCookieLifetime = "auth.cookies.lifetime"
	// ConfigKeyAuthSecureCookiesOnly is the key viper will use to refer to the AuthSettings.SecureCookiesOnly setting.
	ConfigKeyAuthSecureCookiesOnly = "auth.cookies.secure_only"

	// ConfigKeyAuthPASETOAudience is the key for paseto token settings.
	ConfigKeyAuthPASETOAudience = "auth.paseto.audience"
	// ConfigKeyAuthPASETOListener is the key for paseto token settings.
	ConfigKeyAuthPASETOListener = "auth.paseto.listener"
	// ConfigKeyAuthPASETOLocalModeKey is the key for paseto token settings.
	ConfigKeyAuthPASETOLocalModeKey = "auth.paseto.local_mode_key"

	// ConfigKeyAuthEnableUserSignup is the key viper will use to refer to the AuthSettings.nableUserSignup setting.
	ConfigKeyAuthEnableUserSignup = "auth.enable_user_signup"
	// ConfigKeyAuthMinimumUsernameLength is the key viper will use to refer to the AuthSettings.MinimumUsernameLength setting.
	ConfigKeyAuthMinimumUsernameLength = "auth.minimum_username_length"
	// ConfigKeyAuthMinimumPasswordLength is the key viper will use to refer to the AuthSettings.MinimumPasswordLength setting.
	/* #nosec G101 */
	ConfigKeyAuthMinimumPasswordLength = "auth.minimum_password_length"

	// ConfigKeyMetricsProvider is the key viper will use to refer to the MetricsProvider setting.
	ConfigKeyMetricsProvider = "observability.metrics.provider"
	// ConfigKeyMetricsTracer is the key viper will use to refer to the TracingProvider setting.
	ConfigKeyMetricsTracer = "observability.tracing.provider"
	// ConfigKeyObservabilityTracingSpanCollectionProbability is the key viper will use to refer to the SpanCollectionProbability setting.
	ConfigKeyObservabilityTracingSpanCollectionProbability = "observability.tracing.span_collection_probability"
	// ConfigKeyMetricsRuntimeCollectionInterval is the key viper will use to refer to the MetricsSettings.RuntimeCollectionInterval setting.
	ConfigKeyMetricsRuntimeCollectionInterval = "observability.runtime_metrics_collection_interval"

	// ConfigKeyDatabaseDebug is the key viper will use to refer to the DatabaseSettings.debug setting.
	ConfigKeyDatabaseDebug = "database.debug"
	// ConfigKeyDatabaseProvider is the key viper will use to refer to the DatabaseSettings.Provider setting.
	ConfigKeyDatabaseProvider = "database.provider"
	// ConfigKeyDatabaseMaxPingAttempts is the key viper will use to refer to the DatabaseSettings.MaxPingAttempts setting.
	ConfigKeyDatabaseMaxPingAttempts = "database.max_ping_attempts"
	// ConfigKeyDatabaseConnectionDetails is the key viper will use to refer to the DatabaseSettings.ConnectionDetails setting.
	ConfigKeyDatabaseConnectionDetails = "database.connection_details"
	// ConfigKeyDatabaseCreateTestUserUsername is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.Username setting.
	ConfigKeyDatabaseCreateTestUserUsername = "database.create_test_user.username"
	// ConfigKeyDatabaseCreateTestUserPassword is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.Password setting.
	ConfigKeyDatabaseCreateTestUserPassword = "database.create_test_user.password"
	// ConfigKeyDatabaseCreateTestUserIsSiteAdmin is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.IsSiteAdmin setting.
	ConfigKeyDatabaseCreateTestUserIsSiteAdmin = "database.create_test_user.is_site_admin"
	// ConfigKeyDatabaseCreateTestUserHashedPassword is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.HashedPassword setting.
	ConfigKeyDatabaseCreateTestUserHashedPassword = "database.create_test_user.hashed_password"
	// ConfigKeyDatabaseRunMigrations is the key viper will use to refer to the DatabaseSettings.RunMigrations setting.
	ConfigKeyDatabaseRunMigrations = "database.run_migrations"
	// ConfigKeyDatabaseMetricsCollectionInterval is the key viper will use to refer to the database.MetricsCollectionInterval setting.
	ConfigKeyDatabaseMetricsCollectionInterval = "database.metrics_collection_interval"

	// ConfigKeySearchProvider is the key viper will use to refer to the SearchSettings.Provider setting.
	ConfigKeySearchProvider = "search.provider"
	// ConfigKeyItemsSearchIndexPath is the key viper will use to refer to the SearchSettings.ItemsSearchIndexPath setting.
	ConfigKeyItemsSearchIndexPath = "search.items_index_path"

	// ConfigKeyUploaderProvider is the key viper will use to refer to the UploadSettings.Provider value.
	ConfigKeyUploaderProvider = "uploads.storage_config.provider"
	// ConfigKeyUploaderDebug is the key viper will use to refer to the UploadSettings.debug value.
	ConfigKeyUploaderDebug = "uploads.storage_config.debug"
	// ConfigKeyUploaderBucketName is the key viper will use to refer to the UploadSettings.BucketName value.
	ConfigKeyUploaderBucketName = "uploads.storage_config.bucket_name"
	// ConfigKeyUploaderUploadFilename is the key viper will use to refer to the UploadSettings.BucketName value.
	ConfigKeyUploaderUploadFilename = "uploads.storage_config.upload_filename"

	// ConfigKeyUploaderAzureAuthMethod is the key viper will use to refer to UploadSettings.Azure.AuthMethod.
	ConfigKeyUploaderAzureAuthMethod = "uploads.storage_config.azure.auth_method"
	// ConfigKeyUploaderAzureAccountName is the key viper will use to refer to UploadSettings.Azure.AccountName.
	ConfigKeyUploaderAzureAccountName = "uploads.storage_config.azure.account_name"
	// ConfigKeyUploaderAzureBucketName is the key viper will use to refer to UploadSettings.Azure.BucketName.
	ConfigKeyUploaderAzureBucketName = "uploads.storage_config.azure.bucket_name"
	// ConfigKeyUploaderAzureMaxTries is the key viper will use to refer to UploadSettings.Azure.Retrying.MaxTries.
	ConfigKeyUploaderAzureMaxTries = "uploads.storage_config.azure.retrying.max_tries"
	// ConfigKeyUploaderAzureTryTimeout is the key viper will use to refer to UploadSettings.Azure.Retrying.TryTimeout.
	ConfigKeyUploaderAzureTryTimeout = "uploads.storage_config.azure.retrying.try_timeout"
	// ConfigKeyUploaderAzureRetryDelay is the key viper will use to refer to UploadSettings.Azure.Retrying.RetryDelay.
	ConfigKeyUploaderAzureRetryDelay = "uploads.storage_config.azure.retrying.retry_delay"
	// ConfigKeyUploaderAzureMaxRetryDelay is the key viper will use to refer to UploadSettings.Azure.Retrying.MaxRetryDelay.
	ConfigKeyUploaderAzureMaxRetryDelay = "uploads.storage_config.azure.retrying.max_retry_delay"
	// ConfigKeyUploaderAzureRetryReadsFromSecondaryHost is the key viper will use to refer to UploadSettings.Azure.Retrying.RetryReadsFromSecondaryHost.
	ConfigKeyUploaderAzureRetryReadsFromSecondaryHost = "uploads.storage_config.azure.retrying.retry_reads_from_secondary_host"
	// ConfigKeyUploaderAzureTokenCredentialsInitialToken is the key viper will use to refer to UploadSettings.Azure.TokenCredentialsInitialToken.
	ConfigKeyUploaderAzureTokenCredentialsInitialToken = "uploads.storage_config.azure.token_creds_initial_token"
	// ConfigKeyUploaderAzureSharedKeyAccountKey is the key viper will use to refer to UploadSettings.Azure.SharedKeyAccountKey.
	ConfigKeyUploaderAzureSharedKeyAccountKey = "uploads.storage_config.azure.shared_key_account_key"

	// ConfigKeyUploaderGCSAccountKeyFilepath is the key viper will use to refer to UploadSettings.GCS.ServiceAccountKeyFilepath.
	ConfigKeyUploaderGCSAccountKeyFilepath = "uploads.storage_config.gcs.service_account_key_filepath"
	// ConfigKeyUploaderGCSScopes is the key viper will use to refer to UploadSettings.GCS.Scopes.
	ConfigKeyUploaderGCSScopes = "uploads.storage_config.gcs.scopes"
	// ConfigKeyUploaderGCSBucketName is the key viper will use to refer to UploadSettings.GCS.BucketName.
	ConfigKeyUploaderGCSBucketName = "uploads.storage_config.gcs.bucket_name"
	// ConfigKeyUploaderGCSGoogleAccessID is the key viper will use to refer to UploadSettings.GCS.BlobSettingsGoogleAccessID.
	ConfigKeyUploaderGCSGoogleAccessID = "uploads.storage_config.gcs.blob_settings.google_access_id"
	// ConfigKeyUploaderGCSPrivateKeyFilepath is the key viper will use to refer to UploadSettings.GCS.BlobSettings.PrivateKeyFilepath.
	ConfigKeyUploaderGCSPrivateKeyFilepath = "uploads.storage_config.gcs.blob_settings.private_key_filepath"

	// ConfigKeyUploaderS3BucketName is the key viper will use to refer to Uploads.S3.BucketName.
	ConfigKeyUploaderS3BucketName = "uploads.storage_config.s3.bucket_name"

	// ConfigKeyUploaderFilesystemRootDirectory is the key viper will use to refer to Uploads.Filesystem.RootDirectory.
	ConfigKeyUploaderFilesystemRootDirectory = "uploads.storage_config.filesystem.root_directory"

	// ConfigKeyAuditLogEnabled is the key viper will use to refer to the AuditLogSettings.Enabled setting.
	ConfigKeyAuditLogEnabled = "audit_log.enabled"
	// ConfigKeyWebhooksEnabled is the key viper will use to refer to the AuditLogSettings.Enabled setting.
	ConfigKeyWebhooksEnabled = "webhooks.enabled"
)

// BuildViperConfig is a constructor function that initializes a viper config.
func BuildViperConfig() *viper.Viper {
	cfg := viper.New()

	// meta stuff.
	cfg.SetDefault(ConfigKeyMetaRunMode, config.DefaultRunMode)
	cfg.SetDefault(ConfigKeyServerStartupDeadline, config.DefaultStartupDeadline)

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
		return nil, errors.New("nil input provided")
	}

	if err := input.Validate(context.Background()); err != nil {
		return nil, err
	}

	cfg := BuildViperConfig()

	cfg.Set(ConfigKeyMetaDebug, input.Meta.Debug)
	cfg.Set(ConfigKeyMetaRunMode, string(input.Meta.RunMode))

	cfg.Set(ConfigKeyServerStartupDeadline, input.Server.StartupDeadline)
	cfg.Set(ConfigKeyServerHTTPPort, input.Server.HTTPPort)
	cfg.Set(ConfigKeyServerDebug, input.Server.Debug)

	cfg.Set(ConfigKeyFrontendDebug, input.Frontend.Debug)
	cfg.Set(ConfigKeyFrontendStaticFilesDir, input.Frontend.StaticFilesDirectory)
	cfg.Set(ConfigKeyFrontendCacheStatics, input.Frontend.CacheStaticFiles)

	cfg.Set(ConfigKeyAuthDebug, input.Auth.Debug)
	cfg.Set(ConfigKeyAuthEnableUserSignup, input.Auth.EnableUserSignup)
	cfg.Set(ConfigKeyAuthMinimumUsernameLength, input.Auth.MinimumUsernameLength)
	cfg.Set(ConfigKeyAuthMinimumPasswordLength, input.Auth.MinimumPasswordLength)

	cfg.Set(ConfigKeyAuthCookieName, input.Auth.Cookies.Name)
	cfg.Set(ConfigKeyAuthCookieDomain, input.Auth.Cookies.Domain)
	cfg.Set(ConfigKeyAuthCookieSigningKey, input.Auth.Cookies.SigningKey)
	cfg.Set(ConfigKeyAuthCookieLifetime, input.Auth.Cookies.Lifetime)
	cfg.Set(ConfigKeyAuthSecureCookiesOnly, input.Auth.Cookies.SecureOnly)

	cfg.Set(ConfigKeyAuthPASETOAudience, input.Auth.PASETO.Audience)
	cfg.Set(ConfigKeyAuthPASETOListener, input.Auth.PASETO.Listener)
	cfg.Set(ConfigKeyAuthPASETOLocalModeKey, input.Auth.PASETO.LocalModeKey)

	cfg.Set(ConfigKeyMetricsProvider, input.Observability.Metrics.Provider)

	cfg.Set(ConfigKeyMetricsTracer, input.Observability.Tracing.Provider)
	cfg.Set(ConfigKeyObservabilityTracingSpanCollectionProbability, input.Observability.Tracing.SpanCollectionProbability)

	cfg.Set(ConfigKeyMetricsRuntimeCollectionInterval, input.Observability.RuntimeMetricsCollectionInterval)
	cfg.Set(ConfigKeyDatabaseDebug, input.Database.Debug)
	cfg.Set(ConfigKeyDatabaseProvider, input.Database.Provider)
	cfg.Set(ConfigKeyDatabaseMaxPingAttempts, input.Database.MaxPingAttempts)
	cfg.Set(ConfigKeyDatabaseConnectionDetails, string(input.Database.ConnectionDetails))

	if input.Database.CreateTestUser != nil {
		cfg.Set(ConfigKeyDatabaseCreateTestUserUsername, input.Database.CreateTestUser.Username)
		cfg.Set(ConfigKeyDatabaseCreateTestUserPassword, input.Database.CreateTestUser.Password)
		cfg.Set(ConfigKeyDatabaseCreateTestUserIsSiteAdmin, input.Database.CreateTestUser.IsSiteAdmin)
		cfg.Set(ConfigKeyDatabaseCreateTestUserHashedPassword, input.Database.CreateTestUser.HashedPassword)
	}

	cfg.Set(ConfigKeyDatabaseRunMigrations, input.Database.RunMigrations)
	cfg.Set(ConfigKeyDatabaseMetricsCollectionInterval, input.Database.MetricsCollectionInterval)
	cfg.Set(ConfigKeySearchProvider, input.Search.Provider)
	cfg.Set(ConfigKeyItemsSearchIndexPath, string(input.Search.ItemsIndexPath))

	cfg.Set(ConfigKeyUploaderProvider, input.Uploads.Provider)
	cfg.Set(ConfigKeyUploaderDebug, input.Uploads.Debug)

	cfg.Set(ConfigKeyUploaderBucketName, input.Uploads.Storage.BucketName)
	cfg.Set(ConfigKeyUploaderUploadFilename, input.Uploads.Storage.UploadFilename)

	cfg.Set(ConfigKeyAuditLogEnabled, input.AuditLog.Enabled)
	cfg.Set(ConfigKeyWebhooksEnabled, input.Webhooks.Enabled)

	switch {
	case input.Uploads.Storage.AzureConfig != nil:
		cfg.Set(ConfigKeyUploaderProvider, "azure")
		cfg.Set(ConfigKeyUploaderAzureAuthMethod, input.Uploads.Storage.AzureConfig.AuthMethod)
		cfg.Set(ConfigKeyUploaderAzureAccountName, input.Uploads.Storage.AzureConfig.AccountName)
		cfg.Set(ConfigKeyUploaderAzureBucketName, input.Uploads.Storage.AzureConfig.Bucketname)
		cfg.Set(ConfigKeyUploaderAzureMaxTries, input.Uploads.Storage.AzureConfig.Retrying.MaxTries)
		cfg.Set(ConfigKeyUploaderAzureTryTimeout, input.Uploads.Storage.AzureConfig.Retrying.TryTimeout)
		cfg.Set(ConfigKeyUploaderAzureRetryDelay, input.Uploads.Storage.AzureConfig.Retrying.RetryDelay)
		cfg.Set(ConfigKeyUploaderAzureMaxRetryDelay, input.Uploads.Storage.AzureConfig.Retrying.MaxRetryDelay)
		cfg.Set(ConfigKeyUploaderAzureRetryReadsFromSecondaryHost, input.Uploads.Storage.AzureConfig.Retrying.RetryReadsFromSecondaryHost)
		cfg.Set(ConfigKeyUploaderAzureTokenCredentialsInitialToken, input.Uploads.Storage.AzureConfig.TokenCredentialsInitialToken)
		cfg.Set(ConfigKeyUploaderAzureSharedKeyAccountKey, input.Uploads.Storage.AzureConfig.SharedKeyAccountKey)
	case input.Uploads.Storage.GCSConfig != nil:
		cfg.Set(ConfigKeyUploaderProvider, "gcs")
		cfg.Set(ConfigKeyUploaderGCSAccountKeyFilepath, input.Uploads.Storage.GCSConfig.ServiceAccountKeyFilepath)
		cfg.Set(ConfigKeyUploaderGCSScopes, input.Uploads.Storage.GCSConfig.Scopes)
		cfg.Set(ConfigKeyUploaderGCSBucketName, input.Uploads.Storage.GCSConfig.BucketName)
		cfg.Set(ConfigKeyUploaderGCSGoogleAccessID, input.Uploads.Storage.GCSConfig.BlobSettings.GoogleAccessID)
		cfg.Set(ConfigKeyUploaderGCSPrivateKeyFilepath, input.Uploads.Storage.GCSConfig.BlobSettings.PrivateKeyFilepath)
	case input.Uploads.Storage.S3Config != nil:
		cfg.Set(ConfigKeyUploaderProvider, "s3")
		cfg.Set(ConfigKeyUploaderS3BucketName, input.Uploads.Storage.S3Config.BucketName)
	case input.Uploads.Storage.FilesystemConfig != nil:
		cfg.Set(ConfigKeyUploaderProvider, "filesystem")
		cfg.Set(ConfigKeyUploaderFilesystemRootDirectory, input.Uploads.Storage.FilesystemConfig.RootDirectory)
	}

	return cfg, nil
}

var errInvalidTestUserRunModeConfiguration = errors.New("requested test user in production run mode")

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

	if err := serverConfig.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	if serverConfig.Database.CreateTestUser != nil && serverConfig.Meta.RunMode == config.ProductionRunMode {
		return nil, errInvalidTestUserRunModeConfiguration
	}

	if validationErr := serverConfig.Validate(ctx); validationErr != nil {
		return nil, validationErr
	}

	return serverConfig, nil
}

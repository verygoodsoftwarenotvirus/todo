package viper

const (
	x          = "."
	debugKey   = "debug"
	enabledKey = "enabled"

	metaKey = "meta"
	// ConfigKeyMetaDebug is the key viper will use to refer to the MetaSettings.debug setting.
	ConfigKeyMetaDebug = metaKey + x + debugKey
	// ConfigKeyMetaRunMode is the key viper will use to refer to the MetaSettings.RunMode setting.
	ConfigKeyMetaRunMode = metaKey + x + "run_mode"

	serverKey = "server"
	// ConfigKeyServerHTTPPort is the key viper will use to refer to the ServerSettings.HTTPPort setting.
	ConfigKeyServerHTTPPort = serverKey + x + "http_port"
	// ConfigKeyServerStartupDeadline is the key viper will use to refer to the ServerSettings.StartupDeadline setting.
	ConfigKeyServerStartupDeadline = serverKey + x + "startup_deadline"
	// ConfigKeyServerDebug is the key viper will use to refer to the ServerSettings.debug setting.
	ConfigKeyServerDebug = serverKey + x + debugKey

	frontendKey = "frontend"
	// ConfigKeyFrontendDebug is the key viper will use to refer to the FrontendSettings.debug setting.
	ConfigKeyFrontendDebug = frontendKey + x + debugKey
	// ConfigKeyFrontendStaticFilesDir is the key viper will use to refer to the FrontendSettings.StaticFilesDir setting.
	ConfigKeyFrontendStaticFilesDir = frontendKey + x + "static_files_directory"
	// ConfigKeyFrontendCacheStatics is the key viper will use to refer to the FrontendSettings.CacheStatics setting.
	ConfigKeyFrontendCacheStatics = frontendKey + x + "cache_static_files"

	authKey = "auth"
	// ConfigKeyAuthDebug is the key viper will use to refer to the AuthSettings.debug setting.
	ConfigKeyAuthDebug = authKey + x + debugKey
	cookiesKey         = "cookies"
	// ConfigKeyAuthCookieName is the key viper will use to refer to the AuthSettings.CookieName setting.
	ConfigKeyAuthCookieName = authKey + x + cookiesKey + x + "name"
	// ConfigKeyAuthCookieDomain is the key viper will use to refer to the AuthSettings.CookieDomain setting.
	ConfigKeyAuthCookieDomain = authKey + x + cookiesKey + x + "domain"
	// ConfigKeyAuthCookieSigningKey is the key viper will use to refer to the AuthSettings.CookieSecret setting.
	ConfigKeyAuthCookieSigningKey = authKey + x + cookiesKey + x + "signing_key"
	// ConfigKeyAuthCookieLifetime is the key viper will use to refer to the AuthSettings.CookieLifetime setting.
	ConfigKeyAuthCookieLifetime = authKey + x + cookiesKey + x + "lifetime"
	// ConfigKeyAuthSecureCookiesOnly is the key viper will use to refer to the AuthSettings.SecureCookiesOnly setting.
	ConfigKeyAuthSecureCookiesOnly = authKey + x + cookiesKey + x + "secure_only"

	pasetoKey = "paseto"
	// ConfigKeyAuthPASETOAudience is the key for PASETO token settings.
	ConfigKeyAuthPASETOAudience = authKey + x + pasetoKey + x + "audience"
	// ConfigKeyAuthPASETOListener is the key for PASETO token settings.
	ConfigKeyAuthPASETOListener = authKey + x + pasetoKey + x + "listener"
	// ConfigKeyAuthPASETOLocalModeKey is the key for PASETO token settings.
	ConfigKeyAuthPASETOLocalModeKey = authKey + x + pasetoKey + x + "local_mode_key"

	// ConfigKeyAuthEnableUserSignup is the key viper will use to refer to the AuthSettings.EnableUserSignup setting.
	ConfigKeyAuthEnableUserSignup = authKey + x + "enable_user_signup"
	// ConfigKeyAuthMinimumUsernameLength is the key viper will use to refer to the AuthSettings.MinimumUsernameLength setting.
	ConfigKeyAuthMinimumUsernameLength = authKey + x + "minimum_username_length"
	// ConfigKeyAuthMinimumPasswordLength is the key viper will use to refer to the AuthSettings.MinimumPasswordLength setting.
	/* #nosec G101 */
	ConfigKeyAuthMinimumPasswordLength = authKey + x + "minimum_password_length"

	observabilityKey = "observability"
	metricsKey       = "metrics"
	tracingKey       = "tracing"
	// ConfigKeyMetricsProvider is the key viper will use to refer to the MetricsProvider setting.
	ConfigKeyMetricsProvider = observabilityKey + x + metricsKey + x + "provider"
	// ConfigKeyMetricsTracer is the key viper will use to refer to the TracingProvider setting.
	ConfigKeyMetricsTracer = observabilityKey + x + tracingKey + x + "provider"
	// ConfigKeyObservabilityTracingSpanCollectionProbability is the key viper will use to refer to the SpanCollectionProbability setting.
	ConfigKeyObservabilityTracingSpanCollectionProbability = observabilityKey + x + tracingKey + x + "span_collection_probability"
	// ConfigKeyMetricsRuntimeCollectionInterval is the key viper will use to refer to the MetricsSettings.RuntimeCollectionInterval setting.
	ConfigKeyMetricsRuntimeCollectionInterval = observabilityKey + x + "runtime_metrics_collection_interval"

	databaseKey = "database"
	// ConfigKeyDatabaseDebug is the key viper will use to refer to the DatabaseSettings.debug setting.
	ConfigKeyDatabaseDebug = databaseKey + x + debugKey
	// ConfigKeyDatabaseProvider is the key viper will use to refer to the DatabaseSettings.Provider setting.
	ConfigKeyDatabaseProvider = databaseKey + x + "provider"
	// ConfigKeyDatabaseMaxPingAttempts is the key viper will use to refer to the DatabaseSettings.MaxPingAttempts setting.
	ConfigKeyDatabaseMaxPingAttempts = databaseKey + x + "max_ping_attempts"
	// ConfigKeyDatabaseConnectionDetails is the key viper will use to refer to the DatabaseSettings.ConnectionDetails setting.
	ConfigKeyDatabaseConnectionDetails = databaseKey + x + "connection_details"
	createTestUserKey                  = "create_test_user"
	// ConfigKeyDatabaseCreateTestUserUsername is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.Username setting.
	ConfigKeyDatabaseCreateTestUserUsername = databaseKey + x + createTestUserKey + x + "username"
	// ConfigKeyDatabaseCreateTestUserPassword is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.Password setting.
	ConfigKeyDatabaseCreateTestUserPassword = databaseKey + x + createTestUserKey + x + "password"
	// ConfigKeyDatabaseCreateTestUserIsServiceAdmin is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.IsServiceAdmin setting.
	ConfigKeyDatabaseCreateTestUserIsServiceAdmin = databaseKey + x + createTestUserKey + x + "is_site_admin"
	// ConfigKeyDatabaseCreateTestUserHashedPassword is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.HashedPassword setting.
	ConfigKeyDatabaseCreateTestUserHashedPassword = databaseKey + x + createTestUserKey + x + "hashed_password"
	// ConfigKeyDatabaseRunMigrations is the key viper will use to refer to the DatabaseSettings.RunMigrations setting.
	ConfigKeyDatabaseRunMigrations = databaseKey + x + "run_migrations"
	// ConfigKeyDatabaseMetricsCollectionInterval is the key viper will use to refer to the database.MetricsCollectionInterval setting.
	ConfigKeyDatabaseMetricsCollectionInterval = databaseKey + x + "metrics_collection_interval"

	searchKey = "search"
	// ConfigKeySearchProvider is the key viper will use to refer to the SearchSettings.Provider setting.
	ConfigKeySearchProvider = searchKey + x + "provider"
	// ConfigKeyItemsSearchIndexPath is the key viper will use to refer to the SearchSettings.ItemsSearchIndexPath setting.
	ConfigKeyItemsSearchIndexPath = searchKey + x + "items_index_path"

	uploadsKey       = "uploads"
	storageConfigKey = "storage_config"
	// ConfigKeyUploaderProvider is the key viper will use to refer to the UploadSettings.Provider value.
	ConfigKeyUploaderProvider = uploadsKey + x + storageConfigKey + x + "provider"
	// ConfigKeyUploaderDebug is the key viper will use to refer to the UploadSettings.debug value.
	ConfigKeyUploaderDebug = uploadsKey + x + storageConfigKey + x + "debug"
	// ConfigKeyUploaderBucketName is the key viper will use to refer to the UploadSettings.BucketName value.
	ConfigKeyUploaderBucketName = uploadsKey + x + storageConfigKey + x + "bucket_name"
	// ConfigKeyUploaderUploadFilename is the key viper will use to refer to the UploadSettings.BucketName value.
	ConfigKeyUploaderUploadFilename = uploadsKey + x + storageConfigKey + x + "upload_filename"

	azureKey = "azure"
	// ConfigKeyUploaderAzureAuthMethod is the key viper will use to refer to UploadSettings.Azure.AuthMethod.
	ConfigKeyUploaderAzureAuthMethod = uploadsKey + x + storageConfigKey + x + azureKey + x + "auth_method"
	// ConfigKeyUploaderAzureAccountName is the key viper will use to refer to UploadSettings.Azure.AccountName.
	ConfigKeyUploaderAzureAccountName = uploadsKey + x + storageConfigKey + x + azureKey + x + "account_name"
	// ConfigKeyUploaderAzureBucketName is the key viper will use to refer to UploadSettings.Azure.BucketName.
	ConfigKeyUploaderAzureBucketName = uploadsKey + x + storageConfigKey + x + azureKey + x + "bucket_name"
	retryingKey                      = "retrying"
	// ConfigKeyUploaderAzureMaxTries is the key viper will use to refer to UploadSettings.Azure.Retrying.MaxTries.
	ConfigKeyUploaderAzureMaxTries = uploadsKey + x + storageConfigKey + x + azureKey + x + retryingKey + x + "max_tries"
	// ConfigKeyUploaderAzureTryTimeout is the key viper will use to refer to UploadSettings.Azure.Retrying.TryTimeout.
	ConfigKeyUploaderAzureTryTimeout = uploadsKey + x + storageConfigKey + x + azureKey + x + retryingKey + x + "try_timeout"
	// ConfigKeyUploaderAzureRetryDelay is the key viper will use to refer to UploadSettings.Azure.Retrying.RetryDelay.
	ConfigKeyUploaderAzureRetryDelay = uploadsKey + x + storageConfigKey + x + azureKey + x + retryingKey + x + "retry_delay"
	// ConfigKeyUploaderAzureMaxRetryDelay is the key viper will use to refer to UploadSettings.Azure.Retrying.MaxRetryDelay.
	ConfigKeyUploaderAzureMaxRetryDelay = uploadsKey + x + storageConfigKey + x + azureKey + x + retryingKey + x + "max_retry_delay"
	// ConfigKeyUploaderAzureRetryReadsFromSecondaryHost is the key viper will use to refer to UploadSettings.Azure.Retrying.RetryReadsFromSecondaryHost.
	ConfigKeyUploaderAzureRetryReadsFromSecondaryHost = uploadsKey + x + storageConfigKey + x + azureKey + x + retryingKey + x + "retry_reads_from_secondary_host"
	// ConfigKeyUploaderAzureTokenCredentialsInitialToken is the key viper will use to refer to UploadSettings.Azure.TokenCredentialsInitialToken.
	ConfigKeyUploaderAzureTokenCredentialsInitialToken = uploadsKey + x + storageConfigKey + x + azureKey + x + "token_creds_initial_token"
	// ConfigKeyUploaderAzureSharedKeyAccountKey is the key viper will use to refer to UploadSettings.Azure.SharedKeyAccountKey.
	ConfigKeyUploaderAzureSharedKeyAccountKey = uploadsKey + x + storageConfigKey + x + azureKey + x + "shared_key_account_key"

	gcsKey = "gcs"
	// ConfigKeyUploaderGCSAccountKeyFilepath is the key viper will use to refer to UploadSettings.GCS.ServiceAccountKeyFilepath.
	ConfigKeyUploaderGCSAccountKeyFilepath = uploadsKey + x + storageConfigKey + x + gcsKey + x + "service_account_key_filepath"
	// ConfigKeyUploaderGCSScopes is the key viper will use to refer to UploadSettings.GCS.Scopes.
	ConfigKeyUploaderGCSScopes = uploadsKey + x + storageConfigKey + x + gcsKey + x + "scopes"
	// ConfigKeyUploaderGCSBucketName is the key viper will use to refer to UploadSettings.GCS.BucketName.
	ConfigKeyUploaderGCSBucketName = uploadsKey + x + storageConfigKey + x + gcsKey + x + "bucket_name"
	blobSettingsKey                = "blob_settings"
	// ConfigKeyUploaderGCSGoogleAccessID is the key viper will use to refer to UploadSettings.GCS.BlobSettingsGoogleAccessID.
	ConfigKeyUploaderGCSGoogleAccessID = uploadsKey + x + storageConfigKey + x + gcsKey + x + blobSettingsKey + x + "google_access_id"
	// ConfigKeyUploaderGCSPrivateKeyFilepath is the key viper will use to refer to UploadSettings.GCS.BlobSettings.PrivateKeyFilepath.
	ConfigKeyUploaderGCSPrivateKeyFilepath = uploadsKey + x + storageConfigKey + x + gcsKey + x + blobSettingsKey + x + "private_key_filepath"

	s3Key = "s3"
	// ConfigKeyUploaderS3BucketName is the key viper will use to refer to Uploads.S3.BucketName.
	ConfigKeyUploaderS3BucketName = uploadsKey + x + storageConfigKey + x + s3Key + x + "bucket_name"

	filesystemKey = "filesystem"
	// ConfigKeyUploaderFilesystemRootDirectory is the key viper will use to refer to Uploads.Filesystem.RootDirectory.
	ConfigKeyUploaderFilesystemRootDirectory = uploadsKey + x + storageConfigKey + x + filesystemKey + x + "root_directory"

	auditLogKey = "audit_log"
	// ConfigKeyAuditLogEnabled is the key viper will use to refer to the AuditLogSettings.Enabled setting.
	ConfigKeyAuditLogEnabled = auditLogKey + x + enabledKey

	webhooksKey = "webhooks"
	// ConfigKeyWebhooksEnabled is the key viper will use to refer to the AuditLogSettings.Enabled setting.
	ConfigKeyWebhooksEnabled = webhooksKey + x + enabledKey
)

package viper

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
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
	// ConfigKeyMetaDebug is the key viper will use to refer to the MetaSettings.Debug setting.
	ConfigKeyMetaDebug = "meta.debug"
	// ConfigKeyMetaRunMode is the key viper will use to refer to the MetaSettings.RunMode setting.
	ConfigKeyMetaRunMode = "meta.run_mode"
	// ConfigKeyMetaStartupDeadline is the key viper will use to refer to the MetaSettings.StartupDeadline setting.
	ConfigKeyMetaStartupDeadline = "meta.startup_deadline"

	// ConfigKeyServerHTTPPort is the key viper will use to refer to the ServerSettings.HTTPPort setting.
	ConfigKeyServerHTTPPort = "server.http_port"
	// ConfigKeyServerDebug is the key viper will use to refer to the ServerSettings.Debug setting.
	ConfigKeyServerDebug = "server.debug"

	// ConfigKeyFrontendDebug is the key viper will use to refer to the FrontendSettings.Debug setting.
	ConfigKeyFrontendDebug = "frontend.debug"
	// ConfigKeyFrontendStaticFilesDir is the key viper will use to refer to the FrontendSettings.StaticFilesDir setting.
	ConfigKeyFrontendStaticFilesDir = "frontend.static_files_directory"
	// ConfigKeyFrontendCacheStatics is the key viper will use to refer to the FrontendSettings.CacheStatics setting.
	ConfigKeyFrontendCacheStatics = "frontend.cache_static_files"

	// ConfigKeyAuthDebug is the key viper will use to refer to the AuthSettings.Debug setting.
	ConfigKeyAuthDebug = "auth.debug"
	// ConfigKeyAuthCookieName is the key viper will use to refer to the AuthSettings.CookieName setting.
	ConfigKeyAuthCookieName = "auth.cookie_name"
	// ConfigKeyAuthCookieDomain is the key viper will use to refer to the AuthSettings.CookieDomain setting.
	ConfigKeyAuthCookieDomain = "auth.cookie_domain"
	// ConfigKeyAuthCookieSigningKey is the key viper will use to refer to the AuthSettings.CookieSecret setting.
	ConfigKeyAuthCookieSigningKey = "auth.cookie_signing_key"
	// ConfigKeyAuthCookieLifetime is the key viper will use to refer to the AuthSettings.CookieLifetime setting.
	ConfigKeyAuthCookieLifetime = "auth.cookie_lifetime"
	// ConfigKeyAuthSecureCookiesOnly is the key viper will use to refer to the AuthSettings.SecureCookiesOnly setting.
	ConfigKeyAuthSecureCookiesOnly = "auth.secure_cookies_only"
	// ConfigKeyAuthEnableUserSignup is the key viper will use to refer to the AuthSettings.nableUserSignup setting.
	ConfigKeyAuthEnableUserSignup = "auth.enable_user_signup"
	// ConfigKeyAuthMinimumUsernameLength is the key viper will use to refer to the AuthSettings.MinimumUsernameLength setting.
	ConfigKeyAuthMinimumUsernameLength = "auth.minimum_username_length"
	// ConfigKeyAuthMinimumPasswordLength is the key viper will use to refer to the AuthSettings.MinimumPasswordLength setting.
	/* #nosec G101 */
	ConfigKeyAuthMinimumPasswordLength = "auth.minimum_password_length"

	// ConfigKeyMetricsProvider is the key viper will use to refer to the MetricsSettings.MetricsProvider setting.
	ConfigKeyMetricsProvider = "metrics.metrics_provider"
	// ConfigKeyMetricsTracer is the key viper will use to refer to the MetricsSettings.TracingProvider setting.
	ConfigKeyMetricsTracer = "metrics.tracing_provider"
	// ConfigKeyMetricsRuntimeCollectionInterval is the key viper will use to refer to the MetricsSettings.RuntimeCollectionInterval setting.
	ConfigKeyMetricsRuntimeCollectionInterval = "metrics.runtime_metrics_collection_interval"

	// ConfigKeyDatabaseDebug is the key viper will use to refer to the DatabaseSettings.Debug setting.
	ConfigKeyDatabaseDebug = "database.debug"
	// ConfigKeyDatabaseProvider is the key viper will use to refer to the DatabaseSettings.Provider setting.
	ConfigKeyDatabaseProvider = "database.provider"
	// ConfigKeyDatabaseConnectionDetails is the key viper will use to refer to the DatabaseSettings.ConnectionDetails setting.
	ConfigKeyDatabaseConnectionDetails = "database.connection_details"
	// ConfigKeyDatabaseCreateTestUserUsername is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.Username setting.
	ConfigKeyDatabaseCreateTestUserUsername = "database.create_test_user.username"
	// ConfigKeyDatabaseCreateTestUserPassword is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.Password setting.
	ConfigKeyDatabaseCreateTestUserPassword = "database.create_test_user.password"
	// ConfigKeyDatabaseCreateTestUserIsAdmin is the key viper will use to refer to the DatabaseSettings.CreateTestUserConfig.IsAdmin setting.
	ConfigKeyDatabaseCreateTestUserIsAdmin = "database.create_test_user.is_admin"
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
	cfg.SetDefault(ConfigKeyMetaStartupDeadline, config.DefaultStartupDeadline)

	// auth stuff.
	cfg.SetDefault(ConfigKeyAuthCookieDomain, authservice.DefaultCookieDomain)
	cfg.SetDefault(ConfigKeyAuthCookieLifetime, authservice.DefaultCookieLifetime)
	cfg.SetDefault(ConfigKeyAuthEnableUserSignup, true)

	// database stuff
	cfg.SetDefault(ConfigKeyDatabaseRunMigrations, true)
	cfg.SetDefault(ConfigKeyAuthMinimumUsernameLength, 4)
	cfg.SetDefault(ConfigKeyAuthMinimumPasswordLength, 8)

	// metrics stuff.
	cfg.SetDefault(ConfigKeyDatabaseMetricsCollectionInterval, observability.DefaultMetricsCollectionInterval)
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
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	if serverConfig.Database.CreateTestUser != nil && serverConfig.Meta.RunMode == config.ProductionRunMode {
		return nil, errInvalidTestUserRunModeConfiguration
	}

	return serverConfig, nil
}

package viper

import (
	"crypto/rand"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"

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
	// ConfigKeyMetaDebug is the key viper will use to refer to the MetaDebug setting.
	ConfigKeyMetaDebug = "meta.debug"
	// ConfigKeyMetaRunMode is the key viper will use to refer to the MetaRunMode setting.
	ConfigKeyMetaRunMode = "meta.run_mode"
	// ConfigKeyMetaStartupDeadline is the key viper will use to refer to the MetaStartupDeadline setting.
	ConfigKeyMetaStartupDeadline = "meta.startup_deadline"
	// ConfigKeyServerHTTPPort is the key viper will use to refer to the ServerHTTPPort setting.
	ConfigKeyServerHTTPPort = "server.http_port"
	// ConfigKeyServerDebug is the key viper will use to refer to the ServerDebug setting.
	ConfigKeyServerDebug = "server.debug"
	// ConfigKeyFrontendDebug is the key viper will use to refer to the FrontendDebug setting.
	ConfigKeyFrontendDebug = "frontend.debug"
	// ConfigKeyFrontendStaticFilesDir is the key viper will use to refer to the FrontendStaticFilesDir setting.
	ConfigKeyFrontendStaticFilesDir = "frontend.static_files_directory"
	// ConfigKeyFrontendCacheStatics is the key viper will use to refer to the FrontendCacheStatics setting.
	ConfigKeyFrontendCacheStatics = "frontend.cache_static_files"
	// ConfigKeyAuthDebug is the key viper will use to refer to the AuthDebug setting.
	ConfigKeyAuthDebug = "auth.debug"
	// ConfigKeyAuthCookieDomain is the key viper will use to refer to the AuthCookieDomain setting.
	ConfigKeyAuthCookieDomain = "auth.cookie_domain"
	// ConfigKeyAuthCookieSigningKey is the key viper will use to refer to the AuthCookieSecret setting.
	ConfigKeyAuthCookieSigningKey = "auth.cookie_signing_key"
	// ConfigKeyAuthCookieLifetime is the key viper will use to refer to the AuthCookieLifetime setting.
	ConfigKeyAuthCookieLifetime = "auth.cookie_lifetime"
	// ConfigKeyAuthSecureCookiesOnly is the key viper will use to refer to the AuthSecureCookiesOnly setting.
	ConfigKeyAuthSecureCookiesOnly = "auth.secure_cookies_only"
	// ConfigKeyAuthEnableUserSignup is the key viper will use to refer to the AuthEnableUserSignup setting.
	ConfigKeyAuthEnableUserSignup = "auth.enable_user_signup"
	// ConfigKeyMetricsProvider is the key viper will use to refer to the MetricsProvider setting.
	ConfigKeyMetricsProvider = "metrics.metrics_provider"
	// ConfigKeyMetricsTracer is the key viper will use to refer to the MetricsTracer setting.
	ConfigKeyMetricsTracer = "metrics.tracing_provider"
	// ConfigKeyMetricsDBCollectionInterval is the key viper will use to refer to the MetricsDBCollectionInterval setting.
	ConfigKeyMetricsDBCollectionInterval = "metrics.database_metrics_collection_interval"
	// ConfigKeyMetricsRuntimeCollectionInterval is the key viper will use to refer to the MetricsRuntimeCollectionInterval setting.
	ConfigKeyMetricsRuntimeCollectionInterval = "metrics.runtime_metrics_collection_interval"
	// ConfigKeyDatabaseDebug is the key viper will use to refer to the DatabaseDebug setting.
	ConfigKeyDatabaseDebug = "database.debug"
	// ConfigKeyDatabaseProvider is the key viper will use to refer to the DatabaseProvider setting.
	ConfigKeyDatabaseProvider = "database.provider"
	// ConfigKeyDatabaseConnectionDetails is the key viper will use to refer to the DatabaseConnectionDetails setting.
	ConfigKeyDatabaseConnectionDetails = "database.connection_details"
	// ConfigKeyDatabaseCreateTestUserUsername is the key viper will use to refer to the DatabaseCreateTestUserUsername setting.
	ConfigKeyDatabaseCreateTestUserUsername = "database.create_test_user.username"
	// ConfigKeyDatabaseCreateTestUserPassword is the key viper will use to refer to the DatabaseCreateTestUserPassword setting.
	ConfigKeyDatabaseCreateTestUserPassword = "database.create_test_user.password"
	// ConfigKeyDatabaseCreateTestUserIsAdmin is the key viper will use to refer to the DatabaseCreateTestUserIsAdmin setting.
	ConfigKeyDatabaseCreateTestUserIsAdmin = "database.create_test_user.is_admin"
	// ConfigKeyDatabaseRunMigrations is the key viper will use to refer to the DatabaseRunMigrations setting.
	ConfigKeyDatabaseRunMigrations = "database.run_migrations"
	// ConfigKeyItemsSearchIndexPath is the key viper will use to refer to the ItemsSearchIndexPath setting.
	ConfigKeyItemsSearchIndexPath = "search.items_index_path"
	// ConfigKeyAuditLogEnabled is the key viper will use to refer to the AuditLogEnabled setting.
	ConfigKeyAuditLogEnabled = "audit_log.enabled"
	// ConfigKeyWebhooksEnabled is the key viper will use to refer to the AuditLogEnabled setting.
	ConfigKeyWebhooksEnabled = "webhooks.enabled"
)

// BuildViperConfig is a constructor function that initializes a viper config.
func BuildViperConfig() *viper.Viper {
	cfg := viper.New()

	// meta stuff.
	cfg.SetDefault(ConfigKeyMetaRunMode, config.DefaultRunMode)
	cfg.SetDefault(ConfigKeyMetaStartupDeadline, config.DefaultStartupDeadline)

	// auth stuff.
	cfg.SetDefault(ConfigKeyAuthCookieLifetime, config.DefaultCookieLifetime)
	cfg.SetDefault(ConfigKeyAuthEnableUserSignup, true)

	// database stuff
	cfg.SetDefault(ConfigKeyDatabaseRunMigrations, true)

	// metrics stuff.
	cfg.SetDefault(ConfigKeyMetricsDBCollectionInterval, config.DefaultMetricsCollectionInterval)
	cfg.SetDefault(ConfigKeyMetricsRuntimeCollectionInterval, config.DefaultDatabaseMetricsCollectionInterval)

	// audit log stuff.
	cfg.SetDefault(ConfigKeyAuditLogEnabled, true)

	// webhooks stuff.
	cfg.SetDefault(ConfigKeyWebhooksEnabled, true)

	// server stuff.
	cfg.SetDefault(ConfigKeyServerHTTPPort, 80)

	return cfg
}

var errInvalidTestUserRunModeConfiguration = errors.New("requested test user in production run mode")

// ParseConfigFile parses a configuration file.
func ParseConfigFile(logger logging.Logger, filePath string) (*config.ServerConfig, error) {
	logger.WithValue("filepath", filePath).Debug("parsing config file")

	cfg := BuildViperConfig()
	cfg.SetConfigFile(filePath)

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("trying to read the config file: %w", err)
	}

	var serverConfig *config.ServerConfig
	if err := cfg.Unmarshal(&serverConfig); err != nil {
		return nil, fmt.Errorf("trying to unmarshal the config: %w", err)
	}

	if _, ok := config.ValidModes[serverConfig.Meta.RunMode]; !ok {
		return nil, fmt.Errorf("invalid run mode: %q", serverConfig.Meta.RunMode)
	}

	// set the cookie secret to something (relatively) secure if not provided
	if serverConfig.Auth.CookieSigningKey == "" {
		serverConfig.Auth.CookieSigningKey = config.RandString()
	}

	if serverConfig.Database.CreateTestUser != nil && serverConfig.Meta.RunMode == config.ProductionRunMode {
		return nil, errInvalidTestUserRunModeConfiguration
	}

	return serverConfig, nil
}

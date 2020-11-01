package viper

import (
	"crypto/rand"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"

	"github.com/spf13/viper"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

// BuildViperConfig is a constructor function that initializes a viper config.
func BuildViperConfig() *viper.Viper {
	cfg := viper.New()

	// meta stuff.
	cfg.SetDefault("meta.run_mode", config.DefaultRunMode)
	cfg.SetDefault("meta.startup_deadline", config.DefaultStartupDeadline)

	// auth stuff.
	cfg.SetDefault("auth.cookie_lifetime", config.DefaultCookieLifetime)
	cfg.SetDefault("auth.enable_user_signup", true)

	// database stuff
	cfg.SetDefault("database.run_migrations", true)

	// metrics stuff.
	cfg.SetDefault("metrics.database_metrics_collection_interval", config.DefaultMetricsCollectionInterval)
	cfg.SetDefault("metrics.runtime_metrics_collection_interval", config.DefaultDatabaseMetricsCollectionInterval)

	// server stuff.
	cfg.SetDefault("server.http_port", 80)

	return cfg
}

// ParseConfigFile parses a configuration file.
func ParseConfigFile(logger logging.Logger, filePath string) (*config.ServerConfig, error) {
	cfg := BuildViperConfig()

	logger.WithValue("filepath", filePath).Debug("parsing config file")

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
	if serverConfig.Auth.CookieSecret == "" {
		serverConfig.Auth.CookieSecret = config.RandString()
	}

	logger.WithValues(map[string]interface{}{
		"is_nil": serverConfig.Database.CreateTestUser == nil,
		// "username": serverConfig.Database.CreateTestUser.EnsureUsername,
		// "password": serverConfig.Database.CreateTestUser.EnsuredUserPassword,
		// "is_admin": serverConfig.Database.CreateTestUser.IsAdmin,
	}).Debug("CHECK ME CHECK ME CHECK ME CHECK ME CHECK ME")

	return serverConfig, nil
}

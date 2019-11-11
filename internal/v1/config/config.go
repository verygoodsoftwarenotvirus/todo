package config

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"time"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"github.com/spf13/viper"
)

const (
	defaultStartupDeadline                   = time.Minute
	defaultCookieLifetime                    = 24 * time.Hour
	defaultMetricsCollectionInterval         = 2 * time.Second
	defaultDatabaseMetricsCollectionInterval = 2 * time.Second
	randStringSize                           = 32
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

type (
	// MetaSettings is primarily used for development
	MetaSettings struct {
		// Debug enables debug mode service-wide
		// NOTE: this debug should override all other debugs, which is to say, if this is enabled, all of them are enabled.
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// StartupDeadline indicates how long the service can take to spin up. This includes database migrations, configuring services, etc.
		StartupDeadline time.Duration `json:"startup_deadline" mapstructure:"startup_deadline" toml:"startup_deadline,omitempty"`
	}

	// ServerSettings describes the settings pertinent to the HTTP serving portion of the service
	ServerSettings struct {
		// Debug determines if debug logging or other development conditions are active
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// HTTPPort indicates which port to serve HTTP traffic on
		HTTPPort uint16 `json:"http_port" mapstructure:"http_port" toml:"http_port,omitempty"`
	}

	// FrontendSettings describes the settings pertinent to the frontend
	FrontendSettings struct {
		// StaticFilesDirectory indicates which directory contains our static files for the frontend (i.e. CSS/JS/HTML files)
		StaticFilesDirectory string `json:"static_files_directory" mapstructure:"static_files_directory" toml:"static_files_directory,omitempty"`
		// Debug determines if debug logging or other development conditions are active
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// CacheStaticFiles indicates whether or not to load the static files directory into memory via afero's MemMapFs
		CacheStaticFiles bool `json:"cache_static_files" mapstructure:"cache_static_files" toml:"cache_static_files,omitempty"`
	}

	// AuthSettings represents our authentication configuration
	AuthSettings struct {
		// CookieDomain indicates what domain the cookies will have set for them
		CookieDomain string `json:"cookie_domain" mapstructure:"cookie_domain" toml:"cookie_domain,omitempty"`
		// CookieSecret indicates the secret the cookie builder should use
		CookieSecret string `json:"cookie_secret" mapstructure:"cookie_secret" toml:"cookie_secret,omitempty"`
		// CookieLifetime indicates how long the cookies built should last
		CookieLifetime time.Duration `json:"cookie_lifetime" mapstructure:"cookie_lifetime" toml:"cookie_lifetime,omitempty"`
		// Debug determines if debug logging or other development conditions are active
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// SecureCookiesOnly indicates if the cookies built should be marked as HTTPS only
		SecureCookiesOnly bool `json:"secure_cookies_only" mapstructure:"secure_cookies_only" toml:"secure_cookies_only,omitempty"`
		// EnableUserSignup enables user signups
		EnableUserSignup bool `json:"enable_user_signup" mapstructure:"enable_user_signup" toml:"enable_user_signup,omitempty"`
	}

	// DatabaseSettings represents our database configuration
	DatabaseSettings struct {
		// Debug determines if debug logging or other development conditions are active
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// Provider indicates what database we'll connect to (postgres, mysql, etc.)
		Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		// ConnectionDetails indicates how our database driver should connect to the instance
		ConnectionDetails database.ConnectionDetails `json:"connection_details" mapstructure:"connection_details" toml:"connection_details,omitempty"`
	}

	// MetricsSettings contains settings about how we report our metrics
	MetricsSettings struct {
		// MetricsProvider indicates where our metrics should go
		MetricsProvider metricsProvider `json:"metrics_provider" mapstructure:"metrics_provider" toml:"metrics_provider,omitempty"`
		// TracingProvider indicates where our traces should go
		TracingProvider tracingProvider `json:"tracing_provider" mapstructure:"tracing_provider" toml:"tracing_provider,omitempty"`
		// DBMetricsCollectionInterval is the interval we collect database statistics at
		DBMetricsCollectionInterval time.Duration `json:"database_metrics_collection_interval" mapstructure:"database_metrics_collection_interval" toml:"database_metrics_collection_interval,omitempty"`
		// RuntimeMetricsCollectionInterval  is the interval we collect runtime statistics at
		RuntimeMetricsCollectionInterval time.Duration `json:"runtime_metrics_collection_interval" mapstructure:"runtime_metrics_collection_interval" toml:"runtime_metrics_collection_interval,omitempty"`
	}

	// ServerConfig is our server configuration struct. It is comprised of all the other setting structs
	// For information on this structs fields, refer to their definitions
	ServerConfig struct {
		Meta     MetaSettings     `json:"meta" mapstructure:"meta" toml:"meta,omitempty"`
		Frontend FrontendSettings `json:"frontend" mapstructure:"frontend" toml:"frontend,omitempty"`
		Auth     AuthSettings     `json:"auth" mapstructure:"auth" toml:"auth,omitempty"`
		Server   ServerSettings   `json:"server" mapstructure:"server" toml:"server,omitempty"`
		Database DatabaseSettings `json:"database" mapstructure:"database" toml:"database,omitempty"`
		Metrics  MetricsSettings  `json:"metrics" mapstructure:"metrics" toml:"metrics,omitempty"`
	}

	// MarshalFunc is a function that can marshal a config
	MarshalFunc func(v interface{}) ([]byte, error)
)

// EncodeToFile renders your config to a file given your favorite encoder
func (cfg *ServerConfig) EncodeToFile(path string, marshaler MarshalFunc) error {
	byteSlice, err := marshaler(*cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, byteSlice, 0644)
}

// BuildConfig is a constructor function that initializes a viper config.
func BuildConfig() *viper.Viper {
	cfg := viper.New()

	// meta stuff
	cfg.SetDefault("meta.startup_deadline", defaultStartupDeadline)

	// auth stuff
	// NOTE: this will result in an ever-changing cookie secret per server instance running.
	cfg.SetDefault("auth.cookie_secret", randString())
	cfg.SetDefault("auth.cookie_lifetime", defaultCookieLifetime)
	cfg.SetDefault("auth.enable_user_signup", true)

	// metrics stuff
	cfg.SetDefault("metrics.database_metrics_collection_interval", defaultMetricsCollectionInterval)
	cfg.SetDefault("metrics.runtime_metrics_collection_interval", defaultDatabaseMetricsCollectionInterval)

	// server stuff
	cfg.SetDefault("server.http_port", 80)

	return cfg
}

// ParseConfigFile parses a configuration file
func ParseConfigFile(filename string) (*ServerConfig, error) {
	cfg := BuildConfig()
	cfg.SetConfigFile(filename)

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("trying to read the config file: %w", err)
	}

	var serverConfig *ServerConfig
	if err := cfg.Unmarshal(&serverConfig); err != nil {
		return nil, fmt.Errorf("trying to unmarshal the config: %w", err)
	}

	return serverConfig, nil
}

// randString produces a random string
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func randString() string {
	b := make([]byte, randStringSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

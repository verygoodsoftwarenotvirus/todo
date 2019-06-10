package config

import (
	"crypto/rand"
	"encoding/base32"
	"time"

	"github.com/pkg/errors"
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
		Debug bool `mapstructure:"debug"`

		// StartupDeadline indicates how long the service can take to spin up. This includes database migrations, configuring services, etc.
		StartupDeadline time.Duration `mapstructure:"startup_deadline"`
	}

	// ServerSettings describes the settings pertinent to the
	ServerSettings struct {
		// HTTPPort indicates which port to serve HTTP traffic on
		HTTPPort uint16 `mapstructure:"http_port"`
		// Whether or not to enable debug settings for the server
		Debug bool `mapstructure:"debug"`
	}

	// FrontendSettings describes the settings pertinent to the frontend
	FrontendSettings struct {
		// StaticFilesDirectory indicates which directory contains our static files for the frontend (i.e. CSS/JS/HTML files)
		StaticFilesDirectory string `mapstructure:"static_files_dir"`
		// CacheStaticFiles indicates whether or not to load the static files directory into memory via afero's MemMapFs.
		CacheStaticFiles bool `mapstructure:"cache_static_files"`
	}

	// AuthSettings is a container struct for dealing with settings pertaining to
	AuthSettings struct {
		// SecureCookiesOnly indicates if the cookies built should be marked as HTTPS only
		SecureCookiesOnly bool `mapstructure:"secure_cookies_only"`
		// CookieDomain reflects what domain the cookies will have set for them
		CookieDomain string `mapstructure:"cookie_domain"`
		// CookieSecret indicates the secret the cookie builder should use
		CookieSecret string `mapstructure:"cookie_secret"`
		// CookieLifetime indicates how long the cookies built should last
		CookieLifetime time.Duration `mapstructure:"cookie_lifetime"`
		// EnableUserSignup enables user signups
		EnableUserSignup bool `mapstructure:"enable_user_signup"`
	}

	// ServerConfig is our server configuration struct
	ServerConfig struct {
		Meta     MetaSettings     `mapstructure:"meta"`
		Frontend FrontendSettings `mapstructure:"frontend"`
		Auth     AuthSettings     `mapstructure:"auth"`
		Metrics  MetricsSettings  `mapstructure:"metrics"`
		Server   ServerSettings   `mapstructure:"server"`
		Database DatabaseSettings `mapstructure:"database"`
	}
)

func buildConfig() *viper.Viper {
	cfg := viper.New()

	// meta stuff
	cfg.SetDefault("meta.startup_deadline", defaultStartupDeadline)

	// auth stuff
	// NOTE: this will result in an ever-changing cookie secret per server instance running.
	cfg.SetDefault("auth.cookie_secret", randString())
	cfg.SetDefault("auth.cookie_lifetime", defaultCookieLifetime)
	cfg.SetDefault("auth.enable_user_signup", true)

	// metrics stuff
	cfg.SetDefault("metrics.metrics_namespace", MetricsNamespace)
	cfg.SetDefault("metrics.database_metrics_collection_interval", defaultMetricsCollectionInterval)
	cfg.SetDefault("metrics.runtime_metrics_collection_interval", defaultDatabaseMetricsCollectionInterval)

	// server stuff
	cfg.SetDefault("server.http_port", 80)
	cfg.SetDefault("server.metrics_namespace", "todo-server")

	return cfg
}

// ParseConfigFile parses a configuration file
func ParseConfigFile(filename string) (*ServerConfig, error) {
	cfg := buildConfig()

	cfg.SetConfigFile(filename)

	if err := cfg.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "trying to read the config file")
	}

	var serverConfig *ServerConfig
	if err := cfg.Unmarshal(&serverConfig); err != nil {
		return nil, errors.Wrap(err, "trying to unmarshal the config")
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

package config

import (
	"crypto/rand"
	"encoding/base32"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	defaultStartupDeadline = time.Minute
	defaultCookieLifetime = 24 * time.Hour
	randStringSize        = 32
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
		// NOTE: this debug should override all other debugs. That is to say, if it is enabled, all of them are enabled.
		Debug bool `mapstructure:"debug"`
		StartupDeadline time.Duration `mapstructure:"startup_deadline"`
	}

	// ServerSettings describes the settings pertinent to the
	ServerSettings struct {
		HTTPPort uint16 `mapstructure:"http_port"`
		Debug    bool   `mapstructure:"debug"`
	}

	// FrontendSettings describes the settings pertinent to the frontend
	FrontendSettings struct {
		CacheStaticFiles     bool   `mapstructure:"cache_static_files"`
		StaticFilesDirectory string `mapstructure:"static_files_dir"`
		WASMClientPackage    string `mapstructure:"wasm_client_package"`
	}

	// AuthSettings is a container struct for dealing with settings pertaining to
	AuthSettings struct {
		SecureCookiesOnly bool          `mapstructure:"secure_cookies_only"`
		CookieDomain      string        `mapstructure:"cookie_domain"`
		CookieSecret      string        `mapstructure:"cookie_secret"`
		CookieLifetime    time.Duration `mapstructure:"cookie_lifetime"`
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
	cfg.SetDefault("meta.debug", false)
	cfg.SetDefault("meta.startup_deadline", defaultStartupDeadline)

	// auth stuff
	cfg.SetDefault("auth.cookie_secret", randString())
	cfg.SetDefault("auth.cookie_lifetime", defaultCookieLifetime)

	// metrics stuff
	cfg.SetDefault("metrics.metrics_namespace", MetricsNamespace)
	cfg.SetDefault("metrics.database_metrics_collection_interval", 2*time.Second)
	cfg.SetDefault("metrics.runtime_metrics_collection_interval", 2*time.Second)

	// server stuff
	cfg.SetDefault("server.http_port", 80)
	cfg.SetDefault("server.metrics_namespace", "todo-server")

	// database stuff
	cfg.SetDefault("database.debug", false)

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

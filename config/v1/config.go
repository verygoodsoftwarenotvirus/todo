package config

import (
	"crypto/rand"
	"encoding/base32"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

type (
	// MetaSettings is a container struct for dealing with settings pertaining to operations matters for the server.
	MetaSettings struct {
		// NOTE: this debug should override all other debugs. That is to say, if it is enabled, all of them are enabled.
		Debug bool `mapstructure:"debug"`
	}

	// ServerSettings is a container struct for dealing with settings pertaining to
	ServerSettings struct {
		HTTPPort               uint16 `mapstructure:"http_port"`
		Debug                  bool   `mapstructure:"debug"`
		FrontendFilesDirectory string `mapstructure:"frontend_files_directory"`
	}

	// AuthSettings is a container struct for dealing with settings pertaining to
	AuthSettings struct {
		CookieSecret string `mapstructure:"cookie_secret"`
	}
)

// ServerConfig is our server configuration struct
type ServerConfig struct {
	Meta     MetaSettings     `mapstructure:"meta"`
	Auth     AuthSettings     `mapstructure:"auth"`
	Metrics  MetricsSettings  `mapstructure:"metrics"`
	Server   ServerSettings   `mapstructure:"server"`
	Database DatabaseSettings `mapstructure:"database"`
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

func buildConfig() *viper.Viper {
	cfg := viper.New()

	// meta stuff
	cfg.SetDefault("meta.debug", false)

	// auth stuff
	cfg.SetDefault("auth.cookie_secret", randString())

	// metrics stuff
	cfg.SetDefault("metrics.metrics_namespace", "todo_service")
	cfg.SetDefault("metrics.database_metrics_collection_interval", time.Second)
	cfg.SetDefault("metrics.runtime_metrics_collection_interval", time.Second)

	// server stuff
	cfg.SetDefault("server.http_port", 80)
	cfg.SetDefault("server.metrics_namespace", "todo-server")

	// database stuff
	cfg.SetDefault("database.debug", false)

	return cfg
}

// randString produces a random string
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func randString() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	rs := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return rs
}

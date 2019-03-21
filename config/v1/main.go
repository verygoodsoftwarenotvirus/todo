package config

import (
	"crypto/rand"
	"encoding/base32"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	// Providers represents this package's offering to the dependency manager
	Providers = wire.NewSet(
		ProvideConfigServerSettings,
		ProvideConfigAuthSettings,
		ProvideConfigDatabaseSettings,
	)
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

// BEGIN it'd be neat if wire could do this for me one day.

// ProvideConfigServerSettings is an obligatory function that
// we're required to have because wire doesn't do it for us.
func ProvideConfigServerSettings(c *ServerConfig) ServerSettings {
	return c.Server
}

// ProvideConfigAuthSettings is an obligatory function that
// we're required to have because wire doesn't do it for us.
func ProvideConfigAuthSettings(c *ServerConfig) AuthSettings {
	return c.Auth
}

// ProvideConfigDatabaseSettings is an obligatory function that
//  we're required to have because wire doesn't do it for us.
func ProvideConfigDatabaseSettings(c *ServerConfig) DatabaseSettings {
	return c.Database
}

// END it'd be neat if wire could do this for me one day.

type (
	// MetaSettings is a container struct for dealing with settings pertaining to operations matters for the server.
	MetaSettings struct {
		Debug bool // NOTE: this debug will override all other debugs. That is to say, if it is enabled, all of them are enabled
	}

	// ServerSettings is a container struct for dealing with settings pertaining to
	ServerSettings struct {
		Port                   uint16
		Debug                  bool
		FrontendFilesDirectory string            `mapstructure:"frontend_files_directory"`
		MetricsNamespace       metrics.Namespace `mapstructure:"metrics_namespace"`
	}

	// DatabaseSettings is a container struct for dealing with settings pertaining to
	DatabaseSettings struct {
		Type              string
		Debug             bool
		ConnectionDetails database.ConnectionDetails `mapstructure:"connection_details"`
	}

	// AuthSettings is a container struct for dealing with settings pertaining to
	AuthSettings struct {
		CookieSecret string `mapstructure:"cookie_secret"`
	}
)

// ServerConfig is our server configuration struct
type ServerConfig struct {
	Meta     MetaSettings
	Server   ServerSettings
	Auth     AuthSettings
	Database DatabaseSettings
}

// ParseConfigFile parses a configuration
func ParseConfigFile(filename string) (*ServerConfig, error) {
	cfg := viper.New()
	setDefaults(cfg)

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

// ProvideDatabase provides a database base
func (cfg *ServerConfig) ProvideDatabase(logger logging.Logger) (database.Database, error) {
	tracer := tracing.ProvideTracer("database-client")
	debug := cfg.Database.Debug || cfg.Meta.Debug
	connectionDetails := cfg.Database.ConnectionDetails

	switch cfg.Database.Type {
	case "postgres":
		rawDB, err := postgres.ProvidePostgresDB(logger, connectionDetails)
		if err != nil {
			return nil, errors.Wrap(err, "establish postgres database connection")
		}
		pg := postgres.ProvidePostgres(debug, rawDB, logger, connectionDetails)

		return dbclient.ProvideDatabaseClient(pg, debug, logger, tracer)
	case "sqlite":
		sqliteDB, err := sqlite.ProvideSqlite(debug, logger, connectionDetails)
		if err != nil {
			return nil, errors.Wrap(err, "establish postgres database connection")
		}

		return dbclient.ProvideDatabaseClient(sqliteDB, debug, logger, tracer)
	default:
		sqliteDB, err := sqlite.ProvideSqlite(debug, logger, ":memory:")
		if err != nil {
			return nil, errors.Wrap(err, "establish postgres database connection")
		}

		return dbclient.ProvideDatabaseClient(sqliteDB, debug, logger, tracer)
	}
}

func setDefaults(cfg *viper.Viper) {
	// server stuff
	cfg.SetDefault("server.port", 8080)
	cfg.SetDefault("server.debug", false)
	cfg.SetDefault("server.metrics_namespace", "todo-server")

	// database stuff
	cfg.SetDefault("database.type", "sqlite")
	cfg.SetDefault("database.debug", false)

	// auth stuff
	cfg.SetDefault("auth.cookie_secret", randString())
}

// randString produces a random string
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func randString() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	rs := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return rs
}

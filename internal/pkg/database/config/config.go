package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/postgres"
	zqlite "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// PostgresProvider is the string used to refer to postgres.
	PostgresProvider = "postgres"
	// MariaDBProvider is the string used to refer to mariaDB.
	MariaDBProvider = "mariadb"
	// SqliteProvider is the string used to refer to sqlite.
	SqliteProvider = "sqlite"

	// DefaultMetricsCollectionInterval is the default amount of time we wait between database metrics queries.
	DefaultMetricsCollectionInterval = 2 * time.Second
)

var (
	errInvalidDatabase = errors.New("invalid database")
	errNilDBProvided   = errors.New("invalid DB connection provided")
)

type (
	// Config represents our database configuration.
	Config struct {
		CreateTestUser            *types.TestUserCreationConfig `json:"create_test_user" mapstructure:"create_test_user" toml:"create_test_user,omitempty"`
		Provider                  string                        `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		ConnectionDetails         database.ConnectionDetails    `json:"connection_details" mapstructure:"connection_details" toml:"connection_details,omitempty"`
		MetricsCollectionInterval time.Duration                 `json:"metrics_collection_interval" mapstructure:"metrics_collection_interval" toml:"metrics_collection_interval,omitempty"`
		Debug                     bool                          `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		RunMigrations             bool                          `json:"run_migrations" mapstructure:"run_migrations" toml:"run_migrations,omitempty"`
		MaxPingAttempts           uint8                         `json:"max_ping_attempts" mapstructure:"max_ping_attempts" toml:"max_ping_attempts,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates an DatabaseSettings struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.ConnectionDetails, validation.Required),
		validation.Field(&cfg.Provider, validation.In(PostgresProvider, MariaDBProvider, SqliteProvider)),
		validation.Field(&cfg.CreateTestUser, validation.When(cfg.CreateTestUser != nil, validation.Required).Else(validation.Nil)),
	)
}

// ProvideDatabaseConnection provides a database implementation dependent on the configuration.
func (cfg *Config) ProvideDatabaseConnection(logger logging.Logger) (*sql.DB, error) {
	switch cfg.Provider {
	case PostgresProvider:
		return postgres.ProvidePostgresDB(logger, cfg.ConnectionDetails)
	case MariaDBProvider:
		return mariadb.ProvideMariaDBConnection(logger, cfg.ConnectionDetails)
	case SqliteProvider:
		return zqlite.ProvideSqliteDB(logger, cfg.ConnectionDetails, cfg.MetricsCollectionInterval)
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidDatabase, cfg.Provider)
	}
}

// ProvideDatabasePlaceholderFormat provides .
func (cfg *Config) ProvideDatabasePlaceholderFormat() (squirrel.PlaceholderFormat, error) {
	switch cfg.Provider {
	case PostgresProvider:
		return squirrel.Dollar, nil
	case MariaDBProvider, SqliteProvider:
		return squirrel.Question, nil
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidDatabase, cfg.Provider)
	}
}

// ProvideJSONPluckQuery provides a query for extracting a value out of a JSON dictionary for a given database.
func (cfg *Config) ProvideJSONPluckQuery() string {
	switch cfg.Provider {
	case PostgresProvider:
		return `%s.%s->'%s'`
	case MariaDBProvider:
		return `JSON_CONTAINS(%s.%s, '%d', '$.%s')`
	case SqliteProvider:
		return `json_extract(%s.%s, '$.%s')`
	default:
		return ""
	}
}

// ProvideCurrentUnixTimestampQuery provides a database implementation dependent on the configuration.
func (cfg *Config) ProvideCurrentUnixTimestampQuery() string {
	switch cfg.Provider {
	case PostgresProvider:
		return `extract(epoch FROM NOW())`
	case MariaDBProvider:
		return `UNIX_TIMESTAMP()`
	case SqliteProvider:
		return `(strftime('%s','now'))`
	default:
		return ""
	}
}

// ProvideSessionManager provides a session manager based on some settings.
// There's not a great place to put this function. I don't think it belongs in Auth because it accepts a DB connection,
// but it obviously doesn't belong in the database package, or maybe it does.
func ProvideSessionManager(cookieConfig authservice.CookieConfig, dbConf Config, db *sql.DB) (*scs.SessionManager, error) {
	sessionManager := scs.New()

	if db == nil {
		return nil, errNilDBProvided
	}

	switch dbConf.Provider {
	case PostgresProvider:
		sessionManager.Store = postgresstore.New(db)
	case MariaDBProvider:
		sessionManager.Store = mysqlstore.New(db)
	case SqliteProvider:
		sessionManager.Store = sqlite3store.New(db)
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidDatabase, dbConf.Provider)
	}

	sessionManager.Lifetime = cookieConfig.Lifetime
	sessionManager.Lifetime = cookieConfig.Lifetime
	sessionManager.Cookie.Name = cookieConfig.Name
	sessionManager.Cookie.Domain = cookieConfig.Domain
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Secure = cookieConfig.SecureOnly

	return sessionManager, nil
}

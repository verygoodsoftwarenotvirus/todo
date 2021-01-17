package config

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	sqlite "github.com/mattn/go-sqlite3"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers/postgres"
	zqlite "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	postgresql "github.com/lib/pq"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	// PostgresProviderKey is the string we use to refer to postgres.
	PostgresProviderKey = "postgres"
	// MariaDBProviderKey is the string we use to refer to mariaDB.
	MariaDBProviderKey = "mariadb"
	// SqliteProviderKey is the string we use to refer to sqlite.
	SqliteProviderKey = "sqlite"

	// DefaultMetricsCollectionInterval is the default amount of time we wait between database metrics queries.
	DefaultMetricsCollectionInterval = 2 * time.Second
)

// Config represents our database configuration.
type Config struct {
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// RunMigrations determines if we should migrate the database.
	RunMigrations bool `json:"run_migrations" mapstructure:"run_migrations" toml:"run_migrations,omitempty"`
	// CreateTestUser determines if we should create a test user. Doesn't occur if RunMigrations is false.
	CreateTestUser *types.TestUserCreationConfig `json:"create_test_user" mapstructure:"create_test_user" toml:"create_test_user,omitempty"`
	// Provider indicates what database we'll connect to (postgres, mysql, etc.)
	Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
	// ConnectionDetails indicates how our database driver should connect to the instance.
	ConnectionDetails database.ConnectionDetails `json:"connection_details" mapstructure:"connection_details" toml:"connection_details,omitempty"`
	// MetricsCollectionInterval defines the interval at which we poll for metrics.
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval" mapstructure:"metrics_collection_interval" toml:"metrics_collection_interval,omitempty"`
}

// Validate validates an DatabaseSettings struct.
func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.CreateTestUser),
		validation.Field(&cfg.Provider, validation.In(PostgresProviderKey, MariaDBProviderKey, SqliteProviderKey)),
		validation.Field(&cfg.ConnectionDetails, validation.Required),
		validation.Field(&cfg.CreateTestUser, validation.When(cfg.CreateTestUser != nil, validation.Required).Else(validation.Nil)),
	)
}

var errNilDatabaseConnection = errors.New("nil DB connection provided")

// ProvideDatabaseConnection provides a database implementation dependent on the configuration.
func (cfg *Config) ProvideDatabaseConnection(logger logging.Logger) (*sql.DB, error) {
	switch cfg.Provider {
	case PostgresProviderKey:
		return postgres.ProvidePostgresDB(logger, cfg.ConnectionDetails)
	case MariaDBProviderKey:
		return mariadb.ProvideMariaDBConnection(logger, cfg.ConnectionDetails)
	case SqliteProviderKey:
		return zqlite.ProvideSqliteDB(logger, cfg.ConnectionDetails, cfg.MetricsCollectionInterval)
	default:
		return nil, fmt.Errorf("invalid database type selected: %q", cfg.Provider)
	}
}

// ProvideDatabaseClient provides a database implementation dependent on the configuration.
func (cfg *Config) ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	rawDB *sql.DB,
) (database.DataManager, error) {
	if rawDB == nil {
		return nil, errNilDatabaseConnection
	}

	var dbManager database.DataManager

	switch cfg.Provider {
	case PostgresProviderKey:
		dbManager = postgres.ProvidePostgres(cfg.Debug, rawDB, logger)
	case MariaDBProviderKey:
		dbManager = mariadb.ProvideMariaDB(cfg.Debug, rawDB, logger)
	case SqliteProviderKey:
		dbManager = zqlite.ProvideSqlite(cfg.Debug, rawDB, logger)
	default:
		return nil, fmt.Errorf("invalid database type selected: %q", cfg.Provider)
	}

	return dbclient.ProvideDatabaseClient(
		ctx,
		logger,
		dbManager,
		rawDB,
		cfg.CreateTestUser,
		cfg.RunMigrations,
		cfg.Debug,
	)
}

// ProvideDatabaseDriver provides .
func (cfg *Config) ProvideDatabaseDriver() driver.Driver {
	switch cfg.Provider {
	case PostgresProviderKey:
		return &postgresql.Driver{}
	case MariaDBProviderKey:
		return &mysql.MySQLDriver{}
	case SqliteProviderKey:
		return &sqlite.SQLiteDriver{}
	default:
		panic("aaaaaaaaaaaaaaaaaaaaaaaa")
	}
}

// ProvideDatabasePlaceholderFormat provides .
func (cfg *Config) ProvideDatabasePlaceholderFormat() (squirrel.PlaceholderFormat, error) {
	switch cfg.Provider {
	case PostgresProviderKey:
		return squirrel.Dollar, nil
	case MariaDBProviderKey, SqliteProviderKey:
		return squirrel.Question, nil
	default:
		return nil, fmt.Errorf("invalid database type selected: %q", cfg.Provider)
	}
}

// ProvideJSONPluckQuery provides a query for extracting a value out of a JSON dictionary for a given database.
func (cfg *Config) ProvideJSONPluckQuery() string {
	switch cfg.Provider {
	case PostgresProviderKey:
		return `%s.%s->'%s'`
	case MariaDBProviderKey:
		return `JSON_CONTAINS(%s.%s, '%d', '$.%s')`
	case SqliteProviderKey:
		return `json_extract(%s.%s, '$.%s')`
	default:
		return ""
	}
}

// ProvideCurrentUnixTimestampQuery provides a database implementation dependent on the configuration.
func (cfg *Config) ProvideCurrentUnixTimestampQuery() string {
	switch cfg.Provider {
	case PostgresProviderKey:
		return `extract(epoch FROM NOW())`
	case MariaDBProviderKey:
		return `UNIX_TIMESTAMP()`
	case SqliteProviderKey:
		return `(strftime('%s','now'))`
	default:
		return ""
	}
}

// ProvideSessionManager provides a session manager based on some settings.
// There's not a great place to put this function. I don't think it belongs in Auth because it accepts a DB connection,
// but it obviously doesn't belong in the database package, or maybe it does.
func ProvideSessionManager(authConfig authservice.Config, dbConf Config, db *sql.DB) (*scs.SessionManager, error) {
	sessionManager := scs.New()

	if db == nil {
		return nil, errors.New("invalid DB connection provided")
	}

	switch dbConf.Provider {
	case PostgresProviderKey:
		sessionManager.Store = postgresstore.New(db)
	case MariaDBProviderKey:
		sessionManager.Store = mysqlstore.New(db)
	case SqliteProviderKey:
		sessionManager.Store = sqlite3store.New(db)
	default:
		return nil, fmt.Errorf("invalid database provider: %q", dbConf.Provider)
	}

	sessionManager.Lifetime = authConfig.CookieLifetime
	// elaborate further here later if you so choose

	return sessionManager, nil
}

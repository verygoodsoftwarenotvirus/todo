package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/sqlite3store"
	scs "github.com/alexedwards/scs/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	// PostgresProviderKey is the string we use to refer to postgres.
	PostgresProviderKey = "postgres"
	// MariaDBProviderKey is the string we use to refer to mariaDB.
	MariaDBProviderKey = "mariadb"
	// SqliteProviderKey is the string we use to refer to sqlite.
	SqliteProviderKey = "sqlite"

	// DefaultDatabaseMetricsCollectionInterval is the default amount of time we wait between database metrics queries.
	DefaultDatabaseMetricsCollectionInterval = 2 * time.Second
)

// CreateTestUserSettings defines a test user created via config declaration.
type CreateTestUserSettings struct {
	// Username defines our test user's username we create in the event we create them.
	Username string `json:"username" mapstructure:"username" toml:"username,omitempty"`
	// Password defines our test user's password we create in the event we create them.
	Password string `json:"password" mapstructure:"password" toml:"password,omitempty"`
	// IsAdmin defines our test user's admin status we create in the event we create them.
	IsAdmin bool `json:"is_admin" mapstructure:"is_admin" toml:"is_admin,omitempty"`
}

// UserCreationConfig is a helper method for getting around circular imports.
func (s *CreateTestUserSettings) UserCreationConfig() *database.UserCreationConfig {
	if s == nil {
		return nil
	}

	return &database.UserCreationConfig{
		Username: s.Username,
		Password: s.Password,
		IsAdmin:  s.IsAdmin,
	}
}

// DatabaseSettings represents our database configuration.
type DatabaseSettings struct {
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// RunMigrations determines if we should migrate the database.
	RunMigrations bool `json:"run_migrations" mapstructure:"run_migrations" toml:"run_migrations,omitempty"`
	// CreateTestUser determines if we should create a test user. Doesn't occur if RunMigrations is false.
	CreateTestUser *CreateTestUserSettings `json:"create_test_user" mapstructure:"create_test_user" toml:"create_test_user,omitempty"`
	// Provider indicates what database we'll connect to (postgres, mysql, etc.)
	Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
	// ConnectionDetails indicates how our database driver should connect to the instance.
	ConnectionDetails database.ConnectionDetails `json:"connection_details" mapstructure:"connection_details" toml:"connection_details,omitempty"`
}

// ProvideDatabaseConnection provides a database implementation dependent on the configuration.
func (cfg *ServerConfig) ProvideDatabaseConnection(logger logging.Logger) (*sql.DB, error) {
	switch cfg.Database.Provider {
	case PostgresProviderKey:
		return postgres.ProvidePostgresDB(logger, cfg.Database.ConnectionDetails)
	case MariaDBProviderKey:
		return mariadb.ProvideMariaDBConnection(logger, cfg.Database.ConnectionDetails)
	case SqliteProviderKey:
		return sqlite.ProvideSqliteDB(logger, cfg.Database.ConnectionDetails)
	default:
		return nil, fmt.Errorf("invalid database type selected: %q", cfg.Database.Provider)
	}
}

// ProvideDatabaseClient provides a database implementation dependent on the configuration.
func (cfg *ServerConfig) ProvideDatabaseClient(ctx context.Context, logger logging.Logger, rawDB *sql.DB, authenticator auth.Authenticator) (database.DataManager, error) {
	if rawDB == nil {
		return nil, errors.New("nil DB connection provided")
	}

	debug := cfg.Database.Debug || cfg.Meta.Debug

	ocsql.RegisterAllViews()
	ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

	var dbc database.DataManager

	switch cfg.Database.Provider {
	case PostgresProviderKey:
		dbc = postgres.ProvidePostgres(debug, rawDB, logger)
	case MariaDBProviderKey:
		dbc = mariadb.ProvideMariaDB(debug, rawDB, logger)
	case SqliteProviderKey:
		dbc = sqlite.ProvideSqlite(debug, rawDB, logger)
	default:
		return nil, fmt.Errorf("invalid database type selected: %q", cfg.Database.Provider)
	}

	return dbclient.ProvideDatabaseClient(
		ctx,
		logger,
		dbc,
		rawDB,
		authenticator,
		cfg.Database.CreateTestUser.UserCreationConfig(),
		cfg.Database.RunMigrations,
		debug,
	)
}

// ProvideSessionManager provides a session manager based on some settings.
// There's not a great place to put this function. I don't think it belongs in Auth because it accepts a DB connection,
// but it obviously doesn't belong in the database package, or maybe it does
func ProvideSessionManager(authConf AuthSettings, dbConf DatabaseSettings, db *sql.DB) *scs.SessionManager {
	sessionManager := scs.New()

	switch dbConf.Provider {
	case PostgresProviderKey:
		sessionManager.Store = postgresstore.New(db)
	case MariaDBProviderKey:
		sessionManager.Store = mysqlstore.New(db)
	case SqliteProviderKey:
		sessionManager.Store = sqlite3store.New(db)
	}

	sessionManager.Lifetime = authConf.CookieLifetime
	// elaborate further here later if you so choose

	return sessionManager
}

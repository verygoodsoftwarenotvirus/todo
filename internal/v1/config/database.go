package config

import (
	"context"
	"database/sql"
	"errors"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/sqlite"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	// PostgresProviderKey is the string we use to refer to postgres
	PostgresProviderKey = "postgres"
	// MariaDBProviderKey is the string we use to refer to mariaDB
	MariaDBProviderKey = "mariadb"
	// SqliteProviderKey is the string we use to refer to sqlite
	SqliteProviderKey = "sqlite"
)

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
		return nil, errors.New("invalid database type selected")
	}
}

// ProvideDatabaseClient provides a database implementation dependent on the configuration.
func (cfg *ServerConfig) ProvideDatabaseClient(ctx context.Context, logger logging.Logger, rawDB *sql.DB) (database.Database, error) {
	if rawDB == nil {
		return nil, errors.New("nil DB connection provided")
	}

	debug := cfg.Database.Debug || cfg.Meta.Debug

	ocsql.RegisterAllViews()
	ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

	var dbc database.Database
	switch cfg.Database.Provider {
	case PostgresProviderKey:
		dbc = postgres.ProvidePostgres(debug, rawDB, logger)
	case MariaDBProviderKey:
		dbc = mariadb.ProvideMariaDB(debug, rawDB, logger)
	case SqliteProviderKey:
		dbc = sqlite.ProvideSqlite(debug, rawDB, logger)
	default:
		return nil, errors.New("invalid database type selected")
	}

	return dbclient.ProvideDatabaseClient(ctx, rawDB, dbc, debug, logger)
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
	// elaborate further here later

	return sessionManager
}

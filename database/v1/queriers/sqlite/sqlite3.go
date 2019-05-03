package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/mattn/go-sqlite3"
)

const (
	sqliteDriverName = "wrapped-sqlite-driver"
)

func init() {
	// Explicitly wrap the SQLite3 driver with ocsql.
	driver := ocsql.Wrap(&sqlite3.SQLiteDriver{}, ocsql.WithQuery(true))

	// Register our ocsql wrapper as a database driver.
	sql.Register(sqliteDriverName, driver)

}

type (
	// Sqlite is our main Sqlite3 interaction database
	Sqlite struct {
		debug    bool
		database *sql.DB
		logger   logging.Logger
	}
)

// ProvideSqliteDB provides a raw sql.DB object connected to a sqlite3 database
func ProvideSqliteDB(sqliteFilepath database.ConnectionDetails) (*sql.DB, error) {
	return sql.Open(sqliteDriverName, string(sqliteFilepath))
}

// ProvideSqlite provides a sqlite database controller
func ProvideSqlite(debug bool, logger logging.Logger, db *sql.DB) database.Database {
	s := &Sqlite{
		database: db,
		debug:    debug,
		logger:   logger.WithName("sqlite"),
	}

	return s
}

// IsReady reports whether or not Sqlite is ready to be written to.
// Since sqlite3 is a file-based database, it is always ready (as far as I can tell)
func (s *Sqlite) IsReady(ctx context.Context) bool {
	return s.database.PingContext(ctx) == nil
}

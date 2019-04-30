package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	_ "github.com/mattn/go-sqlite3" // for the init import call
)

type (
	// Sqlite is our main Sqlite3 interaction database
	Sqlite struct {
		debug    bool
		database *sql.DB
		logger   logging.Logger
	}

	// Scannable represents any database response (i.e. either a transaction or a regular execution response)
	Scannable interface {
		Scan(dest ...interface{}) error
	}
)

// ProvideSqlite provides a sqlite database controller
func ProvideSqlite(
	debug bool,
	logger logging.Logger,
	sqliteFilepath database.ConnectionDetails,
) (database.Database, error) {
	logger.WithValue("sqlite_filepath", sqliteFilepath).Debug("Establishing connection to sqlite3 file")
	db, err := sql.Open("sqlite3", string(sqliteFilepath))
	if err != nil {
		logger.Error(err, "error encountered establishing database connection")
		return nil, err
	}

	s := &Sqlite{
		debug:    debug,
		logger:   logger.WithName("sqlite"),
		database: db,
	}

	return s, nil
}

// IsReady reports whether or not Sqlite is ready to be written to.
// Since sqlite3 is a file-based database, it is always ready AFAICT
func (s *Sqlite) IsReady(ctx context.Context) (ready bool) {
	return true
}

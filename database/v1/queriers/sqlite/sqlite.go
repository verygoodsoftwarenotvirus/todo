package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	sqlite "github.com/mattn/go-sqlite3"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	loggerName       = "sqlite"
	sqliteDriverName = "wrapped-sqlite-driver"

	// CountQuery is a generic counter query used in a few query builders
	CountQuery = "COUNT(id)"

	// CurrentUnixTimeQuery is the query sqlite uses to determine the current unix time
	CurrentUnixTimeQuery = "(strftime('%s','now'))"
)

func init() {
	// Explicitly wrap the Sqlite driver with ocsql
	driver := ocsql.Wrap(
		&sqlite.SQLiteDriver{},
		ocsql.WithQuery(true),
		ocsql.WithAllowRoot(false),
		ocsql.WithRowsNext(true),
		ocsql.WithRowsClose(true),
		ocsql.WithQueryParams(true),
	)

	// Register our ocsql wrapper as a db driver
	sql.Register(sqliteDriverName, driver)
}

var _ database.Database = (*Sqlite)(nil)

type (
	// Sqlite is our main Sqlite interaction db
	Sqlite struct {
		logger      logging.Logger
		db          *sql.DB
		sqlBuilder  squirrel.StatementBuilderType
		migrateOnce sync.Once
		debug       bool
	}

	// ConnectionDetails is a string alias for a Sqlite url
	ConnectionDetails string

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)

// ProvideSqliteDB provides an instrumented sqlite db
func ProvideSqliteDB(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to sqlite")
	return sql.Open(sqliteDriverName, string(connectionDetails))
}

// ProvideSqlite provides a sqlite db controller
func ProvideSqlite(debug bool, db *sql.DB, logger logging.Logger) database.Database {
	return &Sqlite{
		db:         db,
		debug:      debug,
		logger:     logger.WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder,
	}
}

// IsReady reports whether or not the db is ready
func (s *Sqlite) IsReady(ctx context.Context) (ready bool) {
	return true
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (s *Sqlite) logQueryBuildingError(err error) {
	if err != nil {
		s.logger.WithName("QUERY_ERROR").Error(err, "building query")
	}
}

// logCreationTimeRetrievalError logs errors that may occur during creation time retrieval.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (s *Sqlite) logCreationTimeRetrievalError(err error) {
	if err != nil {
		s.logger.WithName("CREATION_TIME_RETRIEVAL").Error(err, "retrieving creation time")
	}
}

// buildError takes a given error and wraps it with a message, provided that it
// IS NOT sql.ErrNoRows, which we want to preserve and surface to the services.
func buildError(err error, msg string) error {
	if err == sql.ErrNoRows {
		return err
	}

	if !strings.Contains(msg, `%w`) {
		msg += ": %w"
	}

	return fmt.Errorf(msg, err)
}

package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	loggerName        = "mariadb"
	mariaDBDriverName = "wrapped-mariadb-driver"

	// CountQuery is a generic counter query used in a few query builders
	CountQuery = "COUNT(id)"

	// CurrentUnixTimeQuery is the query maria DB uses to determine the current unix time
	CurrentUnixTimeQuery = "UNIX_TIMESTAMP()"
)

func init() {
	// Explicitly wrap the MariaDB driver with ocsql
	driver := ocsql.Wrap(
		&mysql.MySQLDriver{},
		ocsql.WithQuery(true),
		ocsql.WithAllowRoot(false),
		ocsql.WithRowsNext(true),
		ocsql.WithRowsClose(true),
		ocsql.WithQueryParams(true),
	)

	// Register our ocsql wrapper as a db driver
	sql.Register(mariaDBDriverName, driver)
}

var _ database.Database = (*MariaDB)(nil)

type (
	// MariaDB is our main MariaDB interaction db
	MariaDB struct {
		logger      logging.Logger
		db          *sql.DB
		sqlBuilder  squirrel.StatementBuilderType
		migrateOnce sync.Once
		debug       bool
	}

	// ConnectionDetails is a string alias for a MariaDB url
	ConnectionDetails string

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)

// ProvideMariaDBConnection provides an instrumented maria DB db
func ProvideMariaDBConnection(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to maria DB")
	return sql.Open(mariaDBDriverName, string(connectionDetails))
}

// ProvideMariaDB provides a maria DB controller
func ProvideMariaDB(debug bool, db *sql.DB, logger logging.Logger) database.Database {
	return &MariaDB{
		db:         db,
		debug:      debug,
		logger:     logger.WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder,
	}
}

// IsReady reports whether or not the db is ready
func (m *MariaDB) IsReady(ctx context.Context) (ready bool) {
	numberOfUnsuccessfulAttempts := 0

	m.logger.WithValues(map[string]interface{}{
		"interval":     time.Second,
		"max_attempts": 50,
	}).Debug("IsReady called")

	for !ready {
		err := m.db.Ping()
		if err != nil {
			m.logger.Debug("ping failed, waiting for db")
			time.Sleep(time.Second)

			numberOfUnsuccessfulAttempts++
			if numberOfUnsuccessfulAttempts >= 50 {
				return false
			}
		} else {
			ready = true
			return ready
		}
	}
	return false
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (m *MariaDB) logQueryBuildingError(err error) {
	if err != nil {
		m.logger.WithName("QUERY_ERROR").Error(err, "building query")
	}
}

// logCreationTimeRetrievalError logs errors that may occur during creation time retrieval.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (m *MariaDB) logCreationTimeRetrievalError(err error) {
	if err != nil {
		m.logger.WithName("CREATION_TIME_RETRIEVAL").Error(err, "retrieving creation time")
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

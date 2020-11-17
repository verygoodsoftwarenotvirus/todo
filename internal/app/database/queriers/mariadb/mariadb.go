package mariadb

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/queriers"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	loggerName        = "mariadb"
	mariaDBDriverName = "wrapped-mariadb-driver"

	// countQuery is a generic counter query used in a few query builders.
	countQuery = `COUNT(%s.id)`
	// jsonPluckQuery is a generic format string for getting something out of the first layer of a JSON blob.
	jsonPluckQuery = `JSON_CONTAINS(%s.%s, '%d', '$.%s')`
	// currentUnixTimeQuery is the query maria DB uses to determine the current unix time.
	currentUnixTimeQuery = `UNIX_TIMESTAMP()`

	maximumConnectionAttempts        = 50
	defaultBucketSize         uint64 = 1000
)

func init() {
	// Explicitly wrap the MariaDB driver with ocsql.
	driver := ocsql.Wrap(
		&mysql.MySQLDriver{},
		ocsql.WithQuery(true),
		ocsql.WithAllowRoot(false),
		ocsql.WithRowsNext(true),
		ocsql.WithRowsClose(true),
		ocsql.WithQueryParams(true),
	)

	// Register our ocsql wrapper as a db driver.
	sql.Register(mariaDBDriverName, driver)
}

var _ database.DataManager = (*MariaDB)(nil)

type (
	// MariaDB is our main MariaDB interaction db.
	MariaDB struct {
		logger      logging.Logger
		db          *sql.DB
		timeTeller  queriers.TimeTeller
		sqlBuilder  squirrel.StatementBuilderType
		migrateOnce sync.Once
		debug       bool
	}

	// ConnectionDetails is a string alias for a MariaDB url.
	ConnectionDetails string

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects.
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)

// ProvideMariaDBConnection provides an instrumented maria DB db.
func ProvideMariaDBConnection(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to maria DB")
	return sql.Open(mariaDBDriverName, string(connectionDetails))
}

// ProvideMariaDB provides a maria DB controller.
func ProvideMariaDB(debug bool, db *sql.DB, logger logging.Logger) database.DataManager {
	return &MariaDB{
		db:         db,
		debug:      debug,
		timeTeller: &queriers.StandardTimeTeller{},
		logger:     logger.WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}
}

// IsReady reports whether or not the db is ready.
func (m *MariaDB) IsReady(ctx context.Context) (ready bool) {
	attemptCount := 0

	logger := m.logger.WithValues(map[string]interface{}{
		"interval":     time.Second,
		"max_attempts": maximumConnectionAttempts,
	})
	logger.Debug("IsReady called")

	for !ready {
		err := m.db.PingContext(ctx)
		if err != nil {
			logger.WithValue("attempt_count", attemptCount).Debug("ping failed, waiting for db")
			time.Sleep(time.Second)

			attemptCount++
			if attemptCount >= maximumConnectionAttempts {
				return false
			}
		} else {
			ready = true
			return ready
		}
	}
	return false
}

// BeginTx begins a transaction.
func (m *MariaDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (m *MariaDB) logQueryBuildingError(err error) {
	if err != nil {
		m.logger.WithValue("QUERY_ERROR", true).Error(err, "building query")
	}
}

// logIDRetrievalError logs errors that may occur during created db row ID retrieval.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (m *MariaDB) logIDRetrievalError(err error) {
	if err != nil {
		m.logger.WithValue("ROW_ID_ERROR", true).Error(err, "fetching row ID")
	}
}

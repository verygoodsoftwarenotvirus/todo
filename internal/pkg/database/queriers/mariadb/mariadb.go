package mariadb

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	loggerName = "mariadb"
	driverName = "wrapped-mariadb-driverName"

	// columnCountQueryTemplate is a generic counter query used in a few query builders.
	columnCountQueryTemplate = `COUNT(%s.id)`
	// allCountQuery is a generic counter query used in a few query builders.
	allCountQuery = `COUNT(*)`
	// jsonPluckQuery is a generic format string for getting something out of the first layer of a JSON blob.
	jsonPluckQuery = `JSON_CONTAINS(%s.%s, '%d', '$.%s')`
	// currentUnixTimeQuery is the query maria DB uses to determine the current unix time.
	currentUnixTimeQuery = `UNIX_TIMESTAMP()`

	maximumConnectionAttempts        = 50
	defaultBucketSize         uint64 = 1000
)

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

var instrumentedDriverRegistration sync.Once

// ProvideMariaDBConnection provides an instrumented maria DB db.
func ProvideMariaDBConnection(logger logging.Logger, connectionDetails database.ConnectionDetails, metricsCollectionInterval time.Duration) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to maria DB")

	instrumentedDriverRegistration.Do(func() {
		sql.Register(
			driverName,
			instrumentedsql.WrapDriver(
				&mysql.MySQLDriver{},
				instrumentedsql.WithOmitArgs(),
				instrumentedsql.WithTracer(tracing.NewInstrumentedSQLTracer("mariadb_connection")),
				instrumentedsql.WithLogger(tracing.NewInstrumentedSQLLogger(logger)),
			),
		)
	})

	db, err := sql.Open(driverName, string(connectionDetails))
	if err != nil {
		return nil, err
	}

	return db, nil
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
func (q *MariaDB) IsReady(ctx context.Context) (ready bool) {
	attemptCount := 0

	logger := q.logger.WithValues(map[string]interface{}{
		"interval":     time.Second,
		"max_attempts": maximumConnectionAttempts,
	})
	logger.Debug("IsReady called")

	for !ready {
		err := q.db.PingContext(ctx)
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
func (q *MariaDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return q.db.BeginTx(ctx, opts)
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (q *MariaDB) logQueryBuildingError(err error) {
	if err != nil {
		q.logger.WithValue("QUERY_ERROR", true).Error(err, "building query")
	}
}

// logIDRetrievalError logs errors that may occur during created db row ID retrieval.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (q *MariaDB) logIDRetrievalError(err error) {
	if err != nil {
		q.logger.WithValue("ROW_ID_ERROR", true).Error(err, "fetching row ID")
	}
}

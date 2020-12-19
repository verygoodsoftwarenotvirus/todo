package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	loggerName                 = "postgres"
	driverName                 = "wrapped-postgres-driver"
	postgresRowExistsErrorCode = "23505"

	// columnCountQueryTemplate is a generic counter query used in a few query builders.
	columnCountQueryTemplate = `COUNT(%s.id)`
	// allCountQuery is a generic counter query used in a few query builders.
	allCountQuery = `COUNT(*)`
	// jsonPluckQuery is a generic format string for getting something out of the first layer of a JSON blob.
	jsonPluckQuery = `%s.%s->'%s'`
	// currentUnixTimeQuery is the query postgres uses to determine the current unix time.
	currentUnixTimeQuery = `extract(epoch FROM NOW())`

	maximumConnectionAttempts        = 50
	defaultBucketSize         uint64 = 1000
)

func init() {
	// Register our ocsql wrapper as a db driver.
	sql.Register(
		driverName,
		ocsql.Wrap(
			&postgres.Driver{},
			ocsql.WithQuery(true),
			ocsql.WithAllowRoot(false),
			ocsql.WithRowsNext(true),
			ocsql.WithRowsClose(true),
			ocsql.WithQueryParams(true),
		),
	)
}

var _ database.DataManager = (*Postgres)(nil)

type (
	// Postgres is our main Postgres interaction db.
	Postgres struct {
		logger      logging.Logger
		db          *sql.DB
		sqlBuilder  squirrel.StatementBuilderType
		migrateOnce sync.Once
		debug       bool
	}

	// ConnectionDetails is a string alias for a Postgres url.
	ConnectionDetails string

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects.
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)

// ProvidePostgresDB provides an instrumented postgres db.
func ProvidePostgresDB(logger logging.Logger, connectionDetails database.ConnectionDetails, metricsCollectionInterval time.Duration) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to postgres")

	db, err := sql.Open(driverName, string(connectionDetails))
	if err != nil {
		return nil, err
	}

	ocsql.RegisterAllViews()
	ocsql.RecordStats(db, metricsCollectionInterval)

	return db, nil
}

// ProvidePostgres provides a postgres db controller.
func ProvidePostgres(debug bool, db *sql.DB, logger logging.Logger) database.DataManager {
	return &Postgres{
		db:         db,
		debug:      debug,
		logger:     logger.WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// IsReady reports whether or not the db is ready.
func (q *Postgres) IsReady(ctx context.Context) (ready bool) {
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
func (q *Postgres) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return q.db.BeginTx(ctx, opts)
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (q *Postgres) logQueryBuildingError(err error) {
	if err != nil {
		q.logger.WithValue("QUERY_ERROR", true).Error(err, "building query")
	}
}

func joinUint64s(in []uint64) string {
	out := []string{}

	for _, x := range in {
		out = append(out, strconv.FormatUint(x, 10))
	}

	return strings.Join(out, ",")
}

package sqlite

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/Masterminds/squirrel"
	"github.com/luna-duclos/instrumentedsql"
	sqlite "github.com/mattn/go-sqlite3"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	loggerName = "sqlite"
	driverName = "wrapped-sqlite-driverName"

	// columnCountQueryTemplate is a generic counter query used in a few query builders.
	columnCountQueryTemplate = `COUNT(%s.id)`
	// allCountQuery is a generic counter query used in a few query builders.
	allCountQuery = `COUNT(*)`
	// jsonPluckQuery is a generic format string for getting something out of the first layer of a JSON blob.
	jsonPluckQuery = `json_extract(%s.%s, '$.%s')`
	// currentUnixTimeQuery is the query sqlite uses to determine the current unix time.
	currentUnixTimeQuery = `(strftime('%s','now'))`
)

var (
	_ database.SQLQueryBuilder = (*Sqlite)(nil)
)

type (
	// Sqlite is our main Sqlite interaction db.
	Sqlite struct {
		logger     logging.Logger
		db         *sql.DB
		timeTeller queriers.TimeTeller
		sqlBuilder squirrel.StatementBuilderType
		debug      bool
	}
)

var instrumentedDriverRegistration sync.Once

// ProvideSqliteDB provides an instrumented sqlite db.
func ProvideSqliteDB(logger logging.Logger, connectionDetails database.ConnectionDetails, metricsCollectionInterval time.Duration) (*sql.DB, error) {
	logger.WithValue(keys.ConnectionDetailsKey, connectionDetails).Debug("Establishing connection to sqlite")

	instrumentedDriverRegistration.Do(func() {
		sql.Register(
			driverName,
			instrumentedsql.WrapDriver(
				&sqlite.SQLiteDriver{},
				instrumentedsql.WithOmitArgs(),
				instrumentedsql.WithTracer(tracing.NewInstrumentedSQLTracer("sqlite_connection")),
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

// ProvideSqlite provides a sqlite db controller.
func ProvideSqlite(debug bool, db *sql.DB, logger logging.Logger) *Sqlite {
	return &Sqlite{
		db:         db,
		debug:      debug,
		timeTeller: &queriers.StandardTimeTeller{},
		logger:     logger.WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}
}

// IsReady reports whether or not the db is ready.
func (c *Sqlite) IsReady(_ context.Context, maxAttempts uint8) (ready bool) {
	return true
}

// BeginTx begins a transaction.
func (c *Sqlite) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, opts)
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (c *Sqlite) logQueryBuildingError(err error) {
	if err != nil {
		c.logger.WithValue(keys.QueryErrorKey, true).Error(err, "building query")
	}
}

package sqlite

import (
	"database/sql"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/Masterminds/squirrel"
	"github.com/luna-duclos/instrumentedsql"
	sqlite "github.com/mattn/go-sqlite3"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
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
)

var (
	// currentUnixTimeQuery is the query sqlite uses to determine the current unix time.
	currentUnixTimeQuery = squirrel.Expr(`(strftime('%s','now'))`)
)

var _ querybuilding.SQLQueryBuilder = (*Sqlite)(nil)

type (
	// Sqlite is our main Sqlite interaction db.
	Sqlite struct {
		logger              logging.Logger
		tracer              tracing.Tracer
		sqlBuilder          squirrel.StatementBuilderType
		externalIDGenerator querybuilding.ExternalIDGenerator
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
func ProvideSqlite(logger logging.Logger) *Sqlite {
	return &Sqlite{
		logger:              logging.EnsureLogger(logger).WithName(loggerName),
		tracer:              tracing.NewTracer("sqlite_query_builder"),
		sqlBuilder:          squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
		externalIDGenerator: querybuilding.UUIDExternalIDGenerator{},
	}
}

// logQueryBuildingError logs errs that may occur during query construction. Such errors should be few and far between,
// as the generally only occur with type discrepancies or other misuses of SQL. An alert should be set up for any log
// entries with the given name, and those alerts should be investigated quickly.
func (b *Sqlite) logQueryBuildingError(span tracing.Span, err error) {
	if err != nil {
		logger := b.logger.WithValue(keys.QueryErrorKey, true)
		observability.AcknowledgeError(err, logger, span, "building query")
	}
}

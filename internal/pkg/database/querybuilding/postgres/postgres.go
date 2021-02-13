package postgres

import (
	"database/sql"
	"sync"

	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
	"github.com/luna-duclos/instrumentedsql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

const (
	loggerName = "postgres"
	driverName = "wrapped-postgres-driver"

	// columnCountQueryTemplate is a generic counter query used in a few query builders.
	columnCountQueryTemplate = `COUNT(%s.id)`
	// allCountQuery is a generic counter query used in a few query builders.
	allCountQuery = `COUNT(*)`
	// jsonPluckQuery is a generic format string for getting something out of the first layer of a JSON blob.
	jsonPluckQuery = `%s.%s->'%s'`
)

var (
	// currentUnixTimeQuery is the query postgres uses to determine the current unix time.
	currentUnixTimeQuery = squirrel.Expr(`extract(epoch FROM NOW())`)
)

var _ database.SQLQueryBuilder = (*Postgres)(nil)

type (
	// Postgres is our main Postgres interaction db.
	Postgres struct {
		logger              logging.Logger
		sqlBuilder          squirrel.StatementBuilderType
		externalIDGenerator querybuilding.ExternalIDGenerator
	}
)

var instrumentedDriverRegistration sync.Once

// ProvidePostgresDB provides an instrumented postgres db.
func ProvidePostgresDB(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue(keys.ConnectionDetailsKey, connectionDetails).Debug("Establishing connection to postgres")

	instrumentedDriverRegistration.Do(func() {
		sql.Register(
			driverName,
			instrumentedsql.WrapDriver(
				&postgres.Driver{},
				instrumentedsql.WithOmitArgs(),
				instrumentedsql.WithTracer(tracing.NewInstrumentedSQLTracer("postgres_connection")),
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

// ProvidePostgres provides a postgres db controller.
func ProvidePostgres(logger logging.Logger) *Postgres {
	pg := &Postgres{
		logger:              logging.EnsureLogger(logger).WithName(loggerName),
		sqlBuilder:          squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		externalIDGenerator: querybuilding.UUIDExternalIDGenerator{},
	}

	return pg
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (q *Postgres) logQueryBuildingError(err error) {
	if err != nil {
		q.logger.WithValue(keys.QueryErrorKey, true).Error(err, "building query")
	}
}

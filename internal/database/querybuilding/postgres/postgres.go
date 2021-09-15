package postgres

import (
	"database/sql"
	"sync"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
	"github.com/luna-duclos/instrumentedsql"
)

const (
	loggerName = "postgres"
	driverName = "wrapped-postgres-driver"

	// columnCountQueryTemplate is a generic counter query used in a few query builders.
	columnCountQueryTemplate = `COUNT(%s.id)`
)

var (
	// currentUnixTimeQuery is the query postgres uses to determine the current unix time.
	currentUnixTimeQuery = squirrel.Expr(`extract(epoch FROM NOW())`)
)

var _ querybuilding.SQLQueryBuilder = (*Postgres)(nil)

type (
	// Postgres is our main Postgres interaction db.
	Postgres struct {
		logger     logging.Logger
		tracer     tracing.Tracer
		sqlBuilder squirrel.StatementBuilderType
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
		logger:     logging.EnsureLogger(logger).WithName(loggerName),
		tracer:     tracing.NewTracer("postgres_query_builder"),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return pg
}

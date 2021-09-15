package mysql

import (
	"database/sql"
	"sync"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"
)

const (
	loggerName = "mysql"
	driverName = "wrapped-mysql-driverName"

	// columnCountQueryTemplate is a generic counter query used in a few query builders.
	columnCountQueryTemplate = `COUNT(%s.id)`
)

var (
	// currentUnixTimeQuery is the query maria DB uses to determine the current unix time.
	currentUnixTimeQuery = squirrel.Expr(`UNIX_TIMESTAMP()`)
)

var _ querybuilding.SQLQueryBuilder = (*MySQL)(nil)

type (
	// MySQL is our main MySQL interaction db.
	MySQL struct {
		logger     logging.Logger
		tracer     tracing.Tracer
		sqlBuilder squirrel.StatementBuilderType
	}
)

var instrumentedDriverRegistration sync.Once

// ProvideMySQLConnection provides an instrumented maria DB db.
func ProvideMySQLConnection(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue(keys.ConnectionDetailsKey, connectionDetails).Debug("Establishing connection to maria DB")

	instrumentedDriverRegistration.Do(func() {
		sql.Register(
			driverName,
			instrumentedsql.WrapDriver(
				&mysql.MySQLDriver{},
				instrumentedsql.WithOmitArgs(),
				instrumentedsql.WithTracer(tracing.NewInstrumentedSQLTracer("mysql_connection")),
				instrumentedsql.WithLogger(tracing.NewInstrumentedSQLLogger(logger)),
			),
		)
	})

	return sql.Open("mysql", string(connectionDetails))
}

// ProvideMySQL provides a maria DB controller.
func ProvideMySQL(logger logging.Logger) *MySQL {
	return &MySQL{
		logger:     logging.EnsureLogger(logger).WithName(loggerName),
		tracer:     tracing.NewTracer("mysql_query_builder"),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}
}

// logQueryBuildingError logs errs that may occur during query construction. Such errors should be few and far between,
// as the generally only occur with type discrepancies or other misuses of SQL. An alert should be set up for any log
// entries with the given name, and those alerts should be investigated quickly.
func (b *MySQL) logQueryBuildingError(span tracing.Span, err error) {
	if err != nil {
		logger := b.logger.WithValue(keys.QueryErrorKey, true)

		observability.AcknowledgeError(err, logger, span, "building query")
	}
}

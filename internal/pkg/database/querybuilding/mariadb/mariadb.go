package mariadb

import (
	"database/sql"
	"sync"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"
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
)

var (
	// currentUnixTimeQuery is the query maria DB uses to determine the current unix time.
	currentUnixTimeQuery = squirrel.Expr(`UNIX_TIMESTAMP()`)
)

var _ database.SQLQueryBuilder = (*MariaDB)(nil)

type (
	// MariaDB is our main MariaDB interaction db.
	MariaDB struct {
		logger              logging.Logger
		sqlBuilder          squirrel.StatementBuilderType
		externalIDGenerator querybuilding.ExternalIDGenerator
	}
)

var instrumentedDriverRegistration sync.Once

// ProvideMariaDBConnection provides an instrumented maria DB db.
func ProvideMariaDBConnection(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue(keys.ConnectionDetailsKey, connectionDetails).Debug("Establishing connection to maria DB")

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

	return sql.Open("mysql", string(connectionDetails))
}

// ProvideMariaDB provides a maria DB controller.
func ProvideMariaDB(logger logging.Logger) *MariaDB {
	return &MariaDB{
		logger:              logging.EnsureLogger(logger).WithName(loggerName),
		sqlBuilder:          squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
		externalIDGenerator: querybuilding.UUIDExternalIDGenerator{},
	}
}

// logQueryBuildingError logs errs that may occur during query construction.
// Such errs should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (q *MariaDB) logQueryBuildingError(err error) {
	if err != nil {
		q.logger.WithValue(keys.QueryErrorKey, true).Error(err, "building query")
	}
}

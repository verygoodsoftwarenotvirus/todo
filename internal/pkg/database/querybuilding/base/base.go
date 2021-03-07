package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/GuiaBolso/darwin"
	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	loggerName = "querybuilder"

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

var _ database.SQLQueryBuilder = (*BaseQueryBuilder)(nil)

type (
	// BaseQueryBuilder is our main BaseQueryBuilder interaction db.
	BaseQueryBuilder struct {
		logger              logging.Logger
		sqlBuilder          squirrel.StatementBuilderType
		externalIDGenerator querybuilding.ExternalIDGenerator
	}
)

// ProvideBaseQueryBuilder provides a sqlite db controller.
func ProvideBaseQueryBuilder(logger logging.Logger, pf squirrel.PlaceholderFormat) *BaseQueryBuilder {
	return &BaseQueryBuilder{
		logger:              logging.EnsureLogger(logger).WithName(loggerName),
		sqlBuilder:          squirrel.StatementBuilder.PlaceholderFormat(pf),
		externalIDGenerator: querybuilding.UUIDExternalIDGenerator{},
	}
}

// BuildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a sqlite database.
func (q *BaseQueryBuilder) BuildMigrationFunc(db *sql.DB) func() {
	return func() {
		d := darwin.NewGenericDriver(db, darwin.SqliteDialect{})
		if err := darwin.Migrate(d, []darwin.Migration{}, nil); err != nil {
			panic(fmt.Errorf("migrating database: %w", err))
		}
	}
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (q *BaseQueryBuilder) logQueryBuildingError(err error) {
	if err != nil {
		q.logger.WithValue(keys.QueryErrorKey, true).Error(err, "building query")
	}
}

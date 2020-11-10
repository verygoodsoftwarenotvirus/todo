package sqlite

import (
	"context"
	"database/sql"
	"sync"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	sqlite "github.com/mattn/go-sqlite3"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	loggerName       = "sqlite"
	sqliteDriverName = "wrapped-sqlite-driver"

	existencePrefix, existenceSuffix = "SELECT EXISTS (", ")"

	idColumn            = "id"
	createdOnColumn     = "created_on"
	lastUpdatedOnColumn = "last_updated_on"
	archivedOnColumn    = "archived_on"

	// countQuery is a generic counter query used in a few query builders.
	countQuery = "COUNT(%s.id)"

	// currentUnixTimeQuery is the query sqlite uses to determine the current unix time.
	currentUnixTimeQuery = "(strftime('%s','now'))"

	defaultBucketSize uint64 = 1000
)

func init() {
	// Explicitly wrap the Sqlite driver with ocsql.
	driver := ocsql.Wrap(
		&sqlite.SQLiteDriver{},
		ocsql.WithQuery(true),
		ocsql.WithAllowRoot(false),
		ocsql.WithRowsNext(true),
		ocsql.WithRowsClose(true),
		ocsql.WithQueryParams(true),
	)

	// Register our ocsql wrapper as a db driver.
	sql.Register(sqliteDriverName, driver)
}

var _ database.DataManager = (*Sqlite)(nil)

type (
	// Sqlite is our main Sqlite interaction db.
	Sqlite struct {
		logger      logging.Logger
		db          *sql.DB
		timeTeller  timeTeller
		sqlBuilder  squirrel.StatementBuilderType
		migrateOnce sync.Once
		debug       bool
	}

	// ConnectionDetails is a string alias for a Sqlite url.
	ConnectionDetails string

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects.
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)

// ProvideSqliteDB provides an instrumented sqlite db.
func ProvideSqliteDB(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to sqlite")
	return sql.Open(sqliteDriverName, string(connectionDetails))
}

// ProvideSqlite provides a sqlite db controller.
func ProvideSqlite(debug bool, db *sql.DB, logger logging.Logger) database.DataManager {
	return &Sqlite{
		db:         db,
		debug:      debug,
		timeTeller: &stdLibTimeTeller{},
		logger:     logger.WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}
}

// IsReady reports whether or not the db is ready.
func (s *Sqlite) IsReady(_ context.Context) (ready bool) {
	return true
}

// BeginTx begins a transaction.
func (s *Sqlite) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (s *Sqlite) logQueryBuildingError(err error) {
	if err != nil {
		s.logger.WithValue("QUERY_ERROR", true).Error(err, "building query")
	}
}

// logIDRetrievalError logs errors that may occur during created db row ID retrieval.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (s *Sqlite) logIDRetrievalError(err error) {
	if err != nil {
		s.logger.WithValue("ROW_ID_ERROR", true).Error(err, "fetching row ID")
	}
}

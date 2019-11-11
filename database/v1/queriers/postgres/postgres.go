package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	loggerName         = "postgres"
	postgresDriverName = "wrapped-postgres-driver"

	// CountQuery is a generic counter query used in a few query builders
	CountQuery = "COUNT(id)"

	// CurrentUnixTimeQuery is the query postgres uses to determine the current unix time
	CurrentUnixTimeQuery = "extract(epoch FROM NOW())"
)

func init() {
	// Explicitly wrap the Postgres driver with ocsql
	driver := ocsql.Wrap(
		&postgres.Driver{},
		ocsql.WithQuery(true),
		ocsql.WithAllowRoot(false),
		ocsql.WithRowsNext(true),
		ocsql.WithRowsClose(true),
		ocsql.WithQueryParams(true),
	)

	// Register our ocsql wrapper as a db driver
	sql.Register(postgresDriverName, driver)
}

var _ database.Database = (*Postgres)(nil)

type (
	// Postgres is our main Postgres interaction db
	Postgres struct {
		logger      logging.Logger
		db          *sql.DB
		sqlBuilder  squirrel.StatementBuilderType
		migrateOnce sync.Once
		debug       bool
	}

	// ConnectionDetails is a string alias for a Postgres url
	ConnectionDetails string

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)

// ProvidePostgresDB provides an instrumented postgres db
func ProvidePostgresDB(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to postgres")
	return sql.Open(postgresDriverName, string(connectionDetails))
}

// ProvidePostgres provides a postgres db controller
func ProvidePostgres(debug bool, db *sql.DB, logger logging.Logger) database.Database {
	return &Postgres{
		db:         db,
		debug:      debug,
		logger:     logger.WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// IsReady reports whether or not the db is ready
func (p *Postgres) IsReady(ctx context.Context) (ready bool) {
	numberOfUnsuccessfulAttempts := 0

	p.logger.WithValues(map[string]interface{}{
		"interval":     time.Second,
		"max_attempts": 50,
	}).Debug("IsReady called")

	for !ready {
		err := p.db.Ping()
		if err != nil {
			p.logger.Debug("ping failed, waiting for db")
			time.Sleep(time.Second)

			numberOfUnsuccessfulAttempts++
			if numberOfUnsuccessfulAttempts >= 50 {
				return false
			}
		} else {
			ready = true
			return ready
		}
	}
	return false
}

// logQueryBuildingError logs errors that may occur during query construction.
// Such errors should be few and far between, as the generally only occur with
// type discrepancies or other misuses of SQL. An alert should be set up for
// any log entries with the given name, and those alerts should be investigated
// with the utmost priority.
func (p *Postgres) logQueryBuildingError(err error) {
	if err != nil {
		p.logger.WithName("QUERY_ERROR").Error(err, "building query")
	}
}

// buildError takes a given error and wraps it with a message, provided that it
// IS NOT sql.ErrNoRows, which we want to preserve and surface to the services.
func buildError(err error, msg string) error {
	if err == sql.ErrNoRows {
		return err
	}

	if !strings.Contains(msg, `%w`) {
		msg += ": %w"
	}

	return fmt.Errorf(msg, err)
}

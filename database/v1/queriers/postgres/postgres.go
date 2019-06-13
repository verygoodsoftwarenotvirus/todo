package postgres

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
)

const (
	postgresDriverName = "wrapped-postgres-driver"

	// CountQuery is a generic counter query used in a few query builders
	CountQuery = "COUNT(id)"
)

func init() {
	// Explicitly wrap the SQLite3 driver with ocsql.
	driver := ocsql.Wrap(
		&postgres.Driver{},
		ocsql.WithQuery(true),
		ocsql.WithAllowRoot(false),
		ocsql.WithRowsNext(true),
		ocsql.WithRowsClose(true),
		ocsql.WithQueryParams(true),
	)

	// Register our ocsql wrapper as a db driver.
	sql.Register(postgresDriverName, driver)

}

type (
	// Postgres is our main Postgres interaction db
	Postgres struct {
		logger     logging.Logger
		db         *sql.DB
		sqlBuilder squirrel.StatementBuilderType
		migration  sync.Once
		debug      bool
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
func ProvidePostgres(
	debug bool,
	db *sql.DB,
	logger logging.Logger,
) database.Database {

	s := &Postgres{
		db:         db,
		debug:      debug,
		logger:     logger.WithName("postgres"),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return s
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

func logQueryBuildingError(logger logging.Logger, err error) {
	if err != nil {
		logger.WithName("QUERY_ERROR").Error(err, "building query")
	}
}

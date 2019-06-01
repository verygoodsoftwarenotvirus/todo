package postgres

import (
	"context"
	"database/sql"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
)

const (
	postgresDriverName = "wrapped-postgres-driver"
)

func init() {
	// Explicitly wrap the SQLite3 driver with ocsql.
	driver := ocsql.Wrap(&postgres.Driver{}, ocsql.WithQuery(true))

	// Register our ocsql wrapper as a database driver.
	sql.Register(postgresDriverName, driver)

}

type (
	// Postgres is our main Postgres interaction database
	Postgres struct {
		debug       bool
		logger      logging.Logger
		database    *sql.DB
		databaseURL string
		sqlBuilder  squirrel.StatementBuilderType
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

// ProvidePostgresDB provides an instrumented postgres database
func ProvidePostgresDB(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to postgres")

	return sql.Open(postgresDriverName, string(connectionDetails))
}

// ProvidePostgres provides a postgres database controller
func ProvidePostgres(
	debug bool,
	db *sql.DB,
	logger logging.Logger,
	connectionDetails database.ConnectionDetails,
) database.Database {

	s := &Postgres{
		database:    db,
		debug:       debug,
		logger:      logger.WithName("postgres"),
		databaseURL: string(connectionDetails),
		sqlBuilder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return s
}

// IsReady reports whether or not the database is ready
func (p *Postgres) IsReady(ctx context.Context) (ready bool) {
	numberOfUnsuccessfulAttempts := 0

	p.logger.WithValues(map[string]interface{}{
		"database_url": p.databaseURL,
		"interval":     time.Second,
		"max_attempts": 50,
	}).Debug("IsReady called")

	for !ready {
		err := p.database.Ping()
		if err != nil {
			p.logger.Debug("ping failed, waiting for database")
			time.Sleep(time.Second)
			numberOfUnsuccessfulAttempts++
			if numberOfUnsuccessfulAttempts >= 50 {
				return
			}
		} else {
			ready = true
			return
		}
	}
	return
}

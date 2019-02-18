package postgres

import (
	"context"
	"database/sql"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	"github.com/google/wire"
	_ "github.com/lib/pq" // we need the import here
	"github.com/pkg/errors"
)

type (
	// Postgres is our main Postgres interaction database
	Postgres struct {
		debug       bool
		logger      logging.Logger
		database    *sql.DB
		databaseURL string
	}

	// ConnectionDetails is a string alias for a Postgres url
	ConnectionDetails string

	// Scannable represents any database response (i.e. either a transaction or a regular execution response)
	Scannable interface {
		Scan(dest ...interface{}) error
	}

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)

var (
	// Providers is what we provide for dependency injection
	Providers = wire.NewSet(
		ProvidePostgres,
	)
)

// ProvidePostgres provides a postgres database controller
func ProvidePostgres(
	debug bool,
	logger logging.Logger,
	connectionDetails ConnectionDetails,
) (*Postgres, error) {

	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to postgres")
	db, err := sql.Open("postgres", string(connectionDetails))
	if err != nil {
		logger.Error(err, "error encountered establishing database connection")
		return nil, errors.Wrap(err, "establishing database connection")
	}

	s := &Postgres{
		debug:       debug,
		logger:      logger,
		database:    db,
		databaseURL: string(connectionDetails),
	}

	return s, nil
}

// IsReady reports whether or not the database is ready
func (p *Postgres) IsReady(ctx context.Context) (ready bool) {
	var (
		numberOfUnsuccessfulAttempts uint
		databaseIsNotReady           = true
	)

	p.logger.WithValues(map[string]interface{}{
		"database_url": p.databaseURL,
		"interval":     time.Second,
		"max_attempts": 50,
	}).Debug("IsReady called")

	for databaseIsNotReady {
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

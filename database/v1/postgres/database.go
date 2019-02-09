package postgres

import (
	"context"
	"database/sql"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/google/wire"
	_ "github.com/lib/pq" // we need the import here
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type (
	// Tracer is an arbitrary alias for dependency injection
	Tracer opentracing.Tracer

	// Postgres is our main Postgres interaction database
	Postgres struct {
		debug       bool
		logger      logging.Logger
		database    *sql.DB
		databaseURL string
		tracer      opentracing.Tracer
	}
)

var (
	_ database.Database = (*Postgres)(nil)

	// Providers is what we provide for dependency injection
	Providers = wire.NewSet(
		ProvidePostgres,
		ProvidePostgresTracer,
	)
)

// ProvidePostgresTracer provides a Tracer
func ProvidePostgresTracer() (Tracer, error) {
	return tracing.ProvideTracer("postgres-database")
}

// Begin utility funcs

type unixer interface {
	Unix() int64
}

func uip(u uint64) *uint64 {
	return &u
}

func timeToUInt64(t unixer) uint64 {
	_, ok := t.(*time.Time)
	if t == nil || !ok {
		return 0
	}
	return uint64(t.Unix())
}

func timeToPUInt64(t unixer) *uint64 {
	_, ok := t.(*time.Time)
	if t == nil || !ok {
		return nil
	}
	return uip(uint64(t.Unix()))
}

// End utility funcs

// ProvidePostgres provides a postgres database controller
func ProvidePostgres(
	debug bool,
	logger logging.Logger,
	tracer Tracer,
	connectionDetails database.ConnectionDetails,
) (database.Database, error) {

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
		tracer:      tracer,
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

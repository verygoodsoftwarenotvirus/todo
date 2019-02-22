package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/ExpansiveWorlds/instrumentedsql"
	"github.com/google/wire"
	postgres "github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
)

var (
	// Providers is what we provide for dependency injection
	Providers = wire.NewSet(
		ProvidePostgresDB,
		ProvidePostgres,
		ProvidePostgresTracer,
	)
)

type (
	// Tracer is a tracing wrapper
	Tracer opentracing.Tracer

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

// ProvidePostgresTracer provides a Postgres tracer
func ProvidePostgresTracer() Tracer {
	return tracing.ProvideTracer("postgres")
}

func buildLoggerFunc(logger logging.Logger) instrumentedsql.LoggerFunc {
	return func(ctx context.Context, msg string, keyvals ...interface{}) {
		var currentKey string

		for i, x := range keyvals {
			if i%2 == 0 {
				if y, ok := x.(string); ok && strings.TrimSpace(strings.ToLower(y)) == "query" {
					currentKey = y
				}
			} else if currentKey == "query" && x != nil {
				if z, ok := x.(string); ok && z != "" {
					logger.WithName("sql_debug").WithValue("query", x).Debug(msg)
					break
				}
			}
		}
	}
}

// ProvidePostgresDB provides an instrumented postgres database
func ProvidePostgresDB(
	logger logging.Logger,
	tracer Tracer,
	connectionDetails database.ConnectionDetails,
) (*sql.DB, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to postgres")

	loggerFunc := instrumentedsql.LoggerFunc(buildLoggerFunc(logger))

	sql.Register(
		"instrumented-postgres",
		instrumentedsql.WrapDriver(
			&postgres.Driver{},
			// instrumentedsql.WithTracer(instrumentedsql.Tracer),
			instrumentedsql.WithLogger(loggerFunc),
		),
	)
	return sql.Open("instrumented-postgres", string(connectionDetails))
}

// ProvidePostgres provides a postgres database controller
func ProvidePostgres(
	debug bool,
	db *sql.DB,
	logger logging.Logger,
	connectionDetails database.ConnectionDetails,
) (database.Database, error) {
	s := &Postgres{
		debug:       debug,
		logger:      logger.WithName("postgres"),
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

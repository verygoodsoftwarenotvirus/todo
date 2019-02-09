package sqlite

import (
	"context"
	"database/sql"
	// "errors"
	// "io/ioutil"
	// "path"
	// "strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	_ "github.com/mattn/go-sqlite3" // for the init import call
	"github.com/opentracing/opentracing-go"
)

type (
	// Tracer is an arbitrary alias for dependency injection
	Tracer opentracing.Tracer

	// Sqlite is our main Sqlite interaction database
	Sqlite struct {
		debug    bool
		database *sql.DB
		logger   logging.Logger
		tracer   opentracing.Tracer
	}
)

var (
	_ database.Database = (*Sqlite)(nil)

	// Providers is what we provide for dependency injection
	Providers = wire.NewSet(
		ProvideSqlite,
		ProvideSqliteTracer,
	)
)

// ProvideSqliteTracer provides a Tracer
func ProvideSqliteTracer() (Tracer, error) {
	return tracing.ProvideTracer("sqlite-database")
}

// ProvideSqlite provides a sqlite database controller
func ProvideSqlite(
	debug bool,
	logger logging.Logger,
	tracer Tracer,
	connectionDetails database.ConnectionDetails,
) (database.Database, error) {
	logger.WithValue("connection_details", connectionDetails).Debug("Establishing connection to sqlite3 file")
	db, err := sql.Open("sqlite3", string(connectionDetails))
	if err != nil {
		logger.Error(err, "error encountered establishing database connection")
		return nil, err
	}

	s := &Sqlite{
		debug:    debug,
		logger:   logger,
		database: db,
		tracer:   tracer,
	}

	return s, nil
}

func (s *Sqlite) prepareFilter(filter *models.QueryFilter, span opentracing.Span) *models.QueryFilter {
	if filter == nil {
		s.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.SetPage(filter.Page)

	span.SetTag("limit", filter.Limit)
	span.SetTag("page", filter.Page)
	span.SetTag("queryPage", filter.QueryPage)

	return filter
}

// IsReady reports whether or not Sqlite is ready to be written to.
// Since sqlite3 is a file-based database, it is always ready AFAICT
func (s *Sqlite) IsReady(ctx context.Context) (ready bool) {
	return true
}

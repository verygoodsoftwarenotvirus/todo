package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"io/ioutil"
	"path"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/google/wire"
	_ "github.com/mattn/go-sqlite3" // for the init import call
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

type (
	// Tracer is an arbitrary alias for dependency injection
	Tracer opentracing.Tracer

	// Sqlite is our main Sqlite interaction database
	Sqlite struct {
		debug     bool
		logger    *logrus.Logger
		newLogger logging.Logger
		database  *sql.DB
		tracer    opentracing.Tracer
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
	logger *logrus.Logger,
	newLogger logging.Logger,
	tracer Tracer,
	connectionDetails database.ConnectionDetails,
) (database.Database, error) {
	logger.Debugf("Establishing connection to sqlite3 file: %q\n", connectionDetails)
	db, err := sql.Open("sqlite3", string(connectionDetails))
	if err != nil {
		logger.Errorf("error encountered establishing database connection: %v\n", err)
		return nil, err
	}

	s := &Sqlite{
		debug:     debug,
		logger:    logger,
		newLogger: newLogger,
		database:  db,
		tracer:    tracer,
	}

	return s, nil
}

// IsReady reports whether or not Sqlite is ready to be written to. Since Sqlite is a file-based database, it is always ready
func (s *Sqlite) IsReady(ctx context.Context) (ready bool) {
	return true
}

// Migrate migrates a given Sqlite database. The current implementation is pretty primitive.
func (s *Sqlite) Migrate(ctx context.Context, schemaDir database.SchemaDirectory) error {
	sd := string(schemaDir)
	logger := s.logger.WithField("schema_dir", sd)
	logger.Debugln("Migrate called")

	if ready := s.IsReady(ctx); !ready {
		return errors.New("database not ready")
	}

	files, err := ioutil.ReadDir(string(sd))
	if err != nil {
		return err
	}
	logger.Debugf("%d files found in schema directory", len(files))

	for _, file := range files {
		schemaFile := path.Join(string(sd), file.Name())

		if strings.HasSuffix(schemaFile, ".sql") {
			logger.Debugf("migrating schema file: %q", schemaFile)
			data, err := ioutil.ReadFile(schemaFile)
			if err != nil {
				s.logger.Errorf("error encountered reading schema file: %q (%v)\n", schemaFile, err)
				return err
			}

			logger.Debugf("running query: %q", string(data))
			_, err = s.database.Exec(string(data))
			if err != nil {
				logger.Debugln("database.Exec finished, returning err")
				return err
			}
			logger.Debugln("database.Exec finished, error not returned")
		}
	}

	s.logger.Debugln("returning no error from sqlite.Migrate()")
	return nil
}

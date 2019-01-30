package sqlite

import (
	"database/sql"
	"io/ioutil"
	"path"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
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
		debug    bool
		logger   *logrus.Logger
		database *sql.DB
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
	logger *logrus.Logger,
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
		debug:    debug,
		logger:   logger,
		database: db,
		tracer:   tracer,
	}

	return s, nil
}

// Migrate migrates a given Sqlite database. The current implementation is pretty primitive.
func (s *Sqlite) Migrate(schemaDir database.SchemaDirectory) error {
	s.logger.Debugln("Migrate called")

	files, err := ioutil.ReadDir(string(schemaDir))
	if err != nil {
		return err
	}
	s.logger.Debugf("%d files found in schema directory", len(files))

	for _, file := range files {
		schemaFile := path.Join(string(schemaDir), file.Name())

		if strings.HasSuffix(schemaFile, ".sql") {
			s.logger.Debugf("migrating schema file: %q", schemaFile)
			data, err := ioutil.ReadFile(schemaFile)
			if err != nil {
				s.logger.Errorf("error encountered reading schema file: %q (%v)\n", schemaFile, err)
				return err
			}

			s.logger.Debugf("running query: %q", string(data))
			_, err = s.database.Exec(string(data))
			if err != nil {
				s.logger.Debugln("database.Exec finished, returning err")
				return err
			}
			s.logger.Debugln("database.Exec finished, error not returned")
		}
	}

	s.logger.Debugln("returning no error from sqlite.Migrate()")
	return nil

}

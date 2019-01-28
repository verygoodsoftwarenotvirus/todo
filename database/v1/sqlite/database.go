package sqlite

import (
	"database/sql"
	"io/ioutil"
	"path"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	_ "github.com/mattn/go-sqlite3" // for the init import call
	"github.com/sirupsen/logrus"
)

var _ database.Database = (*Sqlite)(nil)

// Sqlite is our main Sqlite interaction database
type Sqlite struct {
	debug    bool
	logger   *logrus.Logger
	database *sql.DB
}

// ProvideSqlite provides a sqlite database controller
func ProvideSqlite(
	debug bool,
	logger *logrus.Logger,
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

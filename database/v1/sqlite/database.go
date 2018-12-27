package sqlite

import (
	"database/sql"
	"io/ioutil"
	"path"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var _ database.Database = (*sqlite)(nil)

type sqlite struct {
	debug    bool
	logger   *logrus.Logger
	database *sql.DB

	clientIDExtractor database.ClientIDExtractor
	secretGenerator   database.SecretGenerator
}

// NewSqlite provides a sqlite database controller
func NewSqlite(config database.Config) (database.Database, error) {
	if config.Logger == nil {
		config.Logger = logrus.New()
	}

	config.Logger.Debugf("Establishing connection to sqlite3 file: %q\n", config.ConnectionString)
	db, err := sql.Open("sqlite3", config.ConnectionString)
	if err != nil {
		config.Logger.Errorf("error encountered establishing database connection: %v\n", err)
		return nil, err
	}

	s := &sqlite{
		debug:    config.Debug,
		logger:   config.Logger,
		database: db,

		// FIXME: these should be allowed to be nil
		clientIDExtractor: config.Extractor,
		secretGenerator:   config.SecretGenerator,
	}

	return s, nil
}

func (s *sqlite) Migrate(schemaDir string) error {
	s.logger.Debugln("Migrate called")

	files, err := ioutil.ReadDir(schemaDir)
	if err != nil {
		return err
	}
	s.logger.Debugf("%d files found in schema directory", len(files))

	for _, file := range files {
		schemaFile := path.Join(schemaDir, file.Name())

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

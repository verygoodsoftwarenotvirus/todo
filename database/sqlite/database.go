package sqlite

import (
	"database/sql"
	"io/ioutil"
	"path"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var _ database.Database = (*sqlite)(nil)

type sqlite struct {
	debug    bool
	logger   *logrus.Logger
	database *sql.DB
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
	}

	return s, nil
}

func (s *sqlite) Migrate(schemaDir string) error {
	s.logger.Debugln("Migrate called")

	if s.debug {
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
	} else {
		s.logger.Errorln("debug not enabled")
		return errors.New("I haven't implemented embedded data yet")
	}

	s.logger.Debugln("returning no error from sqlite.Migrate()")
	return nil

}

func (s *sqlite) GetItem(id uint) (*models.Item, error) {
	query := `
	SELECT
		id, name, details, created_on, completed_on
	FROM
		items
	WHERE
		id = ?
	`

	i := &models.Item{}

	err := s.database.QueryRow(query, id).Scan(
		&i.ID,
		&i.Name,
		&i.Details,
		&i.CreatedOn,
		&i.CompletedOn,
	)

	if err != nil {
		return nil, err
	}

	return i, err
}

func (s *sqlite) GetItems(filter *models.QueryFilter) ([]models.Item, error) {
	if filter == nil {
		s.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}

	list := []models.Item{}

	query := `
	SELECT
		id, name, details, created_on, completed_on
	FROM
		items
	LIMIT ?
	OFFSET ?
	`

	s.logger.Infof("query limit: %d, query page: %d, calculated page: %d", filter.Limit, filter.Page, uint(filter.Limit*(filter.Page-1)))

	rows, err := s.database.Query(query, filter.Limit, uint(filter.Limit*(filter.Page-1)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Details,
			&item.CreatedOn,
			&item.CompletedOn,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, err
}
func (s *sqlite) CreateItem(input *models.ItemInput) (*models.Item, error) {
	query := `
	INSERT INTO items
	(
		name, details
	)
	VALUES
	(
		?, ?
	)
	` /*
		RETURNING id, created_on
		err := s.database.QueryRow(query, input.Name, input.Details).Scan(
			&i.ID,
			&i.CreatedOn,
		) */
	i := &models.Item{
		Name:    input.Name,
		Details: input.Details,
	}

	_, err := s.database.Exec(query, input.Name, input.Details)
	return i, err
}
func (s *sqlite) UpdateItem(input *models.Item) error {
	query := `
	UPDATE items SET
		name = ?,
		details = ?
	WHERE id = ?
	`

	_, err := s.database.Exec(query, input.Name, input.Details, input.ID)

	return err
}
func (s *sqlite) DeleteItem(id uint) error {
	query := `
	DELETE FROM
		items
	WHERE
		id = ?
	`

	_, err := s.database.Exec(query, id)
	return err
}

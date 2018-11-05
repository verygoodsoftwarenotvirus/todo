package sqlite

import (
	"database/sql"

	db "gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var _ db.Database = (*database)(nil)

type database struct {
	debug    bool
	logger   *logrus.Logger
	database *sql.DB
}

type Config struct {
	Debug            bool
	ConnectionString string
}

func NewSqlite3Database(config *Config, logger *logrus.Logger) (*database, error) {
	if logger == nil {
		logger = logrus.New()
	}

	logger.Debugf("Establishing connection to sqlite3 file: %q\n", config.ConnectionString)
	db, err := sql.Open("sqlite3", config.ConnectionString)
	if err != nil {
		logger.Errorf("error encountered establishing database connection: %v\n", err)
		return nil, err
	}

	return &database{
		debug:    config.Debug,
		logger:   logger,
		database: db,
	}, nil
}

func (db *database) GetItem(id uint) (*models.Item, error) {
	query := `
	SELECT
		id, name, details, created_on, completed_on
	FROM
		items
	WHERE
		id = ?
	`

	i := &models.Item{}

	err := db.database.QueryRow(query, id).Scan(
		&i.ID,
		&i.Name,
		&i.Details,
		&i.CreatedOn,
		&i.CompletedOn,
	)

	return i, err
}
func (db *database) GetItems(filter *models.QueryFilter) ([]models.Item, error) {
	var list []models.Item

	query := `
	SELECT
		id, name, details, created_on, completed_on
	FROM
		items
	LIMIT ?
	OFFSET ?
	`

	rows, err := db.database.Query(query, filter.Limit, filter.Limit*filter.Page)
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
func (db *database) CreateItem(input *models.ItemInput) (*models.Item, error) {
	query := `
	INSERT INTO items
		(
			name, details
		)
	VALUES
		(
			?, ?
		)
	RETURNING id, created_on
	`

	i := &models.Item{
		Name:    input.Name,
		Details: input.Details,
	}

	err := db.database.QueryRow(query, input.Name, input.Details).Scan(
		&i.ID,
		&i.CreatedOn,
	)

	return i, err
}
func (db *database) UpdateItem(input *models.Item) error {
	query := `
	UPDATE items SET
		name = ?,
		details = ?
	WHERE id = ?
	`

	_, err := db.database.Exec(query, input.Name, input.Details, input.ID)

	return err
}
func (db *database) DeleteItem(id uint) error {
	query := `
	DELETE FROM
		items
	WHERE
		id = ?
	`

	_, err := db.database.Exec(query, id)
	return err
}

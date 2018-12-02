package sqlite

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	getItemQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on
		FROM
			items
		WHERE
			id = ? AND completed_on is null
	`
	getItemsQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on
		FROM
			items
		WHERE
			completed_on is null
		LIMIT ?
		OFFSET ?
	`
	createItemQuery = `
		INSERT INTO items
		(
			name, details
		)
		VALUES
		(
			?, ?
		)
	`
	updateItemQuery = `
		UPDATE items SET
			name = ?,
			details = ?,
			updated_on = (strftime('%s','now'))
		WHERE id = ?
	`
	archiveItemQuery = `
		UPDATE items SET
			updated_on = (strftime('%s','now')),
			completed_on = (strftime('%s','now'))
		WHERE id = ?
	`
)

func scanItem(scan database.Scannable) (*models.Item, error) {
	i := &models.Item{}
	err := scan.Scan(
		&i.ID,
		&i.Name,
		&i.Details,
		&i.CreatedOn,
		&i.UpdatedOn,
		&i.CompletedOn,
	)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (s *sqlite) GetItem(id uint) (*models.Item, error) {
	row := s.database.QueryRow(getItemQuery, id)
	i, err := scanItem(row)
	return i, err
}

func (s *sqlite) GetItems(filter *models.QueryFilter) ([]models.Item, error) {
	if filter == nil {
		s.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}

	list := []models.Item{}

	s.logger.Infof("query limit: %d, query page: %d, calculated page: %d", filter.Limit, filter.Page, uint(filter.Limit*(filter.Page-1)))

	rows, err := s.database.Query(getItemsQuery, filter.Limit, uint(filter.Limit*(filter.Page-1)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, err
}

func (s *sqlite) CreateItem(input *models.ItemInput) (i *models.Item, err error) {
	i = &models.Item{
		Name:    input.Name,
		Details: input.Details,
	}

	tx, err := s.database.Begin()
	if err != nil {
		s.logger.Errorf("error beginning database connection: %v", err)
		return nil, err
	}

	// create the item
	res, err := tx.Exec(createItemQuery, input.Name, input.Details)
	if err != nil {
		s.logger.Errorf("error executing item creation query: %v", err)
		tx.Rollback()
		return nil, err
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		s.logger.Errorf("error fetching last inserted item ID: %v", err)
		return nil, err
	}

	// fetch full updated item
	row := tx.QueryRow(getItemQuery, id)
	i, err = scanItem(row)
	if err != nil {
		s.logger.Errorf("error fetching newly created item %d: %v", id, err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorf("error committing transaction: %v", err)
		return nil, err
	}

	s.logger.Debugln("returning from CreateItem")
	return i, nil
}

func (s *sqlite) UpdateItem(input *models.Item) (err error) {
	tx, err := s.database.Begin()
	if err != nil {
		return
	}

	// update the item
	if _, err = tx.Exec(updateItemQuery, input.Name, input.Details, input.ID); err != nil {
		tx.Rollback()
		return
	}

	// fetch full updated item
	if err = tx.QueryRow(getItemQuery, input.ID).Scan(
		&input.ID,
		&input.Name,
		&input.Details,
		&input.CreatedOn,
		&input.UpdatedOn,
		&input.CompletedOn,
	); err != nil {
		tx.Rollback()
		return
	}

	// commit the changes
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return
	}

	return
}

func (s *sqlite) DeleteItem(id uint) error {
	_, err := s.database.Exec(archiveItemQuery, id)
	return err
}

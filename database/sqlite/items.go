package sqlite

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

const (
	getItemQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on
		FROM
			items
		WHERE
			id = ? AND completed_on = null
	`
	getItemsQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on
		FROM
			items
		WHERE
			completed_on = null
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
	createdOnItemQuery = `
		SELECT
			created_on
		FROM
			items
		WHERE id = ?
	`
	updatedOnItemQuery = `
		SELECT
			created_on
		FROM
			items
		WHERE id = ?
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

func (s *sqlite) GetItem(id uint) (*models.Item, error) {
	i := &models.Item{}

	if err := s.database.QueryRow(getItemQuery, id).Scan(
		&i.ID,
		&i.Name,
		&i.Details,
		&i.CreatedOn,
		&i.UpdatedOn,
		&i.CompletedOn,
	); err != nil {
		return nil, err
	}

	return i, nil
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
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Details,
			&item.CreatedOn,
			&item.UpdatedOn,
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
	i := &models.Item{
		Name:    input.Name,
		Details: input.Details,
	}

	tx, err := s.database.Begin()
	if err != nil {
		return nil, err
	}

	res, err := tx.Exec(createItemQuery, input.Name, input.Details)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, e
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, e
		}
		return nil, err
	}
	i.ID = uint(id)

	if err := tx.QueryRow(createdOnItemQuery, id).Scan(&i.CreatedOn); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return i, nil
}

func (s *sqlite) UpdateItem(input *models.Item) (err error) {
	tx, err := s.database.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec(updateItemQuery, input.Name, input.Details, input.ID)
	if err != nil {
		tx.Rollback()
		return
	}

	if err = tx.QueryRow(updatedOnItemQuery, input.ID).Scan(&input.UpdatedOn); err != nil {
		tx.Rollback()
		return
	}

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

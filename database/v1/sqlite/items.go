package sqlite

import (
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	getItemQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on, belongs_to
		FROM
			items
		WHERE
			id = ? AND belongs_to = ?
	`
	getItemCountQuery = `
		SELECT
			COUNT(*)
		FROM
			items
		WHERE completed_on IS NULL
	`
	getItemsQuery = `
		SELECT
			id, name, details, created_on, updated_on, completed_on, belongs_to
		FROM
			items
		WHERE
			completed_on IS NULL
		LIMIT ?
		OFFSET ?
	`
	createItemQuery = `
		INSERT INTO items
		(
			name, details, belongs_to
		)
		VALUES
		(
			?, ?, ?
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
		WHERE id = ? AND completed_on IS NULL
	`
)

func scanItem(scan database.Scannable) (*models.Item, error) {
	x := &models.Item{}
	err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.CompletedOn,
		&x.BelongsTo,
	)
	if err != nil {
		return nil, err
	}
	return x, nil
}

func (s *sqlite) GetItem(itemID, userID uint64) (*models.Item, error) {
	row := s.database.QueryRow(getItemQuery, itemID, userID)
	return scanItem(row)
}

func (s *sqlite) GetItemCount(filter *models.QueryFilter) (count uint64, err error) {
	return count, s.database.QueryRow(getItemCountQuery).Scan(&count)
}

func (s *sqlite) GetItems(filter *models.QueryFilter) (*models.ItemList, error) {
	if filter == nil {
		s.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.Page = uint64(math.Max(1, float64(filter.Page)))
	queryPage := uint(filter.Limit * (filter.Page - 1))

	list := []models.Item{}

	s.logger.Debugf("query limit: %d, query page: %d, calculated page: %d", filter.Limit, filter.Page, queryPage)

	rows, err := s.database.Query(getItemsQuery, filter.Limit, queryPage)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	count, err := s.GetItemCount(filter)
	if err != nil {
		return nil, err
	}

	x := &models.ItemList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			TotalCount: count,
			Limit:      filter.Limit,
		},
		Items: list,
	}

	return x, err
}

func (s *sqlite) CreateItem(input *models.ItemInput) (i *models.Item, err error) {
	tx, err := s.database.Begin()
	if err != nil {
		s.logger.Errorf("error beginning database connection: %v", err)
		return nil, err
	}

	// create the item
	res, err := tx.Exec(createItemQuery, input.Name, input.Details, input.BelongsTo)
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
	row := tx.QueryRow(getItemQuery, id, input.BelongsTo)
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
	row := tx.QueryRow(getItemQuery, input.ID, input.BelongsTo)
	input, err = scanItem(row)
	if err != nil {
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

func (s *sqlite) DeleteItem(id uint64) error {
	_, err := s.database.Exec(archiveItemQuery, id)
	return err
}

package sqlite

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	getUserQuery = `
		SELECT
			id,
			username,
			password,
			password_last_changed_on,
			two_factor_secret,
			created_on,
			updated_on,
			archived_on
		FROM
			users
		WHERE
			username = ?
	`
	getUserCountQuery = `
		SELECT
			COUNT(*)
		FROM
			users
		WHERE archived_on is null
	`
	getUserQueryByID = `
		SELECT
			id,
			username,
			password,
			password_last_changed_on,
			two_factor_secret,
			created_on,
			updated_on,
			archived_on
		FROM
			users
		WHERE
			id = ? AND archived_on is null
	`
	getUsersQuery = `
		SELECT
			id,
			username,
			password,
			password_last_changed_on,
			two_factor_secret,
			created_on,
			updated_on,
			archived_on
		FROM
			users
		WHERE
		archived_on is null
		LIMIT ?
		OFFSET ?
	`
	createUserQuery = `
		INSERT INTO users
		(
			username, password, two_factor_secret
		)
		VALUES
		(
			?, ?, ?
		)
	`
	updateUserQuery = `
		UPDATE users SET
			username = ?,
			password = ?,
			updated_on = (strftime('%s','now'))
		WHERE id = ?
	`
	archiveUserQuery = `
		UPDATE users SET
			updated_on = (strftime('%s','now')),
			archived_on = (strftime('%s','now'))
		WHERE id = ?
	`
)

func scanUser(scan database.Scannable) (*models.User, error) {
	x := &models.User{}
	err := scan.Scan(
		&x.ID,
		&x.Username,
		&x.HashedPassword,
		&x.PasswordLastChangedOn,
		&x.TwoFactorSecret,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
	)
	if err != nil {
		return nil, err
	}
	return x, nil
}

func (s *sqlite) GetUser(identifier string) (*models.User, error) {
	row := s.database.QueryRow(getUserQuery, identifier)
	return scanUser(row)
}

func (s *sqlite) GetUserCount(filter *models.QueryFilter) (uint64, error) {
	var count uint64
	err := s.database.QueryRow(getUserCountQuery).Scan(&count)
	return count, err
}

func (s *sqlite) GetUsers(filter *models.QueryFilter) (*models.UserList, error) {
	if filter == nil {
		s.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}
	list := []models.User{}

	s.logger.Infof("query limit: %d, query page: %d, calculated page: %d", filter.Limit, filter.Page, uint(filter.Limit*(filter.Page-1)))
	rows, err := s.database.Query(getUsersQuery, filter.Limit, uint(filter.Limit*(filter.Page-1)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	count, err := s.GetUserCount(filter)
	if err != nil {
		return nil, err
	}

	x := &models.UserList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Users: list,
	}

	return x, err
}

func (s *sqlite) CreateUser(input *models.UserInput, totpSecret string) (x *models.User, err error) {
	s.logger.Debugf("CreateUser called for %s", input.Username)

	tx, err := s.database.Begin()
	if err != nil {
		s.logger.Errorf("error beginning database connection: %v", err)
		return nil, err
	}

	// create the user
	res, err := tx.Exec(createUserQuery, input.Username, input.Password, totpSecret)
	if err != nil {
		s.logger.Errorf("error executing user creation query: %v", err)
		tx.Rollback()
		return nil, err
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		s.logger.Errorf("error fetching last inserted user ID: %v", err)
		return nil, err
	}

	// fetch full updated user
	finalRow := tx.QueryRow(getUserQueryByID, id)
	x, err = scanUser(finalRow)
	if err != nil {
		s.logger.Errorf("error fetching newly created user %d: %v", id, err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorf("error committing transaction: %v", err)
		return nil, err
	}

	s.logger.Debugln("returning from CreateUser")
	return
}

func (s *sqlite) UpdateUser(input *models.User) (err error) {
	tx, err := s.database.Begin()
	if err != nil {
		return
	}

	// update the user
	if _, err = tx.Exec(updateUserQuery, input.Username, input.HashedPassword, input.ID); err != nil {
		tx.Rollback()
		return
	}

	// fetch full updated user
	finalRow := tx.QueryRow(getUserQuery, input.ID)
	input, err = scanUser(finalRow)
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

func (s *sqlite) DeleteUser(id uint) error {
	_, err := s.database.Exec(archiveUserQuery, id)
	return err
}

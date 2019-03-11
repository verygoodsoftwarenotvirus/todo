package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

func scanUser(scan Scannable) (*models.User, error) {
	x := &models.User{}
	err := scan.Scan(
		&x.ID,
		&x.Username,
		&x.HashedPassword,
		&x.PasswordLastChangedOn,
		&x.TwoFactorSecret,
		&x.IsAdmin,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
	)
	if err != nil {
		return nil, err
	}
	return x, nil
}

func (s *Sqlite) scanUsers(rows *sql.Rows) ([]models.User, error) {
	var list []models.User

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scanning user result")
		}
		list = append(list, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

const adminUserExistsQuery = `
	SELECT EXISTS(SELECT id FROM users WHERE is_admin = true)
`

// AdminUserExists validates whether or not an admin user exists
func (s *Sqlite) AdminUserExists(ctx context.Context) (bool, error) {
	var exists string

	err := s.database.QueryRowContext(ctx, adminUserExistsQuery).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return exists == "true", err
}

const getUserQuery = `
	SELECT
		id,
		username,
		hashed_password,
		password_last_changed_on,
		two_factor_secret,
		is_admin,
		created_on,
		updated_on,
		archived_on
	FROM
		users
	WHERE
		id = ?
`

// GetUser fetches a user by their username
func (s *Sqlite) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	row := s.database.QueryRowContext(ctx, getUserQuery, userID)
	u, err := scanUser(row)
	return u, err
}

const getUserByUsernameQuery = `
	SELECT
		id,
		username,
		hashed_password,
		password_last_changed_on,
		two_factor_secret,
		is_admin,
		created_on,
		updated_on,
		archived_on
	FROM
		users
	WHERE
		username = ?
`

// GetUserByUsername fetches a user by their username
func (s *Sqlite) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	row := s.database.QueryRowContext(ctx, getUserByUsernameQuery, username)
	u, err := scanUser(row)
	return u, err
}

const getUserCountQuery = `
	SELECT
		COUNT(*)
	FROM
		users
	WHERE
		archived_on IS NULL
`

// GetUserCount fetches a count of users from the sqlite database that meet a particular filter
func (s *Sqlite) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	err = s.database.QueryRowContext(ctx, getUserCountQuery).Scan(&count)
	return
}

const getUsersQuery = `
	SELECT
		id,
		username,
		hashed_password,
		password_last_changed_on,
		two_factor_secret,
		is_admin,
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

// GetUsers fetches a list of users from the sqlite database that meet a particular filter
func (s *Sqlite) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	rows, err := s.database.QueryContext(ctx, getUsersQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, err
	}

	list, err := s.scanUsers(rows)
	if err != nil {
		return nil, err
	}

	count, err := s.GetUserCount(ctx, filter)
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

	return x, nil
}

const getUserQueryByID = `
	SELECT
		id,
		username,
		hashed_password,
		password_last_changed_on,
		two_factor_secret,
		is_admin,
		created_on,
		updated_on,
		archived_on
	FROM
		users
	WHERE
		id = ? AND archived_on is null
`

const createUserQuery = `
	INSERT INTO users
	(
		username,
		hashed_password,
		two_factor_secret,
		is_admin
	)
	VALUES
	(
		?, ?, ?, ?
	)
`

// CreateUser creates a user
func (s *Sqlite) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	res, err := s.database.ExecContext(ctx, createUserQuery, input.Username, input.Password, input.TwoFactorSecret, input.IsAdmin)
	if err != nil {
		return nil, errors.Wrap(err, "error executing user creation query")
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching last inserted user ID")
	}

	// fetch full updated user
	finalRow := s.database.QueryRowContext(ctx, getUserQueryByID, id)
	x, err := scanUser(finalRow)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching newly created user %d: %v")
	}

	return x, nil
}

const updateUserQuery = `
	UPDATE users
	SET
		username = ?,
		password = ?,
		updated_on = (strftime('%s','now'))
	WHERE
		id = ?
`

// UpdateUser updates a user. Note that this function expects the provided user to have a valid ID.
func (s *Sqlite) UpdateUser(ctx context.Context, input *models.User) error {
	if _, err := s.database.ExecContext(ctx, updateUserQuery, input.Username, input.HashedPassword, input.ID); err != nil {
		return err
	}

	return nil
}

const archiveUserQuery = `
	UPDATE users
	SET
		updated_on = (strftime('%s','now')),
		archived_on = (strftime('%s','now'))
	WHERE
		id = ?
`

// DeleteUser deletes a user by their username
func (s *Sqlite) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := s.database.ExecContext(ctx, archiveUserQuery, userID)
	return err
}

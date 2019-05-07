package postgres

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

func (p Postgres) scanUser(scan database.Scanner) (*models.User, error) {
	var x = &models.User{}
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

const adminUserExistsQuery = `
	SELECT EXISTS(SELECT id FROM users WHERE is_admin = true)
`

// AdminUserExists validates whether or not an admin user exists
func (p *Postgres) AdminUserExists(ctx context.Context) (bool, error) {
	var exists string

	err := p.database.QueryRowContext(ctx, adminUserExistsQuery).Scan(&exists)
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
		id = $1
`

// GetUser fetches a user by their username
func (p *Postgres) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	row := p.database.QueryRowContext(ctx, getUserQuery, userID)
	u, err := p.scanUser(row)
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
		username = $1
`

// GetUserByUsername fetches a user by their username
func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	row := p.database.QueryRowContext(ctx, getUserByUsernameQuery, username)
	u, err := p.scanUser(row)
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

// GetUserCount fetches a count of users from the postgres database that meet a particular filter
func (p *Postgres) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	err = p.database.QueryRowContext(ctx, getUserCountQuery).Scan(&count)
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
		archived_on IS NULL
	LIMIT $1
	OFFSET $2
`

// GetUsers fetches a list of users from the postgres database that meet a particular filter
func (p *Postgres) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	var list []models.User

	rows, err := p.database.QueryContext(ctx, getUsersQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		var user *models.User
		user, err = p.scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	count, err := p.GetUserCount(ctx, filter)
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
		$1, $2, $3, $4
	)
	RETURNING
		id,
		created_on
`

// CreateUser creates a user
func (p *Postgres) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	x := &models.User{
		Username:        input.Username,
		TwoFactorSecret: input.TwoFactorSecret,
		IsAdmin:         input.IsAdmin,
	}

	// create the user
	err := p.database.
		QueryRow(createUserQuery, input.Username, input.Password, input.TwoFactorSecret, input.IsAdmin).
		Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, errors.Wrap(err, "error executing user creation query")
	}

	return x, nil
}

const updateUserQuery = `
	UPDATE users
	SET
		username = $1,
		password = $2,
		updated_on = extract(epoch FROM NOW())
	WHERE
		id = $3
	RETURNING
		updated_on
`

// UpdateUser receives a complete User struct and updates its place in the database.
// NOTE this function uses the ID provided in the input to make its query.
func (p *Postgres) UpdateUser(ctx context.Context, input *models.User) error {
	// update the user
	err := p.database.
		QueryRow(updateUserQuery, input.Username, input.HashedPassword, input.ID).
		Scan(&input.UpdatedOn)

	return err
}

const archiveUserQuery = `
	UPDATE users
	SET
		updated_on = extract(epoch FROM NOW()),
		archived_on = extract(epoch FROM NOW())
	WHERE
		id = $1
	RETURNING
		archived_on
`

// DeleteUser deletes a user by their username
func (p *Postgres) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := p.database.ExecContext(ctx, archiveUserQuery, userID)
	return err
}

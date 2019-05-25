package postgres

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	usersTableName = "users"
)

var (
	usersTableColumns = []string{
		"id",
		"username",
		"hashed_password",
		"password_last_changed_on",
		"two_factor_secret",
		"is_admin",
		"created_on",
		"updated_on",
		"archived_on",
	}
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

func (p *Postgres) scanUsers(rows *sql.Rows) ([]models.User, error) {
	var list []models.User

	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		user, err := p.scanUser(rows)
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

// GetUserCount fetches a count of users from the postgres database that meet a particular filter
func (p *Postgres) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	builder := p.sqlBuilder.
		Select("COUNT(*)").
		From(usersTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"archived_on": nil,
		}))

	builder = filter.ApplyToQueryBuilder(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "generating query")
	}

	err = p.database.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

// GetUsers fetches a list of users from the postgres database that meet a particular filter
func (p *Postgres) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	builder := p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"archived_on": nil,
		}))

	builder = filter.ApplyToQueryBuilder(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "generating query")
	}

	rows, err := p.database.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	userList, err := p.scanUsers(rows)
	if err != nil {
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
		Users: userList,
	}

	return x, nil
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
		QueryRowContext(
			ctx,
			createUserQuery,
			input.Username,
			input.Password,
			input.TwoFactorSecret,
			input.IsAdmin,
		).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == pq.ErrorCode("23505") {
				return nil, dbclient.ErrUserExists
			}

		}

		return nil, errors.Wrap(err, "error executing user creation query")
	}

	return x, nil
}

const updateUserQuery = `
	UPDATE users
	SET
		username = $1,
		hashed_password = $2,
		two_factor_secret = $3,
		updated_on = extract(epoch FROM NOW())
	WHERE
		id = $4
	RETURNING
		updated_on
`

// UpdateUser receives a complete User struct and updates its place in the database.
// NOTE this function uses the ID provided in the input to make its query.
func (p *Postgres) UpdateUser(ctx context.Context, input *models.User) error {
	// update the user
	err := p.database.QueryRowContext(
		ctx,
		updateUserQuery,
		input.Username,
		input.HashedPassword,
		input.TwoFactorSecret,
		input.ID,
	).Scan(&input.UpdatedOn)

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

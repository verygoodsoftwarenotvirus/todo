package postgres

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	getUserQuery = `
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
	getUserCountQuery = `
		SELECT
			COUNT(*)
		FROM
			users
		WHERE archived_on is null
	`
	getUsersQuery = `
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
		LIMIT $1
		OFFSET $2
	`
	createUserQuery = `
		INSERT INTO users
		(
			username, hashed_password, two_factor_secret, is_admin
		)
		VALUES
		(
			$1, $2, $3, $4
		)
		RETURNING
			id, created_on
	`
	updateUserQuery = `
		UPDATE users SET
			username = $1,
			password = $2,
			updated_on = extract(epoch FROM NOW())
		WHERE id = $3
		RETURNING
			updated_on
	`
	archiveUserQuery = `
		UPDATE users SET
			updated_on = extract(epoch FROM NOW()),
			archived_on = extract(epoch FROM NOW())
		WHERE username = $1
		RETURNING
			archived_on
	`
)

func (p Postgres) scanUser(scan database.Scannable) (*models.User, error) {
	var (
		x = &models.User{}
	)

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

// GetUser fetches a user by their username
func (p *Postgres) GetUser(ctx context.Context, username string) (*models.User, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetUser")
	defer span.Finish()

	p.logger.WithValue("username", username).Debug("GetUser called")
	row := p.database.QueryRow(getUserQuery, username)
	u, err := p.scanUser(row)
	return u, err
}

// GetUserCount fetches a count of users from the postgres database that meet a particular filter
func (p *Postgres) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetUserCount")
	defer span.Finish()

	p.logger.WithValue("filter", filter).Debug("GetUserCount called")
	err = p.database.QueryRow(getUserCountQuery).Scan(&count)
	return
}

// GetUsers fetches a list of users from the postgres database that meet a particular filter
func (p *Postgres) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetUsers")
	defer span.Finish()

	p.logger.WithValue("filter", filter).Debug("GetUsers called")

	if filter == nil {
		p.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}
	var list []models.User

	rows, err := p.database.Query(getUsersQuery, filter.Limit, filter.QueryPage())
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

// CreateUser creates a user
func (p *Postgres) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "CreateUser")
	defer span.Finish()

	logger := p.logger.WithValues(map[string]interface{}{
		"username": input.Username,
		"is_admin": input.IsAdmin,
	})
	logger.Debug("CreateUser called")

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
		logger.Error(err, "error executing user creation query")
		return nil, err
	}

	p.logger.Debug("returning from CreateUser")
	return x, nil
}

// UpdateUser receives a complete User struct and updates its place in the database.
// NOTE this function uses the ID provided in the input to make its query.
func (p *Postgres) UpdateUser(ctx context.Context, input *models.User) error {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "UpdateUser")
	defer span.Finish()

	p.logger.WithValues(map[string]interface{}{
		"username": input.Username,
		"is_admin": input.IsAdmin,
	}).Debug("UpdateUser called")

	// update the user
	err := p.database.
		QueryRow(updateUserQuery, input.Username, input.HashedPassword, input.ID).
		Scan(&input.UpdatedOn)

	return err
}

// DeleteUser deletes a user by their username
func (p *Postgres) DeleteUser(ctx context.Context, username string) error {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "DeleteUser")
	defer span.Finish()

	p.logger.WithValue("username", username).Debug("DeleteUser called")
	_, err := p.database.Exec(archiveUserQuery, username)
	return err
}

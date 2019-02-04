package postgres

import (
	"context"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
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
	getUserQueryByID = `
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
			id = $1 AND archived_on is null
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
		LIMIT $
		OFFSET $
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
			updated_on = to_timestamp(extract(epoch FROM NOW()))
		WHERE id = $3
		RETURNING
			updated_on
	`
	archiveUserQuery = `
		UPDATE users SET
			updated_on = to_timestamp(extract(epoch FROM NOW())),
			archived_on = to_timestamp(extract(epoch FROM NOW()))
		WHERE username = $1
		RETURNING
			archived_on
	`
)

func scanUser(scan database.Scannable) (*models.User, error) {
	var (
		x = &models.User{}

		co time.Time
		uo *time.Time
		ao *time.Time
	)

	err := scan.Scan(
		&x.ID,
		&x.Username,
		&x.HashedPassword,
		&x.PasswordLastChangedOn,
		&x.TwoFactorSecret,
		&x.IsAdmin,
		&co,
		&uo,
		&ao,
	)
	if err != nil {
		return nil, err
	}

	x.CreatedOn = timeToUInt64(co)
	if uo != nil {
		x.UpdatedOn = timeToPUInt64(uo)
	}
	if ao != nil {
		x.ArchivedOn = timeToPUInt64(ao)
	}

	return x, nil
}

// GetUser fetches a user by their username
func (p *Postgres) GetUser(ctx context.Context, username string) (*models.User, error) {
	p.logger.WithField("username", username).Debugln("GetUser called")
	u, err := scanUser(p.database.QueryRow(getUserQuery, username))
	return u, err
}

// GetUserCount fetches a count of users from the postgres database that meet a particular filter
func (p *Postgres) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	p.logger.WithField("filter", filter).Debugln("GetUserCount called")
	err = p.database.QueryRow(getUserCountQuery).Scan(&count)
	return
}

// GetUsers fetches a list of users from the postgres database that meet a particular filter
func (p *Postgres) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	p.logger.WithField("filter", filter).Debugln("GetUsers called")

	if filter == nil {
		p.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}
	list := []models.User{}

	rows, err := p.database.Query(getUsersQuery, filter.Limit, filter.QueryPage())
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
	p.logger.WithFields(map[string]interface{}{
		"username": input.Username,
		"is_admin": input.IsAdmin,
	}).Debugln("CreateUser called")

	x := &models.User{
		Username:        input.Username,
		TwoFactorSecret: input.TwoFactorSecret,
		IsAdmin:         input.IsAdmin,
	}
	p.logger.Debugf("CreateUser called for %s", input.Username)

	// create the user
	var t time.Time
	if err := p.database.
		QueryRow(createUserQuery, input.Username, input.Password, input.TwoFactorSecret, input.IsAdmin).
		Scan(&x.ID, &t); err != nil {
		p.logger.Errorf("error executing user creation query: %v", err)
		return nil, err
	}
	x.CreatedOn = timeToUInt64(t)

	p.logger.Debugln("returning from CreateUser")
	return x, nil
}

// UpdateUser receives a complete User struct and updates its place in the database.
// NOTE this function uses the ID provided in the input to make its query.
func (p *Postgres) UpdateUser(ctx context.Context, input *models.User) error {
	p.logger.WithFields(map[string]interface{}{
		"username": input.Username,
		"is_admin": input.IsAdmin,
	}).Debugln("UpdateUser called")

	// update the user
	var t *time.Time
	if err := p.database.QueryRow(updateUserQuery, input.Username, input.HashedPassword, input.ID).Scan(&t); err != nil {
		return err
	}
	input.UpdatedOn = timeToPUInt64(t)

	return nil
}

// DeleteUser deletes a user by their username
func (p *Postgres) DeleteUser(ctx context.Context, username string) error {
	p.logger.WithField("username", username).Debugln("DeleteUser called")
	_, err := p.database.Exec(archiveUserQuery, username)
	return err
}

package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
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
		LIMIT ?
		OFFSET ?
	`
	createUserQuery = `
		INSERT INTO users
		(
			username, hashed_password, two_factor_secret, is_admin
		)
		VALUES
		(
			?, ?, ?, ?
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
		WHERE username = ?
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

func scanUsers(rows *sql.Rows) ([]models.User, error) {
	list := []models.User{}
	defer rows.Close()
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

// GetUser fetches a user by their username
func (s *Sqlite) GetUser(ctx context.Context, username string) (*models.User, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetUser")
	defer span.Finish()

	return scanUser(s.database.QueryRow(getUserQuery, username))
}

// GetUserCount fetches a count of users from the sqlite database that meet a particular filter
func (s *Sqlite) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetUserCount")
	defer span.Finish()

	return count, s.database.QueryRow(getUserCountQuery).Scan(&count)
}

// GetUsers fetches a list of users from the sqlite database that meet a particular filter
func (s *Sqlite) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetUsers")
	defer span.Finish()

	filter = s.prepareFilter(filter, span)
	rows, err := s.database.Query(getUsersQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, err
	}

	list, err := scanUsers(rows)
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

// CreateUser creates a user
func (s *Sqlite) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "CreateUser")
	defer span.Finish()

	logger := s.logger.WithValue("username", input.Username)
	logger.Debug("CreateUser called")

	// create the user
	res, err := s.database.Exec(createUserQuery, input.Username, input.Password, input.TwoFactorSecret, input.IsAdmin)
	if err != nil {
		logger.Error(err, "error executing user creation query")
		return nil, err
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error(err, "error fetching last inserted user ID")
		return nil, err
	}

	// fetch full updated user
	finalRow := s.database.QueryRow(getUserQueryByID, id)
	x, err := scanUser(finalRow)
	if err != nil {
		logger.Error(err, "error fetching newly created user %d: %v")
		return nil, err
	}

	logger.Debug("returning from CreateUser")
	return x, nil
}

// UpdateUser updates a user. Note that this function expects the provided user to have a valid ID.
func (s *Sqlite) UpdateUser(ctx context.Context, input *models.User) error {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "UpdateUser")
	defer span.Finish()

	// update the user
	if _, err := s.database.Exec(updateUserQuery, input.Username, input.HashedPassword, input.ID); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user by their username
func (s *Sqlite) DeleteUser(ctx context.Context, username string) error {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "DeleteUser")
	defer span.Finish()

	_, err := s.database.Exec(archiveUserQuery, username)
	return err
}

package postgres

import (
	"context"
	"database/sql"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

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
		"created_on",
		"updated_on",
		"archived_on",
	}
)

func scanUser(scan database.Scanner) (*models.User, error) {
	var x = &models.User{}
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

func scanUsers(logger logging.Logger, rows *sql.Rows) ([]models.User, error) {
	var list []models.User

	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(err, "closing rows")
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

func (p *Postgres) buildGetUserQuery(userID uint64) (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetUser fetches a user by their username
func (p *Postgres) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := p.buildGetUserQuery(userID)
	row := p.db.QueryRowContext(ctx, query, args...)
	u, err := scanUser(row)
	return u, err
}

func (p *Postgres) buildGetUserByUsernameQuery(username string) (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"username": username}).
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetUserByUsername fetches a user by their username
func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query, args := p.buildGetUserByUsernameQuery(username)
	row := p.db.QueryRowContext(ctx, query, args...)
	u, err := scanUser(row)
	return u, err
}

// GetUserCount fetches a count of users from the postgres db that meet a particular filter
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

	err = p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

func (p *Postgres) buildGetUsersQuery(filter *models.QueryFilter) (string, []interface{}) {
	builder := p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"archived_on": nil,
		}))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err := builder.ToSql()
	logQueryBuildingError(p.logger, err)
	return query, args
}

// GetUsers fetches a list of users from the postgres db that meet a particular filter
func (p *Postgres) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	query, args := p.buildGetUsersQuery(filter)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	userList, err := scanUsers(p.logger, rows)
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

func (p *Postgres) buildCreateUserQuery(input *models.UserInput) (string, []interface{}) {
	query, args, err := p.sqlBuilder.Insert(usersTableName).
		Columns(
			"username",
			"hashed_password",
			"two_factor_secret",
		).
		Values(
			input.Username,
			input.Password,
			input.TwoFactorSecret,
		).
		Suffix("RETURNING id, created_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// CreateUser creates a user
func (p *Postgres) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	x := &models.User{
		Username:        input.Username,
		TwoFactorSecret: input.TwoFactorSecret,
	}

	query, args := p.buildCreateUserQuery(input)

	// create the user
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if e.Code == pq.ErrorCode("23505") {
				return nil, dbclient.ErrUserExists
			}
		default:
			return nil, errors.Wrap(err, "error executing user creation query")
		}
	}

	return x, nil
}

func (p *Postgres) buildUpdateUserQuery(input *models.User) (string, []interface{}) {
	query, args, err := p.sqlBuilder.Update(usersTableName).
		Set("username", input.Username).
		Set("hashed_password", input.HashedPassword).
		Set("two_factor_secret", input.TwoFactorSecret).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{"id": input.ID}).
		Suffix("RETURNING updated_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// anonymous structs or incomplete models at your peril.
func (p *Postgres) UpdateUser(ctx context.Context, input *models.User) error {
	query, args := p.buildUpdateUserQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.UpdatedOn)
}

func (p *Postgres) buildArchiveUserQuery(userID uint64) (string, []interface{}) {
	query, args, err := p.sqlBuilder.Update(usersTableName).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Set("archived_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{"id": userID}).
		Suffix("RETURNING archived_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// DeleteUser deletes a user by their username
func (p *Postgres) DeleteUser(ctx context.Context, userID uint64) error {
	query, args := p.buildArchiveUserQuery(userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

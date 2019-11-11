package postgres

import (
	"context"
	"database/sql"
	"fmt"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
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

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct
func scanUser(scan database.Scanner) (*models.User, error) {
	var x = &models.User{}

	if err := scan.Scan(
		&x.ID,
		&x.Username,
		&x.HashedPassword,
		&x.PasswordLastChangedOn,
		&x.TwoFactorSecret,
		&x.IsAdmin,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
	); err != nil {
		return nil, err
	}

	return x, nil
}

// scanUsers takes database rows and loads them into a slice of User structs
func scanUsers(logger logging.Logger, rows *sql.Rows) ([]models.User, error) {
	var list []models.User

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning user result: %w", err)
		}
		list = append(list, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID
func (p *Postgres) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user
func (p *Postgres) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := p.buildGetUserQuery(userID)
	row := p.db.QueryRowContext(ctx, query, args...)
	u, err := scanUser(row)

	if err != nil {
		return nil, buildError(err, "fetching user from database")
	}

	return u, err
}

// buildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username
func (p *Postgres) buildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"username": username}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username
func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query, args := p.buildGetUserByUsernameQuery(username)
	row := p.db.QueryRowContext(ctx, query, args...)
	u, err := scanUser(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, nil
}

// buildGetUserCountQuery returns a SQL query (and arguments) for retrieving the number of users who adhere
// to a given filter's criteria.
func (p *Postgres) buildGetUserCountQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
		Select(CountQuery).
		From(usersTableName).
		Where(squirrel.Eq{"archived_on": nil})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}
	query, args, err = builder.ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetUserCount fetches a count of users from the database that meet a particular filter
func (p *Postgres) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	query, args := p.buildGetUserCountQuery(filter)
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

// buildGetUserCountQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (p *Postgres) buildGetUsersQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"archived_on": nil})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter
func (p *Postgres) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	query, args := p.buildGetUsersQuery(filter)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying for user")
	}

	userList, err := scanUsers(p.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	count, err := p.GetUserCount(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("fetching user count: %w", err)
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

// buildCreateUserQuery returns a SQL query (and arguments) that would create a given User
func (p *Postgres) buildCreateUserQuery(input *models.UserInput) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Insert(usersTableName).
		Columns(
			"username",
			"hashed_password",
			"two_factor_secret",
			"is_admin",
		).
		Values(
			input.Username,
			input.Password,
			input.TwoFactorSecret,
			false,
		).
		Suffix("RETURNING id, created_on").
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should also be no way to update a user via this structure
	// such that they would have admin privileges

	p.logQueryBuildingError(err)

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
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if e.Code == pq.ErrorCode("23505") {
				return nil, dbclient.ErrUserExists
			}
		default:
			return nil, fmt.Errorf("error executing user creation query: %w", err)
		}
	}

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row
func (p *Postgres) buildUpdateUserQuery(input *models.User) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Update(usersTableName).
		Set("username", input.Username).
		Set("hashed_password", input.HashedPassword).
		Set("two_factor_secret", input.TwoFactorSecret).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{"id": input.ID}).
		Suffix("RETURNING updated_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// anonymous structs or incomplete models at your peril.
func (p *Postgres) UpdateUser(ctx context.Context, input *models.User) error {
	query, args := p.buildUpdateUserQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.UpdatedOn)
}

func (p *Postgres) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Update(usersTableName).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{"id": userID}).
		Suffix("RETURNING archived_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser archives a user by their username
func (p *Postgres) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := p.buildArchiveUserQuery(userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

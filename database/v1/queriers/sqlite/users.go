package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
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
func (s *Sqlite) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user
func (s *Sqlite) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := s.buildGetUserQuery(userID)
	row := s.db.QueryRowContext(ctx, query, args...)
	u, err := scanUser(row)

	if err != nil {
		return nil, buildError(err, "fetching user from database")
	}

	return u, err
}

// buildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username
func (s *Sqlite) buildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"username": username}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username
func (s *Sqlite) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query, args := s.buildGetUserByUsernameQuery(username)
	row := s.db.QueryRowContext(ctx, query, args...)
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
func (s *Sqlite) buildGetUserCountQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error
	builder := s.sqlBuilder.
		Select(CountQuery).
		From(usersTableName).
		Where(squirrel.Eq{"archived_on": nil})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}
	query, args, err = builder.ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUserCount fetches a count of users from the database that meet a particular filter
func (s *Sqlite) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	query, args := s.buildGetUserCountQuery(filter)
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

// buildGetUserCountQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (s *Sqlite) buildGetUsersQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error
	builder := s.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{"archived_on": nil})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter
func (s *Sqlite) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	query, args := s.buildGetUsersQuery(filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying for user")
	}

	userList, err := scanUsers(s.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	count, err := s.GetUserCount(ctx, filter)
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
func (s *Sqlite) buildCreateUserQuery(input *models.UserInput) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
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
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should also be no way to update a user via this structure
	// such that they would have admin privileges

	s.logQueryBuildingError(err)

	return query, args
}

// buildUserCreationTimeQuery returns a SQL query (and arguments) that would create a given User
func (s *Sqlite) buildUserCreationTimeQuery(userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.Select("created_on").
		From(usersTableName).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user
func (s *Sqlite) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	x := &models.User{
		Username:        input.Username,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := s.buildCreateUserQuery(input)

	// create the user
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing user creation query: %w", err)
	}

	// fetch the last inserted ID
	if id, idErr := res.LastInsertId(); idErr == nil {
		x.ID = uint64(id)

		query, args := s.buildUserCreationTimeQuery(x.ID)
		s.logCreationTimeRetrievalError(s.db.QueryRowContext(ctx, query, args...).Scan(&x.CreatedOn))
	}

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row
func (s *Sqlite) buildUpdateUserQuery(input *models.User) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Update(usersTableName).
		Set("username", input.Username).
		Set("hashed_password", input.HashedPassword).
		Set("two_factor_secret", input.TwoFactorSecret).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{"id": input.ID}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// anonymous structs or incomplete models at your peril.
func (s *Sqlite) UpdateUser(ctx context.Context, input *models.User) error {
	query, args := s.buildUpdateUserQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Sqlite) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Update(usersTableName).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser archives a user by their username
func (s *Sqlite) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := s.buildArchiveUserQuery(userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

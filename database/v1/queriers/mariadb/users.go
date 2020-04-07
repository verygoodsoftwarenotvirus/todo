package mariadb

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
		fmt.Sprintf("%s.id", usersTableName),
		fmt.Sprintf("%s.username", usersTableName),
		fmt.Sprintf("%s.hashed_password", usersTableName),
		fmt.Sprintf("%s.password_last_changed_on", usersTableName),
		fmt.Sprintf("%s.two_factor_secret", usersTableName),
		fmt.Sprintf("%s.is_admin", usersTableName),
		fmt.Sprintf("%s.created_on", usersTableName),
		fmt.Sprintf("%s.updated_on", usersTableName),
		fmt.Sprintf("%s.archived_on", usersTableName),
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
func (m *MariaDB) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", usersTableName): userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user
func (m *MariaDB) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := m.buildGetUserQuery(userID)
	row := m.db.QueryRowContext(ctx, query, args...)
	u, err := scanUser(row)

	if err != nil {
		return nil, buildError(err, "fetching user from database")
	}

	return u, err
}

// buildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username
func (m *MariaDB) buildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.username", usersTableName): username,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username
func (m *MariaDB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query, args := m.buildGetUserByUsernameQuery(username)
	row := m.db.QueryRowContext(ctx, query, args...)
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
func (m *MariaDB) buildGetUserCountQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(fmt.Sprintf(countQuery, usersTableName)).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", usersTableName): nil,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, usersTableName)
	}
	query, args, err = builder.ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetAllUserCount fetches a count of users from the database that meet a particular filter
func (m *MariaDB) GetAllUserCount(ctx context.Context) (count uint64, err error) {
	query, args := m.buildGetUserCountQuery(nil)
	err = m.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

// GetUserCount fetches a count of users from the database that meet a particular filter
func (m *MariaDB) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	query, args := m.buildGetUserCountQuery(filter)
	err = m.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

// buildGetUserCountQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (m *MariaDB) buildGetUsersQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", usersTableName): nil,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, usersTableName)
	}

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter
func (m *MariaDB) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	query, args := m.buildGetUsersQuery(filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying for user")
	}

	userList, err := scanUsers(m.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	count, err := m.GetUserCount(ctx, filter)
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
func (m *MariaDB) buildCreateUserQuery(input models.UserDatabaseCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Insert(usersTableName).
		Columns(
			"username",
			"hashed_password",
			"two_factor_secret",
			"is_admin",
			"created_on",
		).
		Values(
			input.Username,
			input.HashedPassword,
			input.TwoFactorSecret,
			false,
			squirrel.Expr(currentUnixTimeQuery),
		).
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should also be no way to update a user via this structure
	// such that they would have admin privileges

	m.logQueryBuildingError(err)

	return query, args
}

// buildUserCreationTimeQuery returns a SQL query (and arguments) that would create a given User
func (m *MariaDB) buildUserCreationTimeQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.Select(fmt.Sprintf("%s.created_on", usersTableName)).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", usersTableName): userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user
func (m *MariaDB) CreateUser(ctx context.Context, input models.UserDatabaseCreationInput) (*models.User, error) {
	x := &models.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := m.buildCreateUserQuery(input)

	// create the user
	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing user creation query: %w", err)
	}

	// fetch the last inserted ID
	if id, idErr := res.LastInsertId(); idErr == nil {
		x.ID = uint64(id)

		query, args := m.buildUserCreationTimeQuery(x.ID)
		m.logCreationTimeRetrievalError(m.db.QueryRowContext(ctx, query, args...).Scan(&x.CreatedOn))
	}

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row
func (m *MariaDB) buildUpdateUserQuery(input *models.User) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(usersTableName).
		Set("username", input.Username).
		Set("hashed_password", input.HashedPassword).
		Set("two_factor_secret", input.TwoFactorSecret).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{"id": input.ID}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// anonymous structs or incomplete models at your peril.
func (m *MariaDB) UpdateUser(ctx context.Context, input *models.User) error {
	query, args := m.buildUpdateUserQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

func (m *MariaDB) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(usersTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser archives a user by their username
func (m *MariaDB) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := m.buildArchiveUserQuery(userID)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

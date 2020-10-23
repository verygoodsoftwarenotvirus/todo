package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
)

const (
	usersTableName                         = "users"
	usersTableUsernameColumn               = "username"
	usersTableHashedPasswordColumn         = "hashed_password"
	usersTableSaltColumn                   = "salt"
	usersTableRequiresPasswordChangeColumn = "requires_password_change"
	usersTablePasswordLastChangedOnColumn  = "password_last_changed_on"
	usersTableTwoFactorColumn              = "two_factor_secret"
	usersTableTwoFactorVerifiedOnColumn    = "two_factor_secret_verified_on"
	usersTableIsAdminColumn                = "is_admin"
)

var (
	usersTableColumns = []string{
		fmt.Sprintf("%s.%s", usersTableName, idColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableUsernameColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableHashedPasswordColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableSaltColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableRequiresPasswordChangeColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTablePasswordLastChangedOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableIsAdminColumn),
		fmt.Sprintf("%s.%s", usersTableName, createdOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, lastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn),
	}
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (m *MariaDB) scanUser(scan database.Scanner) (*models.User, error) {
	var (
		x = &models.User{}
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Username,
		&x.HashedPassword,
		&x.Salt,
		&x.RequiresPasswordChange,
		&x.PasswordLastChangedOn,
		&x.TwoFactorSecret,
		&x.TwoFactorSecretVerifiedOn,
		&x.IsAdmin,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	return x, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (m *MariaDB) scanUsers(rows database.ResultIterator) ([]models.User, error) {
	var (
		list []models.User
	)

	for rows.Next() {
		user, err := m.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning user result: %w", err)
		}

		list = append(list, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		m.logger.Error(err, "closing rows")
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
			fmt.Sprintf("%s.%s", usersTableName, idColumn):         userID,
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user.
func (m *MariaDB) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := m.buildGetUserQuery(userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	u, err := m.scanUser(row)
	if err != nil {
		return nil, buildError(err, "fetching user from database")
	}

	return u, err
}

// buildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret
func (m *MariaDB) buildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, idColumn):                            userID,
			fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn):                    nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret
func (m *MariaDB) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := m.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	u, err := m.scanUser(row)
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
			fmt.Sprintf("%s.%s", usersTableName, usersTableUsernameColumn): username,
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn):         nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username.
func (m *MariaDB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query, args := m.buildGetUserByUsernameQuery(username)
	row := m.db.QueryRowContext(ctx, query, args...)

	u, err := m.scanUser(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, nil
}

// buildGetAllUsersCountQuery returns a SQL query (and arguments) for retrieving the number of users who adhere
// to a given filter's criteria.
func (m *MariaDB) buildGetAllUsersCountQuery() (query string) {
	var err error

	builder := m.sqlBuilder.
		Select(fmt.Sprintf(countQuery, usersTableName)).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
		})

	query, _, err = builder.ToSql()

	m.logQueryBuildingError(err)

	return query
}

// GetAllUsersCount fetches a count of users from the database.
func (m *MariaDB) GetAllUsersCount(ctx context.Context) (count uint64, err error) {
	query := m.buildGetAllUsersCountQuery()
	err = m.db.QueryRowContext(ctx, query).Scan(&count)
	return
}

// buildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (m *MariaDB) buildGetUsersQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", usersTableName, idColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, usersTableName)
	}

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (m *MariaDB) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	query, args := m.buildGetUsersQuery(filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying for user")
	}

	userList, err := m.scanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	x := &models.UserList{
		Pagination: models.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
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
			usersTableUsernameColumn,
			usersTableHashedPasswordColumn,
			usersTableSaltColumn,
			usersTableTwoFactorColumn,
			usersTableIsAdminColumn,
		).
		Values(
			input.Username,
			input.HashedPassword,
			input.Salt,
			input.TwoFactorSecret,
			false,
		).
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should also be no way to update a user via this structure
	// such that they would have admin privileges.

	m.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user.
func (m *MariaDB) CreateUser(ctx context.Context, input models.UserDatabaseCreationInput) (*models.User, error) {
	x := &models.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := m.buildCreateUserQuery(input)

	// create the user.
	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing user creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	m.logIDRetrievalError(err)
	x.ID = uint64(id)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = m.timeTeller.Now()

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row
func (m *MariaDB) buildUpdateUserQuery(input *models.User) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(usersTableName).
		Set(usersTableUsernameColumn, input.Username).
		Set(usersTableHashedPasswordColumn, input.HashedPassword).
		Set(usersTableSaltColumn, input.Salt).
		Set(usersTableTwoFactorColumn, input.TwoFactorSecret).
		Set(usersTableTwoFactorVerifiedOnColumn, input.TwoFactorSecretVerifiedOn).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: input.ID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// incomplete models at your peril.
func (m *MariaDB) UpdateUser(ctx context.Context, input *models.User) error {
	query, args := m.buildUpdateUserQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's password.
func (m *MariaDB) buildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(usersTableName).
		Set(usersTableHashedPasswordColumn, newHash).
		Set(usersTableRequiresPasswordChangeColumn, false).
		Set(usersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateUserPassword updates a user's password.
func (m *MariaDB) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	query, args := m.buildUpdateUserPasswordQuery(userID, newHash)

	_, err := m.db.ExecContext(ctx, query, args...)

	return err
}

// buildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret
func (m *MariaDB) buildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(usersTableName).
		Set(usersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (m *MariaDB) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	query, args := m.buildVerifyUserTwoFactorSecretQuery(userID)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveUserQuery builds a SQL query that marks a user as archived.
func (m *MariaDB) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(usersTableName).
		Set(archivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser marks a user as archived.
func (m *MariaDB) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := m.buildArchiveUserQuery(userID)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

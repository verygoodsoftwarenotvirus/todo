package sqlite

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
	usersTableRequiresPasswordChangeColumn = "requires_password_change"
	usersTablePasswordLastChangedOnColumn  = "password_last_changed_on"
	usersTableTwoFactorColumn              = "two_factor_secret"
	usersTableIsAdminColumn                = "is_admin"
	usersTableTwoFactorVerifiedOnColumn    = "two_factor_secret_verified_on"
)

var (
	usersTableColumns = []string{
		fmt.Sprintf("%s.%s", usersTableName, idColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableUsernameColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableHashedPasswordColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableRequiresPasswordChangeColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTablePasswordLastChangedOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableIsAdminColumn),
		fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, createdOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, lastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn),
	}
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (s *Sqlite) scanUser(scan database.Scanner) (*models.User, error) {
	var (
		x = &models.User{}
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Username,
		&x.HashedPassword,
		&x.RequiresPasswordChange,
		&x.PasswordLastChangedOn,
		&x.TwoFactorSecret,
		&x.IsAdmin,
		&x.TwoFactorSecretVerifiedOn,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	return x, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (s *Sqlite) scanUsers(rows database.ResultIterator) ([]models.User, error) {
	var (
		list []models.User
	)

	for rows.Next() {
		user, err := s.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning user result: %w", err)
		}

		list = append(list, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		s.logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID
func (s *Sqlite) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
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

	s.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user.
func (s *Sqlite) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := s.buildGetUserQuery(userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, err := s.scanUser(row)
	if err != nil {
		return nil, buildError(err, "fetching user from database")
	}

	return u, err
}

// buildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret
func (s *Sqlite) buildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, idColumn):                            userID,
			fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn):                    nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret
func (s *Sqlite) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := s.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, err := s.scanUser(row)
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
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, usersTableUsernameColumn): username,
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn):         nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.two_factor_secret_verified_on", usersTableName): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username.
func (s *Sqlite) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query, args := s.buildGetUserByUsernameQuery(username)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, err := s.scanUser(row)
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
func (s *Sqlite) buildGetAllUsersCountQuery() (query string) {
	var err error

	builder := s.sqlBuilder.
		Select(fmt.Sprintf(countQuery, usersTableName)).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
		})

	query, _, err = builder.ToSql()

	s.logQueryBuildingError(err)

	return query
}

// GetAllUsersCount fetches a count of users from the database.
func (s *Sqlite) GetAllUsersCount(ctx context.Context) (count uint64, err error) {
	query := s.buildGetAllUsersCountQuery()
	err = s.db.QueryRowContext(ctx, query).Scan(&count)
	return
}

// buildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (s *Sqlite) buildGetUsersQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
		}).
		GroupBy(fmt.Sprintf("%s.id", usersTableName))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, usersTableName)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (s *Sqlite) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	query, args := s.buildGetUsersQuery(filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying for user")
	}

	userList, err := s.scanUsers(rows)
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
func (s *Sqlite) buildCreateUserQuery(input models.UserDatabaseCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Insert(usersTableName).
		Columns(
			usersTableUsernameColumn,
			usersTableHashedPasswordColumn,
			usersTableTwoFactorColumn,
			usersTableIsAdminColumn,
		).
		Values(
			input.Username,
			input.HashedPassword,
			input.TwoFactorSecret,
			false,
		).
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should also be no way to update a user via this structure
	// such that they would have admin privileges.

	s.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user.
func (s *Sqlite) CreateUser(ctx context.Context, input models.UserDatabaseCreationInput) (*models.User, error) {
	x := &models.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := s.buildCreateUserQuery(input)

	// create the user.
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing user creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	s.logIDRetrievalError(err)
	x.ID = uint64(id)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = s.timeTeller.Now()

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row
func (s *Sqlite) buildUpdateUserQuery(input *models.User) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(usersTableName).
		Set(usersTableUsernameColumn, input.Username).
		Set(usersTableHashedPasswordColumn, input.HashedPassword).
		Set(usersTableTwoFactorColumn, input.TwoFactorSecret).
		Set(usersTableTwoFactorVerifiedOnColumn, input.TwoFactorSecretVerifiedOn).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: input.ID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row
func (s *Sqlite) buildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(usersTableName).
		Set(usersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (s *Sqlite) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	query, args := s.buildVerifyUserTwoFactorSecretQuery(userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// anonymous structs or incomplete models at your peril.
func (s *Sqlite) UpdateUser(ctx context.Context, input *models.User) error {
	query, args := s.buildUpdateUserQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// buildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's password
func (s *Sqlite) buildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(usersTableName).
		Set(usersTableHashedPasswordColumn, newHash).
		Set(usersTableRequiresPasswordChangeColumn, false).
		Set(usersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateUserPassword updates a user's password
func (s *Sqlite) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	query, args := s.buildUpdateUserPasswordQuery(userID, newHash)

	_, err := s.db.ExecContext(ctx, query, args...)

	return err
}

func (s *Sqlite) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(usersTableName).
		Set(archivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser archives a user by their username.
func (s *Sqlite) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := s.buildArchiveUserQuery(userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

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
		fmt.Sprintf("%s.two_factor_secret_verified_on", usersTableName),
		fmt.Sprintf("%s.created_on", usersTableName),
		fmt.Sprintf("%s.updated_on", usersTableName),
		fmt.Sprintf("%s.archived_on", usersTableName),
	}
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (s *Sqlite) scanUser(scan database.Scanner, includeCount bool) (*models.User, uint64, error) {
	var (
		x     = &models.User{}
		count uint64
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Username,
		&x.HashedPassword,
		&x.PasswordLastChangedOn,
		&x.TwoFactorSecret,
		&x.IsAdmin,
		&x.TwoFactorSecretVerifiedOn,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	return x, count, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (s *Sqlite) scanUsers(rows database.ResultIterator) ([]models.User, uint64, error) {
	var (
		list  []models.User
		count uint64
	)

	for rows.Next() {
		user, c, err := s.scanUser(rows, true)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning user result: %w", err)
		}

		if count == 0 {
			count = c
		}

		list = append(list, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := rows.Close(); err != nil {
		s.logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID
func (s *Sqlite) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", usersTableName): userID,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.two_factor_secret_verified_on", usersTableName): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user.
func (s *Sqlite) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := s.buildGetUserQuery(userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, _, err := s.scanUser(row, false)
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
			fmt.Sprintf("%s.id", usersTableName):                            userID,
			fmt.Sprintf("%s.two_factor_secret_verified_on", usersTableName): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret
func (s *Sqlite) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := s.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, _, err := s.scanUser(row, false)
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
			fmt.Sprintf("%s.username", usersTableName): username,
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

	u, _, err := s.scanUser(row, false)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, nil
}

// buildGetAllUserCountQuery returns a SQL query (and arguments) for retrieving the number of users who adhere
// to a given filter's criteria.
func (s *Sqlite) buildGetAllUserCountQuery() (query string) {
	var err error

	builder := s.sqlBuilder.
		Select(fmt.Sprintf(countQuery, usersTableName)).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", usersTableName): nil,
		})

	query, _, err = builder.ToSql()

	s.logQueryBuildingError(err)

	return query
}

// GetAllUserCount fetches a count of users from the database.
func (s *Sqlite) GetAllUserCount(ctx context.Context) (count uint64, err error) {
	query := s.buildGetAllUserCountQuery()
	err = s.db.QueryRowContext(ctx, query).Scan(&count)
	return
}

// buildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (s *Sqlite) buildGetUsersQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
		Select(append(usersTableColumns, fmt.Sprintf(countQuery, usersTableName))...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", usersTableName): nil,
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

	userList, count, err := s.scanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
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
func (s *Sqlite) buildCreateUserQuery(input models.UserDatabaseCreationInput) (query string, args []interface{}) {
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
		Set("username", input.Username).
		Set("hashed_password", input.HashedPassword).
		Set("two_factor_secret", input.TwoFactorSecret).
		Set("two_factor_secret_verified_on", input.TwoFactorSecretVerifiedOn).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id": input.ID,
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
		Set("two_factor_secret_verified_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id": userID,
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

func (s *Sqlite) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(usersTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id": userID,
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

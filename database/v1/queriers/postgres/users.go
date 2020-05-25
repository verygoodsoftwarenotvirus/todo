package postgres

import (
	"context"
	"database/sql"
	"fmt"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
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
func (p *Postgres) scanUser(scan database.Scanner, includeCount bool) (*models.User, uint64, error) {
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
func (p *Postgres) scanUsers(rows database.ResultIterator) ([]models.User, uint64, error) {
	var (
		list  []models.User
		count uint64
	)

	for rows.Next() {
		user, c, err := p.scanUser(rows, true)
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
		p.logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID
func (p *Postgres) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", usersTableName): userID,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.two_factor_secret_verified_on", usersTableName): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user.
func (p *Postgres) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := p.buildGetUserQuery(userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	u, _, err := p.scanUser(row, false)
	if err != nil {
		return nil, buildError(err, "fetching user from database")
	}

	return u, err
}

// buildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret
func (p *Postgres) buildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", usersTableName):                            userID,
			fmt.Sprintf("%s.two_factor_secret_verified_on", usersTableName): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret
func (p *Postgres) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*models.User, error) {
	query, args := p.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	u, _, err := p.scanUser(row, false)
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
		Where(squirrel.Eq{
			fmt.Sprintf("%s.username", usersTableName): username,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.two_factor_secret_verified_on", usersTableName): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username.
func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query, args := p.buildGetUserByUsernameQuery(username)
	row := p.db.QueryRowContext(ctx, query, args...)

	u, _, err := p.scanUser(row, false)
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
func (p *Postgres) buildGetAllUserCountQuery() (query string) {
	var err error

	builder := p.sqlBuilder.
		Select(fmt.Sprintf(countQuery, usersTableName)).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.archived_on", usersTableName): nil,
		})

	query, _, err = builder.ToSql()

	p.logQueryBuildingError(err)

	return query
}

// GetAllUserCount fetches a count of users from the database.
func (p *Postgres) GetAllUserCount(ctx context.Context) (count uint64, err error) {
	query := p.buildGetAllUserCountQuery()
	err = p.db.QueryRowContext(ctx, query).Scan(&count)
	return
}

// buildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (p *Postgres) buildGetUsersQuery(filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := p.sqlBuilder.
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
	p.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (p *Postgres) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	query, args := p.buildGetUsersQuery(filter)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, buildError(err, "querying for user")
	}

	userList, count, err := p.scanUsers(rows)
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
func (p *Postgres) buildCreateUserQuery(input models.UserDatabaseCreationInput) (query string, args []interface{}) {
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
			input.HashedPassword,
			input.TwoFactorSecret,
			false,
		).
		Suffix("RETURNING id, created_on").
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should also be no way to update a user via this structure
	// such that they would have admin privileges.

	p.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user.
func (p *Postgres) CreateUser(ctx context.Context, input models.UserDatabaseCreationInput) (*models.User, error) {
	x := &models.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := p.buildCreateUserQuery(input)

	// create the user.
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		switch e := err.(type) {
		case *postgres.Error:
			if e.Code == postgres.ErrorCode(postgresRowExistsErrorCode) {
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
		Set("two_factor_secret_verified_on", input.TwoFactorSecretVerifiedOn).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id": input.ID,
		}).
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

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row
func (p *Postgres) buildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(usersTableName).
		Set("two_factor_secret_verified_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id": userID,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (p *Postgres) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	query, args := p.buildVerifyUserTwoFactorSecretQuery(userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

func (p *Postgres) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(usersTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id": userID,
		}).
		Suffix("RETURNING archived_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser archives a user by their username.
func (p *Postgres) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := p.buildArchiveUserQuery(userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

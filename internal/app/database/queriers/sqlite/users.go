package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

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
func (s *Sqlite) scanUser(scan database.Scanner) (*types.User, error) {
	var (
		x = &types.User{}
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
func (s *Sqlite) scanUsers(rows database.ResultIterator) ([]types.User, error) {
	var (
		list []types.User
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

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
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
func (s *Sqlite) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := s.buildGetUserQuery(userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, err := s.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret.
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

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret.
func (s *Sqlite) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := s.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, err := s.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
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
			fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username.
func (s *Sqlite) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	query, args := s.buildGetUserByUsernameQuery(username)
	row := s.db.QueryRowContext(ctx, query, args...)

	u, err := s.scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
func (s *Sqlite) buildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
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
	s.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (s *Sqlite) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	query, args := s.buildGetUsersQuery(filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	userList, err := s.scanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	x := &types.UserList{
		Pagination: types.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Users: userList,
	}

	return x, nil
}

// buildCreateUserQuery returns a SQL query (and arguments) that would create a given User.
func (s *Sqlite) buildCreateUserQuery(input types.UserDatabaseCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
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

	s.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user.
func (s *Sqlite) CreateUser(ctx context.Context, input types.UserDatabaseCreationInput) (*types.User, error) {
	x := &types.User{
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

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = s.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (s *Sqlite) buildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
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

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// incomplete types at your peril.
func (s *Sqlite) UpdateUser(ctx context.Context, input *types.User) error {
	query, args := s.buildUpdateUserQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// buildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's password.
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

// UpdateUserPassword updates a user's password.
func (s *Sqlite) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	query, args := s.buildUpdateUserPasswordQuery(userID, newHash)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// buildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
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

// buildArchiveUserQuery builds a SQL query that marks a user as archived.
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

// ArchiveUser marks a user as archived.
func (s *Sqlite) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := s.buildArchiveUserQuery(userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (s *Sqlite) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildCycleCookieSecretEvent(userID))
}

// LogSuccessfulLoginEvent saves a SuccessfulLoginEvent in the audit log table.
func (s *Sqlite) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildSuccessfulLoginEventEntry(userID))
}

// LogUnsuccessfulLoginBadPasswordEvent saves a UnsuccessfulLoginBadPasswordEvent in the audit log table.
func (s *Sqlite) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(userID))
}

// LogUnsuccessfulLoginBad2FATokenEvent saves a UnsuccessfulLoginBad2FATokenEvent in the audit log table.
func (s *Sqlite) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(userID))
}

// LogLogoutEvent saves a LogoutEvent in the audit log table.
func (s *Sqlite) LogLogoutEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildLogoutEventEntry(userID))
}

// LogUserCreationEvent saves a UserCreationEvent in the audit log table.
func (s *Sqlite) LogUserCreationEvent(ctx context.Context, user *types.User) {
	s.createAuditLogEntry(ctx, audit.BuildUserCreationEventEntry(user))
}

// LogUserVerifyTwoFactorSecretEvent saves a UserVerifyTwoFactorSecretEvent in the audit log table.
func (s *Sqlite) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID))
}

// LogUserUpdateTwoFactorSecretEvent saves a UserUpdateTwoFactorSecretEvent in the audit log table.
func (s *Sqlite) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID))
}

// LogUserUpdatePasswordEvent saves a UserUpdatePasswordEvent in the audit log table.
func (s *Sqlite) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUserUpdatePasswordEventEntry(userID))
}

// LogUserArchiveEvent saves a UserArchiveEvent in the audit log table.
func (s *Sqlite) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUserArchiveEventEntry(userID))
}

// buildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (s *Sqlite) buildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	userIDKey := fmt.Sprintf("json_extract(%s.%s, '$.%s')", auditLogEntriesTableName, auditLogEntriesTableContextColumn, audit.UserAssignmentKey)
	performedByIDKey := fmt.Sprintf("json_extract(%s.%s, '$.%s')", auditLogEntriesTableName, auditLogEntriesTableContextColumn, audit.ActorAssignmentKey)
	builder := s.sqlBuilder.
		Select(auditLogEntriesTableColumns...).
		From(auditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Eq{userIDKey: userID},
			squirrel.Eq{performedByIDKey: userID},
		}).
		OrderBy(fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn))

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForUser fetches an audit log entry from the database.
func (s *Sqlite) GetAuditLogEntriesForUser(ctx context.Context, itemID uint64) ([]types.AuditLogEntry, error) {
	query, args := s.buildGetAuditLogEntriesForUserQuery(itemID)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := s.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

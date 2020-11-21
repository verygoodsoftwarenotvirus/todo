package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (m *MariaDB) scanUser(scan database.Scanner) (*types.User, error) {
	var (
		x     = &types.User{}
		perms uint32
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
		&perms,
		&x.AccountStatus,
		&x.AccountStatusExplanation,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}
	x.AdminPermissions = bitmask.NewPermissionBitmask(perms)

	return x, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (m *MariaDB) scanUsers(rows database.ResultIterator) ([]types.User, error) {
	var (
		list []types.User
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

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (m *MariaDB) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.IDColumn):         userID,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user.
func (m *MariaDB) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := m.buildGetUserQuery(userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	u, err := m.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret.
func (m *MariaDB) buildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.IDColumn):                            userID,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn):                    nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret.
func (m *MariaDB) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := m.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	u, err := m.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (m *MariaDB) buildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableUsernameColumn): username,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn):         nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username.
func (m *MariaDB) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	query, args := m.buildGetUserByUsernameQuery(username)
	row := m.db.QueryRowContext(ctx, query, args...)

	u, err := m.scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, nil
}

// buildSearchForUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (m *MariaDB) buildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Expr(
			fmt.Sprintf("%s.%s LIKE ?", queriers.UsersTableName, queriers.UsersTableUsernameColumn),
			fmt.Sprintf("%s%%", usernameQuery),
		)).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (m *MariaDB) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]types.User, error) {
	query, args := m.buildSearchForUserByUsernameQuery(usernameQuery)
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying database for users: %w", err)
	}

	u, err := m.scanUsers(rows)
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
func (m *MariaDB) buildGetAllUsersCountQuery() (query string) {
	var err error

	builder := m.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.UsersTableName)).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
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
func (m *MariaDB) buildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.IDColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, queriers.UsersTableName)
	}

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (m *MariaDB) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	query, args := m.buildGetUsersQuery(filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	userList, err := m.scanUsers(rows)
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
func (m *MariaDB) buildCreateUserQuery(input types.UserDataStoreCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Insert(queriers.UsersTableName).
		Columns(
			queriers.UsersTableUsernameColumn,
			queriers.UsersTableHashedPasswordColumn,
			queriers.UsersTableSaltColumn,
			queriers.UsersTableTwoFactorColumn,
			queriers.UsersTableAccountStatusColumn,
			queriers.UsersTableIsAdminColumn,
			queriers.UsersTableAdminPermissionsColumn,
		).
		Values(
			input.Username,
			input.HashedPassword,
			input.Salt,
			input.TwoFactorSecret,
			types.UnverifiedStandingAccountStatus,
			false,
			0,
		).
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should be no way to update a user via this structure
	// such that they would have admin privileges.

	m.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user.
func (m *MariaDB) CreateUser(ctx context.Context, input types.UserDataStoreCreationInput) (*types.User, error) {
	x := &types.User{
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

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = m.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (m *MariaDB) buildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableUsernameColumn, input.Username).
		Set(queriers.UsersTableHashedPasswordColumn, input.HashedPassword).
		Set(queriers.UsersTableSaltColumn, input.Salt).
		Set(queriers.UsersTableTwoFactorColumn, input.TwoFactorSecret).
		Set(queriers.UsersTableTwoFactorVerifiedOnColumn, input.TwoFactorSecretVerifiedOn).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn: input.ID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// incomplete types at your peril.
func (m *MariaDB) UpdateUser(ctx context.Context, input *types.User) error {
	query, args := m.buildUpdateUserQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's password.
func (m *MariaDB) buildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableHashedPasswordColumn, newHash).
		Set(queriers.UsersTableRequiresPasswordChangeColumn, false).
		Set(queriers.UsersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
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

// buildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (m *MariaDB) buildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.UsersTableAccountStatusColumn, types.GoodStandingAccountStatus).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
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

// buildBanUserQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (m *MariaDB) buildBanUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableAccountStatusColumn, types.BannedStandingAccountStatus).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// BanUser bans a user.
func (m *MariaDB) BanUser(ctx context.Context, userID uint64) error {
	query, args := m.buildBanUserQuery(userID)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveUserQuery builds a SQL query that marks a user as archived.
func (m *MariaDB) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
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

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (m *MariaDB) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildCycleCookieSecretEvent(userID))
}

// LogSuccessfulLoginEvent saves a SuccessfulLoginEvent in the audit log table.
func (m *MariaDB) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildSuccessfulLoginEventEntry(userID))
}

// LogBannedUserLoginAttemptEvent saves a SuccessfulLoginEvent in the audit log table.
func (m *MariaDB) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildBannedUserLoginAttemptEventEntry(userID))
}

// LogUnsuccessfulLoginBadPasswordEvent saves a UnsuccessfulLoginBadPasswordEvent in the audit log table.
func (m *MariaDB) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(userID))
}

// LogUnsuccessfulLoginBad2FATokenEvent saves a UnsuccessfulLoginBad2FATokenEvent in the audit log table.
func (m *MariaDB) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(userID))
}

// LogLogoutEvent saves a LogoutEvent in the audit log table.
func (m *MariaDB) LogLogoutEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildLogoutEventEntry(userID))
}

// LogUserCreationEvent saves a UserCreationEvent in the audit log table.
func (m *MariaDB) LogUserCreationEvent(ctx context.Context, user *types.User) {
	m.createAuditLogEntry(ctx, audit.BuildUserCreationEventEntry(user))
}

// LogUserVerifyTwoFactorSecretEvent saves a UserVerifyTwoFactorSecretEvent in the audit log table.
func (m *MariaDB) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID))
}

// LogUserUpdateTwoFactorSecretEvent saves a UserUpdateTwoFactorSecretEvent in the audit log table.
func (m *MariaDB) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID))
}

// LogUserUpdatePasswordEvent saves a UserUpdatePasswordEvent in the audit log table.
func (m *MariaDB) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildUserUpdatePasswordEventEntry(userID))
}

// LogUserArchiveEvent saves a UserArchiveEvent in the audit log table.
func (m *MariaDB) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildUserArchiveEventEntry(userID))
}

// buildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (m *MariaDB) buildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					queriers.AuditLogEntriesTableName,
					queriers.AuditLogEntriesTableContextColumn,
					userID,
					audit.ActorAssignmentKey,
				),
			),
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					queriers.AuditLogEntriesTableName,
					queriers.AuditLogEntriesTableContextColumn,
					userID,
					audit.UserAssignmentKey,
				),
			),
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForUser fetches an audit log entry from the database.
func (m *MariaDB) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]types.AuditLogEntry, error) {
	query, args := m.buildGetAuditLogEntriesForUserQuery(userID)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := m.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

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
func (q *MariaDB) scanUser(scan database.Scanner, includeCount bool) (*types.User, uint64, error) {
	var (
		x     = &types.User{}
		perms uint32
		count uint64
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

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	x.AdminPermissions = bitmask.NewPermissionBitmask(perms)

	return x, count, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (q *MariaDB) scanUsers(rows database.ResultIterator, includeCount bool) ([]types.User, uint64, error) {
	var (
		list  []types.User
		count uint64
	)

	for rows.Next() {
		user, c, err := q.scanUser(rows, includeCount)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning user result: %w", err)
		}

		if count == 0 && includeCount {
			count = c
		}

		list = append(list, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := rows.Close(); err != nil {
		q.logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (q *MariaDB) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
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

	q.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user.
func (q *MariaDB) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := q.buildGetUserQuery(userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	u, _, err := q.scanUser(row, false)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret.
func (q *MariaDB) buildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.IDColumn):                            userID,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn):                    nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret.
func (q *MariaDB) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := q.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	u, _, err := q.scanUser(row, false)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (q *MariaDB) buildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
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

	q.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username.
func (q *MariaDB) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	query, args := q.buildGetUserByUsernameQuery(username)
	row := q.db.QueryRowContext(ctx, query, args...)

	u, _, err := q.scanUser(row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, nil
}

// buildSearchForUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (q *MariaDB) buildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
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

	q.logQueryBuildingError(err)

	return query, args
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (q *MariaDB) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]types.User, error) {
	query, args := q.buildSearchForUserByUsernameQuery(usernameQuery)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying database for users: %w", err)
	}

	u, _, err := q.scanUsers(rows, false)
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
func (q *MariaDB) buildGetAllUsersCountQuery() (query string) {
	var err error

	builder := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.UsersTableName)).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		})

	query, _, err = builder.ToSql()

	q.logQueryBuildingError(err)

	return query
}

// GetAllUsersCount fetches a count of users from the database.
func (q *MariaDB) GetAllUsersCount(ctx context.Context) (count uint64, err error) {
	query := q.buildGetAllUsersCountQuery()
	err = q.db.QueryRowContext(ctx, query).Scan(&count)

	return count, err
}

// buildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (q *MariaDB) buildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.
		Select(allCountQuery).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		})

	if filter != nil {
		countQueryBuilder = queriers.ApplyFilterToSubCountQueryBuilder(filter, countQueryBuilder, queriers.ItemsTableName)
	}

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(queriers.UsersTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.IDColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.UsersTableName)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (q *MariaDB) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	query, args := q.buildGetUsersQuery(filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	userList, count, err := q.scanUsers(rows, true)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	x := &types.UserList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Users: userList,
	}

	return x, nil
}

// buildCreateUserQuery returns a SQL query (and arguments) that would create a given User.
func (q *MariaDB) buildCreateUserQuery(input types.UserDataStoreCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
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
			types.UnverifiedAccountStatus,
			false,
			0,
		).
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should be no way to update a user via this structure
	// such that they would have admin privileges.

	q.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user.
func (q *MariaDB) CreateUser(ctx context.Context, input types.UserDataStoreCreationInput) (*types.User, error) {
	x := &types.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := q.buildCreateUserQuery(input)

	// create the user.
	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing user creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	q.logIDRetrievalError(err)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = q.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (q *MariaDB) buildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
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

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// incomplete types at your peril.
func (q *MariaDB) UpdateUser(ctx context.Context, input *types.User) error {
	query, args := q.buildUpdateUserQuery(input)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// buildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's password.
func (q *MariaDB) buildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableHashedPasswordColumn, newHash).
		Set(queriers.UsersTableRequiresPasswordChangeColumn, false).
		Set(queriers.UsersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateUserPassword updates a user's password.
func (q *MariaDB) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	query, args := q.buildUpdateUserPasswordQuery(userID, newHash)

	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// buildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (q *MariaDB) buildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.UsersTableAccountStatusColumn, types.GoodStandingAccountStatus).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (q *MariaDB) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	query, args := q.buildVerifyUserTwoFactorSecretQuery(userID)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveUserQuery builds a SQL query that marks a user as archived.
func (q *MariaDB) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser marks a user as archived.
func (q *MariaDB) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := q.buildArchiveUserQuery(userID)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (q *MariaDB) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildCycleCookieSecretEvent(userID))
}

// LogSuccessfulLoginEvent saves a SuccessfulLoginEvent in the audit log table.
func (q *MariaDB) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildSuccessfulLoginEventEntry(userID))
}

// LogBannedUserLoginAttemptEvent saves a SuccessfulLoginEvent in the audit log table.
func (q *MariaDB) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildBannedUserLoginAttemptEventEntry(userID))
}

// LogUnsuccessfulLoginBadPasswordEvent saves a UnsuccessfulLoginBadPasswordEvent in the audit log table.
func (q *MariaDB) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(userID))
}

// LogUnsuccessfulLoginBad2FATokenEvent saves a UnsuccessfulLoginBad2FATokenEvent in the audit log table.
func (q *MariaDB) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(userID))
}

// LogLogoutEvent saves a LogoutEvent in the audit log table.
func (q *MariaDB) LogLogoutEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildLogoutEventEntry(userID))
}

// LogUserCreationEvent saves a UserCreationEvent in the audit log table.
func (q *MariaDB) LogUserCreationEvent(ctx context.Context, user *types.User) {
	q.createAuditLogEntry(ctx, audit.BuildUserCreationEventEntry(user))
}

// LogUserVerifyTwoFactorSecretEvent saves a UserVerifyTwoFactorSecretEvent in the audit log table.
func (q *MariaDB) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID))
}

// LogUserUpdateTwoFactorSecretEvent saves a UserUpdateTwoFactorSecretEvent in the audit log table.
func (q *MariaDB) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID))
}

// LogUserUpdatePasswordEvent saves a UserUpdatePasswordEvent in the audit log table.
func (q *MariaDB) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildUserUpdatePasswordEventEntry(userID))
}

// LogUserArchiveEvent saves a UserArchiveEvent in the audit log table.
func (q *MariaDB) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildUserArchiveEventEntry(userID))
}

// buildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *MariaDB) buildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	builder := q.sqlBuilder.
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
	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForUser fetches an audit log entry from the database.
func (q *MariaDB) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]types.AuditLogEntry, error) {
	query, args := q.buildGetAuditLogEntriesForUserQuery(userID)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

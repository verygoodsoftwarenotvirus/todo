package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (p *Postgres) scanUser(scan database.Scanner) (*types.User, error) {
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
func (p *Postgres) scanUsers(rows database.ResultIterator) ([]types.User, error) {
	var (
		list []types.User
	)

	for rows.Next() {
		user, err := p.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning user result: %w", err)
		}

		list = append(list, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		p.logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (p *Postgres) buildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
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

	p.logQueryBuildingError(err)

	return query, args
}

// GetUser fetches a user.
func (p *Postgres) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := p.buildGetUserQuery(userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	u, err := p.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret.
func (p *Postgres) buildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.IDColumn):                            userID,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn):                    nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified two factor secret.
func (p *Postgres) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	query, args := p.buildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	u, err := p.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// buildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (p *Postgres) buildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
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

	p.logQueryBuildingError(err)

	return query, args
}

// GetUserByUsername fetches a user by their username.
func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	query, args := p.buildGetUserByUsernameQuery(username)
	row := p.db.QueryRowContext(ctx, query, args...)

	u, err := p.scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, nil
}

// buildSearchForUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (p *Postgres) buildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(queriers.UsersTableColumns...).
		From(queriers.UsersTableName).
		Where(squirrel.Expr(
			fmt.Sprintf("%s.%s ILIKE ?", queriers.UsersTableName, queriers.UsersTableUsernameColumn),
			fmt.Sprintf("%s%%", usernameQuery),
		)).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.UsersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (p *Postgres) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]types.User, error) {
	query, args := p.buildSearchForUserByUsernameQuery(usernameQuery)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying database for users: %w", err)
	}

	u, err := p.scanUsers(rows)
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
func (p *Postgres) buildGetAllUsersCountQuery() (query string) {
	var err error

	builder := p.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.UsersTableName)).
		From(queriers.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.UsersTableName, queriers.ArchivedOnColumn): nil,
		})

	query, _, err = builder.ToSql()

	p.logQueryBuildingError(err)

	return query
}

// GetAllUsersCount fetches a count of users from the database.
func (p *Postgres) GetAllUsersCount(ctx context.Context) (count uint64, err error) {
	query := p.buildGetAllUsersCountQuery()
	err = p.db.QueryRowContext(ctx, query).Scan(&count)
	return
}

// buildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (p *Postgres) buildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := p.sqlBuilder.
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
	p.logQueryBuildingError(err)
	return query, args
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (p *Postgres) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	query, args := p.buildGetUsersQuery(filter)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	userList, err := p.scanUsers(rows)
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
func (p *Postgres) buildCreateUserQuery(input types.UserDatabaseCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
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
		Suffix(fmt.Sprintf("RETURNING %s, %s", queriers.IDColumn, queriers.CreatedOnColumn)).
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should be no way to update a user via this structure
	// such that they would have admin privileges.

	p.logQueryBuildingError(err)

	return query, args
}

// CreateUser creates a user.
func (p *Postgres) CreateUser(ctx context.Context, input types.UserDatabaseCreationInput) (*types.User, error) {
	x := &types.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := p.buildCreateUserQuery(input)

	// create the user.
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		pge := &postgres.Error{}
		if errors.As(err, &pge) && pge.Code == postgresRowExistsErrorCode {
			return nil, dbclient.ErrUserExists
		}
		return nil, fmt.Errorf("error executing user creation query: %w", err)
	}

	return x, nil
}

// buildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (p *Postgres) buildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
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
		Suffix(fmt.Sprintf("RETURNING %s", queriers.LastUpdatedOnColumn)).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// UpdateUser receives a complete User struct and updates its place in the db.
// NOTE this function uses the ID provided in the input to make its query. Pass in
// incomplete types at your peril.
func (p *Postgres) UpdateUser(ctx context.Context, input *types.User) error {
	query, args := p.buildUpdateUserQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// buildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's password.
func (p *Postgres) buildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableHashedPasswordColumn, newHash).
		Set(queriers.UsersTableRequiresPasswordChangeColumn, false).
		Set(queriers.UsersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.LastUpdatedOnColumn)).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// UpdateUserPassword updates a user's password.
func (p *Postgres) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	query, args := p.buildUpdateUserPasswordQuery(userID, newHash)

	_, err := p.db.ExecContext(ctx, query, args...)

	return err
}

// buildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (p *Postgres) buildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.UsersTableAccountStatusColumn, types.GoodStandingAccountStatus).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
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

// buildBanUserQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (p *Postgres) buildBanUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableAccountStatusColumn, types.BannedStandingAccountStatus).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// BanUser bans a user.
func (p *Postgres) BanUser(ctx context.Context, userID uint64) error {
	query, args := p.buildBanUserQuery(userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveUserQuery builds a SQL query that marks a user as archived.
func (p *Postgres) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.ArchivedOnColumn)).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// ArchiveUser marks a user as archived.
func (p *Postgres) ArchiveUser(ctx context.Context, userID uint64) error {
	query, args := p.buildArchiveUserQuery(userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (p *Postgres) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildCycleCookieSecretEvent(userID))
}

// LogSuccessfulLoginEvent saves a SuccessfulLoginEvent in the audit log table.
func (p *Postgres) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildSuccessfulLoginEventEntry(userID))
}

// LogBannedUserLoginAttemptEvent saves a SuccessfulLoginEvent in the audit log table.
func (p *Postgres) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildBannedUserLoginAttemptEventEntry(userID))
}

// LogUnsuccessfulLoginBadPasswordEvent saves a UnsuccessfulLoginBadPasswordEvent in the audit log table.
func (p *Postgres) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(userID))
}

// LogUnsuccessfulLoginBad2FATokenEvent saves a UnsuccessfulLoginBad2FATokenEvent in the audit log table.
func (p *Postgres) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(userID))
}

// LogLogoutEvent saves a LogoutEvent in the audit log table.
func (p *Postgres) LogLogoutEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildLogoutEventEntry(userID))
}

// LogUserCreationEvent saves a UserCreationEvent in the audit log table.
func (p *Postgres) LogUserCreationEvent(ctx context.Context, user *types.User) {
	p.createAuditLogEntry(ctx, audit.BuildUserCreationEventEntry(user))
}

// LogUserVerifyTwoFactorSecretEvent saves a UserVerifyTwoFactorSecretEvent in the audit log table.
func (p *Postgres) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID))
}

// LogUserUpdateTwoFactorSecretEvent saves a UserUpdateTwoFactorSecretEvent in the audit log table.
func (p *Postgres) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID))
}

// LogUserUpdatePasswordEvent saves a UserUpdatePasswordEvent in the audit log table.
func (p *Postgres) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUserUpdatePasswordEventEntry(userID))
}

// LogUserArchiveEvent saves a UserArchiveEvent in the audit log table.
func (p *Postgres) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUserArchiveEventEntry(userID))
}

// buildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching audit log entries
// associated with a given user.
func (p *Postgres) buildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	userIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.UserAssignmentKey)
	performedByIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.ActorAssignmentKey)
	builder := p.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Eq{userIDKey: userID},
			squirrel.Eq{performedByIDKey: userID},
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForUser fetches a audit log entries for a given user from the database.
func (p *Postgres) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]types.AuditLogEntry, error) {
	query, args := p.buildGetAuditLogEntriesForUserQuery(userID)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := p.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

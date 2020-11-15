package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
	postgres "github.com/lib/pq"
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
	usersTableAdminPermissionsColumn       = "admin_permissions"
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
		fmt.Sprintf("%s.%s", usersTableName, usersTableAdminPermissionsColumn),
		fmt.Sprintf("%s.%s", usersTableName, createdOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, lastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn),
	}
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
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, idColumn):                            userID,
			fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn):                    nil,
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
		Select(usersTableColumns...).
		From(usersTableName).
		Where(squirrel.Expr(
			fmt.Sprintf("%s.%s ILIKE ?", usersTableName, usersTableUsernameColumn),
			fmt.Sprintf("%%%s%%", usernameQuery),
		)).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", usersTableName, usersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// SearchForUserByUsername fetches a user by their username.
func (p *Postgres) SearchForUserByUsername(ctx context.Context, usernameQuery string) (*types.User, error) {
	query, args := p.buildSearchForUserByUsernameQuery(usernameQuery)
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

// buildGetAllUsersCountQuery returns a SQL query (and arguments) for retrieving the number of users who adhere
// to a given filter's criteria.
func (p *Postgres) buildGetAllUsersCountQuery() (query string) {
	var err error

	builder := p.sqlBuilder.
		Select(fmt.Sprintf(countQuery, usersTableName)).
		From(usersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", usersTableName, archivedOnColumn): nil,
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
		Suffix(fmt.Sprintf("RETURNING %s, %s", idColumn, createdOnColumn)).
		ToSql()

	// NOTE: we always default is_admin to false, on the assumption that
	// admins have DB access and will change that value via SQL query.
	// There should also be no way to update a user via this structure
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
		Suffix(fmt.Sprintf("RETURNING %s", lastUpdatedOnColumn)).
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
		Update(usersTableName).
		Set(usersTableHashedPasswordColumn, newHash).
		Set(usersTableRequiresPasswordChangeColumn, false).
		Set(usersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", lastUpdatedOnColumn)).
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
		Update(usersTableName).
		Set(usersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
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

// buildArchiveUserQuery builds a SQL query that marks a user as archived.
func (p *Postgres) buildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(usersTableName).
		Set(archivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn: userID,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", archivedOnColumn)).
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

	userIDKey := fmt.Sprintf("%s.%s->'%s'", auditLogEntriesTableName, auditLogEntriesTableContextColumn, audit.UserAssignmentKey)
	performedByIDKey := fmt.Sprintf("%s.%s->'%s'", auditLogEntriesTableName, auditLogEntriesTableContextColumn, audit.ActorAssignmentKey)
	builder := p.sqlBuilder.
		Select(auditLogEntriesTableColumns...).
		From(auditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Eq{userIDKey: userID},
			squirrel.Eq{performedByIDKey: userID},
		}).
		OrderBy(fmt.Sprintf("%s.%s", auditLogEntriesTableName, idColumn))

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

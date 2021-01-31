package sqlite

import (
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.UserSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (q *Sqlite) BuildGetUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn):         userID,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret.
func (q *Sqlite) BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn):                            userID,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn):                    nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (q *Sqlite) BuildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableUsernameColumn): username,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn):         nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildSearchForUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (q *Sqlite) BuildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Expr(
			fmt.Sprintf("%s.%s LIKE ?", querybuilding.UsersTableName, querybuilding.UsersTableUsernameColumn),
			fmt.Sprintf("%s%%", usernameQuery),
		)).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllUsersCountQuery returns a SQL query (and arguments) for retrieving the number of users who adhere
// to a given filter's criteria.
func (q *Sqlite) BuildGetAllUsersCountQuery() (query string) {
	var err error

	builder := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.UsersTableName)).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		})

	query, _, err = builder.ToSql()

	q.logQueryBuildingError(err)

	return query
}

// BuildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (q *Sqlite) BuildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.CreatedOnColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, builder, querybuilding.UsersTableName)
	}

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// BuildTestUserCreationQuery builds a query and arguments that creates a test user.
func (q *Sqlite) BuildTestUserCreationQuery(testUserConfig *types.TestUserCreationConfig) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Insert(querybuilding.UsersTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.UsersTableUsernameColumn,
			querybuilding.UsersTableHashedPasswordColumn,
			querybuilding.UsersTableSaltColumn,
			querybuilding.UsersTableTwoFactorColumn,
			querybuilding.UsersTableIsAdminColumn,
			querybuilding.UsersTableReputationColumn,
			querybuilding.UsersTableAdminPermissionsColumn,
			querybuilding.UsersTableTwoFactorVerifiedOnColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			testUserConfig.Username,
			testUserConfig.HashedPassword,
			[]byte("aaaaaaaaaaaaaaaa"),
			// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
			"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			testUserConfig.IsSiteAdmin,
			types.GoodStandingAccountStatus,
			math.MaxUint32,
			squirrel.Expr(currentUnixTimeQuery),
		).
		ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// BuildCreateUserQuery returns a SQL query (and arguments) that would create a given User.
func (q *Sqlite) BuildCreateUserQuery(input types.UserDataStoreCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(querybuilding.UsersTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.UsersTableUsernameColumn,
			querybuilding.UsersTableHashedPasswordColumn,
			querybuilding.UsersTableSaltColumn,
			querybuilding.UsersTableTwoFactorColumn,
			querybuilding.UsersTableReputationColumn,
			querybuilding.UsersTableIsAdminColumn,
			querybuilding.UsersTableAdminPermissionsColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
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

// BuildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (q *Sqlite) BuildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableUsernameColumn, input.Username).
		Set(querybuilding.UsersTableHashedPasswordColumn, input.HashedPassword).
		Set(querybuilding.UsersTableSaltColumn, input.Salt).
		Set(querybuilding.UsersTableTwoFactorColumn, input.TwoFactorSecret).
		Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, input.TwoFactorSecretVerifiedOn).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn: input.ID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildSetUserStatusQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (q *Sqlite) BuildSetUserStatusQuery(userID uint64, input types.UserReputationUpdateInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableReputationColumn, input.NewReputation).
		Set(querybuilding.UsersTableStatusExplanationColumn, input.Reason).
		Where(squirrel.Eq{querybuilding.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's password.
func (q *Sqlite) BuildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableHashedPasswordColumn, newHash).
		Set(querybuilding.UsersTableRequiresPasswordChangeColumn, false).
		Set(querybuilding.UsersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{querybuilding.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (q *Sqlite) BuildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.UsersTableReputationColumn, types.GoodStandingAccountStatus).
		Where(squirrel.Eq{querybuilding.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveUserQuery builds a SQL query that marks a user as archived.
func (q *Sqlite) BuildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{querybuilding.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Sqlite) BuildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	userIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.UserAssignmentKey,
	)
	performedByIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.ActorAssignmentKey,
	)
	builder := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Eq{userIDKey: userID},
			squirrel.Eq{performedByIDKey: userID},
		}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

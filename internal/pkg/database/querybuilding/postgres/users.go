package postgres

import (
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.UserSQLQueryBuilder = (*Postgres)(nil)

// BuildUserIsBannedQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (q *Postgres) BuildUserIsBannedQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn)).
		Prefix(querybuilding.ExistencePrefix).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn):         userID,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		Where(squirrel.Or{
			squirrel.Eq{fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableReputationColumn): types.BannedAccountStatus},
			squirrel.Eq{fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableReputationColumn): types.TerminatedAccountStatus},
		}).
		Suffix(querybuilding.ExistenceSuffix),
	)
}

// BuildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (q *Postgres) BuildGetUserQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn):         userID,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
		}),
	)
}

// BuildGetUserWithUnverifiedTwoFactorSecretQuery returns a SQL query (and argument) for retrieving a user
// by their database ID, who has an unverified two factor secret.
func (q *Postgres) BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn):                            userID,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn):                    nil,
		}),
	)
}

// BuildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (q *Postgres) BuildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableUsernameColumn): username,
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn):         nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
		}),
	)
}

// BuildSearchForUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (q *Postgres) BuildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.UsersTableColumns...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Expr(
			fmt.Sprintf("%s.%s ILIKE ?", querybuilding.UsersTableName, querybuilding.UsersTableUsernameColumn),
			fmt.Sprintf("%s%%", usernameQuery),
		)).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		Where(squirrel.NotEq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
		}),
	)
}

// BuildGetAllUsersCountQuery returns a SQL query (and arguments) for retrieving the number of users who adhere
// to a given filter's criteria.
func (q *Postgres) BuildGetAllUsersCountQuery() (query string) {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.UsersTableName)).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria. It is assumed that this is only accessible to site administrators.
func (q *Postgres) BuildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.PlaceholderFormat(squirrel.Question).
		Select(allCountQuery).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		})

	if filter != nil {
		countQueryBuilder = querybuilding.ApplyFilterToSubCountQueryBuilder(filter, querybuilding.ItemsTableName, countQueryBuilder)
	}

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(querybuilding.UsersTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.CreatedOnColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, querybuilding.UsersTableName, builder)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// BuildTestUserCreationQuery returns a SQL query (and arguments) that would create a test user.
func (q *Postgres) BuildTestUserCreationQuery(testUserConfig *types.TestUserCreationConfig) (query string, args []interface{}) {
	perms := 0
	if testUserConfig.IsServiceAdmin {
		perms = math.MaxUint32
	}

	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.UsersTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.UsersTableUsernameColumn,
			querybuilding.UsersTableHashedPasswordColumn,
			querybuilding.UsersTableSaltColumn,
			querybuilding.UsersTableTwoFactorSekretColumn,
			querybuilding.UsersTableReputationColumn,
			querybuilding.UsersTableAdminPermissionsColumn,
			querybuilding.UsersTableTwoFactorVerifiedOnColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			testUserConfig.Username,
			testUserConfig.HashedPassword,
			querybuilding.DefaultTestUserSalt,
			querybuilding.DefaultTestUserTwoFactorSecret,
			types.GoodStandingAccountStatus,
			perms,
			currentUnixTimeQuery,
		).
		Suffix(fmt.Sprintf("RETURNING %s", querybuilding.IDColumn)),
	)
}

// BuildCreateUserQuery returns a SQL query (and arguments) that would create a given User.
// NOTE: we always default is_admin to false, on the assumption that
// admins have DB access and will change that value via SQL query.
// There should be no way to update a user via this structure
// such that they would have admin privileges.
func (q *Postgres) BuildCreateUserQuery(input types.UserDataStoreCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.UsersTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.UsersTableUsernameColumn,
			querybuilding.UsersTableHashedPasswordColumn,
			querybuilding.UsersTableSaltColumn,
			querybuilding.UsersTableTwoFactorSekretColumn,
			querybuilding.UsersTableReputationColumn,
			querybuilding.UsersTableAdminPermissionsColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Username,
			input.HashedPassword,
			input.Salt,
			input.TwoFactorSecret,
			types.UnverifiedAccountStatus,
			0,
		).
		Suffix(fmt.Sprintf("RETURNING %s", querybuilding.IDColumn)),
	)
}

// BuildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (q *Postgres) BuildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableUsernameColumn, input.Username).
		Set(querybuilding.UsersTableHashedPasswordColumn, input.HashedPassword).
		Set(querybuilding.UsersTableSaltColumn, input.Salt).
		Set(querybuilding.UsersTableTwoFactorSekretColumn, input.TwoFactorSecret).
		Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, input.TwoFactorSecretVerifiedOn).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         input.ID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's authentication.
func (q *Postgres) BuildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableHashedPasswordColumn, newHash).
		Set(querybuilding.UsersTableRequiresPasswordChangeColumn, false).
		Set(querybuilding.UsersTablePasswordLastChangedOnColumn, currentUnixTimeQuery).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildUpdateUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (q *Postgres) BuildUpdateUserTwoFactorSecretQuery(userID uint64, newSecret string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, nil).
		Set(querybuilding.UsersTableTwoFactorSekretColumn, newSecret).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (q *Postgres) BuildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, currentUnixTimeQuery).
		Set(querybuilding.UsersTableReputationColumn, types.GoodStandingAccountStatus).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildSetUserStatusQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (q *Postgres) BuildSetUserStatusQuery(userID uint64, input types.UserReputationUpdateInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableReputationColumn, input.NewReputation).
		Set(querybuilding.UsersTableStatusExplanationColumn, input.Reason).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildArchiveUserQuery builds a SQL query that marks a user as archived.
func (q *Postgres) BuildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching audit log entries
// associated with a given user.
func (q *Postgres) BuildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	userIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.UserAssignmentKey)
	performedByIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.ActorAssignmentKey)
	builder := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Eq{userIDKey: userID},
			squirrel.Eq{performedByIDKey: userID},
		}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	return q.buildQuery(builder)
}

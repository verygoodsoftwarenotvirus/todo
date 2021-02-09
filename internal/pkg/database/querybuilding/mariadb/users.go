package mariadb

import (
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.UserSQLQueryBuilder = (*MariaDB)(nil)

// BuildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (q *MariaDB) BuildGetUserQuery(userID uint64) (query string, args []interface{}) {
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
func (q *MariaDB) BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
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
func (q *MariaDB) BuildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
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
func (q *MariaDB) BuildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
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
		}),
	)
}

// BuildGetAllUsersCountQuery returns a SQL query (and arguments) for retrieving the number of users who adhere
// to a given filter's criteria.
func (q *MariaDB) BuildGetAllUsersCountQuery() (query string) {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.UsersTableName)).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (q *MariaDB) BuildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.
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

// BuildTestUserCreationQuery returns a SQL query (and arguments) that would create a given test user.
func (q *MariaDB) BuildTestUserCreationQuery(testUserConfig *types.TestUserCreationConfig) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.UsersTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.UsersTableUsernameColumn,
			querybuilding.UsersTableHashedPasswordColumn,
			querybuilding.UsersTableSaltColumn,
			querybuilding.UsersTableTwoFactorSekretColumn,
			querybuilding.UsersTableIsAdminColumn,
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
			testUserConfig.IsSiteAdmin,
			types.GoodStandingAccountStatus,
			math.MaxUint32,
			squirrel.Expr(currentUnixTimeQuery),
		),
	)
}

// BuildCreateUserQuery returns a SQL query (and arguments) that would create a given User.
// NOTE: we always default is_admin to false, on the assumption that
// admins have DB access and will change that value via SQL query.
// There should be no way to update a user via this structure
// such that they would have admin privileges.
func (q *MariaDB) BuildCreateUserQuery(input types.UserDataStoreCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.UsersTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.UsersTableUsernameColumn,
			querybuilding.UsersTableHashedPasswordColumn,
			querybuilding.UsersTableSaltColumn,
			querybuilding.UsersTableTwoFactorSekretColumn,
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
		),
	)
}

// BuildSetUserStatusQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (q *MariaDB) BuildSetUserStatusQuery(userID uint64, input types.UserReputationUpdateInput) (query string, args []interface{}) {
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

// BuildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (q *MariaDB) BuildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableUsernameColumn, input.Username).
		Set(querybuilding.UsersTableHashedPasswordColumn, input.HashedPassword).
		Set(querybuilding.UsersTableSaltColumn, input.Salt).
		Set(querybuilding.UsersTableTwoFactorSekretColumn, input.TwoFactorSecret).
		Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, input.TwoFactorSecretVerifiedOn).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         input.ID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's authentication.
func (q *MariaDB) BuildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableHashedPasswordColumn, newHash).
		Set(querybuilding.UsersTableRequiresPasswordChangeColumn, false).
		Set(querybuilding.UsersTablePasswordLastChangedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildUpdateUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (q *MariaDB) BuildUpdateUserTwoFactorSecretQuery(userID uint64, newSecret string) (query string, args []interface{}) {
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
func (q *MariaDB) BuildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.UsersTableReputationColumn, types.GoodStandingAccountStatus).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildArchiveUserQuery builds a SQL query that marks a user as archived.
func (q *MariaDB) BuildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.UsersTableName).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:         userID,
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *MariaDB) BuildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	var err error

	builder := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Or{
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					querybuilding.AuditLogEntriesTableName,
					querybuilding.AuditLogEntriesTableContextColumn,
					userID,
					audit.ActorAssignmentKey,
				),
			),
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					querybuilding.AuditLogEntriesTableName,
					querybuilding.AuditLogEntriesTableContextColumn,
					userID,
					audit.UserAssignmentKey,
				),
			),
		}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

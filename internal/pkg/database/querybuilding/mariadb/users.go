package mariadb

import (
	"context"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ querybuilding.UserSQLQueryBuilder = (*MariaDB)(nil)

// BuildUserHasStatusQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (b *MariaDB) BuildUserHasStatusQuery(ctx context.Context, userID uint64, statuses ...string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	whereStatuses := squirrel.Or{}
	for _, status := range statuses {
		whereStatuses = append(whereStatuses, squirrel.Eq{fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableReputationColumn): status})
	}

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn)).
			Prefix(querybuilding.ExistencePrefix).
			From(querybuilding.UsersTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn):         userID,
				fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
			}).
			Where(whereStatuses).
			Suffix(querybuilding.ExistenceSuffix))
}

// BuildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (b *MariaDB) BuildGetUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.UsersTableColumns...).
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
func (b *MariaDB) BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.UsersTableColumns...).
			From(querybuilding.UsersTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.IDColumn):                            userID,
				fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.UsersTableTwoFactorVerifiedOnColumn): nil,
				fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn):                    nil,
			}),
	)
}

// BuildGetUserByUsernameQuery returns a SQL query (and argument) for retrieving a user by their username.
func (b *MariaDB) BuildGetUserByUsernameQuery(ctx context.Context, username string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, username)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.UsersTableColumns...).
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
func (b *MariaDB) BuildSearchForUserByUsernameQuery(ctx context.Context, usernameQuery string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachSearchQueryToSpan(span, usernameQuery)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.UsersTableColumns...).
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
func (b *MariaDB) BuildGetAllUsersCountQuery(ctx context.Context) (query string) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQueryOnly(span, b.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.UsersTableName)).
		From(querybuilding.UsersTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.UsersTableName, querybuilding.ArchivedOnColumn): nil,
		}))
}

// BuildGetUsersQuery returns a SQL query (and arguments) for retrieving a slice of users who adhere
// to a given filter's criteria.
func (b *MariaDB) BuildGetUsersQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	return b.buildListQuery(ctx, querybuilding.UsersTableName, "", querybuilding.UsersTableColumns, 0, false, filter)
}

// BuildTestUserCreationQuery returns a SQL query (and arguments) that would create a given test user.
func (b *MariaDB) BuildTestUserCreationQuery(ctx context.Context, testUserConfig *types.TestUserCreationConfig) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	perms := 0
	if testUserConfig.IsServiceAdmin {
		perms = math.MaxInt64
	}

	tracing.AttachUsernameToSpan(span, testUserConfig.Username)

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.UsersTableName).
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
				b.externalIDGenerator.NewExternalID(),
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				querybuilding.DefaultTestUserSalt,
				querybuilding.DefaultTestUserTwoFactorSecret,
				types.GoodStandingAccountStatus,
				perms,
				currentUnixTimeQuery,
			),
	)
}

// BuildCreateUserQuery returns a SQL query (and arguments) that would create a given Requester.
// NOTE: we always default is_admin to false, on the assumption that
// admins have DB access and will change that value via SQL query.
// There should be no way to update a user via this structure
// such that they would have admin privileges.
func (b *MariaDB) BuildCreateUserQuery(ctx context.Context, input *types.UserDataStoreCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, input.Username)

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.UsersTableName).
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
				b.externalIDGenerator.NewExternalID(),
				input.Username,
				input.HashedPassword,
				input.Salt,
				input.TwoFactorSecret,
				types.UnverifiedAccountStatus,
				0,
			),
	)
}

// BuildSetUserStatusQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (b *MariaDB) BuildSetUserStatusQuery(ctx context.Context, input types.UserReputationUpdateInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.TargetUserID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
			Set(querybuilding.UsersTableReputationColumn, input.NewReputation).
			Set(querybuilding.UsersTableStatusExplanationColumn, input.Reason).
			Where(squirrel.Eq{
				querybuilding.IDColumn:         input.TargetUserID,
				querybuilding.ArchivedOnColumn: nil,
			}),
	)
}

// BuildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (b *MariaDB) BuildUpdateUserQuery(ctx context.Context, input *types.User) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.ID)
	tracing.AttachUsernameToSpan(span, input.Username)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
			Set(querybuilding.UsersTableUsernameColumn, input.Username).
			Set(querybuilding.UsersTableHashedPasswordColumn, input.HashedPassword).
			Set(querybuilding.UsersTableSaltColumn, input.Salt).
			Set(querybuilding.UsersTableAvatarColumn, input.AvatarSrc).
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
func (b *MariaDB) BuildUpdateUserPasswordQuery(ctx context.Context, userID uint64, newHash string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
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
func (b *MariaDB) BuildUpdateUserTwoFactorSecretQuery(ctx context.Context, userID uint64, newSecret string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
			Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, nil).
			Set(querybuilding.UsersTableTwoFactorSekretColumn, newSecret).
			Where(squirrel.Eq{
				querybuilding.IDColumn:         userID,
				querybuilding.ArchivedOnColumn: nil,
			}),
	)
}

// BuildVerifyUserTwoFactorSecretQuery returns a SQL query (and arguments) that would update a given user's two factor secret.
func (b *MariaDB) BuildVerifyUserTwoFactorSecretQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
			Set(querybuilding.UsersTableTwoFactorVerifiedOnColumn, currentUnixTimeQuery).
			Set(querybuilding.UsersTableReputationColumn, types.GoodStandingAccountStatus).
			Where(squirrel.Eq{
				querybuilding.IDColumn:         userID,
				querybuilding.ArchivedOnColumn: nil,
			}),
	)
}

// BuildArchiveUserQuery builds a SQL query that marks a user as archived.
func (b *MariaDB) BuildArchiveUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
			Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:         userID,
				querybuilding.ArchivedOnColumn: nil,
			}),
	)
}

// BuildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (b *MariaDB) BuildGetAuditLogEntriesForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.
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
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

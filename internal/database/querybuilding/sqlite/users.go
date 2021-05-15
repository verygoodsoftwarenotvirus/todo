package sqlite

import (
	"context"
	"fmt"
	"math"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ querybuilding.UserSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildUserHasStatusQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (b *Sqlite) BuildUserHasStatusQuery(ctx context.Context, userID uint64, statuses ...string) (query string, args []interface{}) {
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
func (b *Sqlite) BuildGetUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildGetUserByUsernameQuery(ctx context.Context, username string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildSearchForUserByUsernameQuery(ctx context.Context, usernameQuery string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildGetAllUsersCountQuery(ctx context.Context) (query string) {
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
func (b *Sqlite) BuildGetUsersQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	return b.buildListQuery(ctx, querybuilding.UsersTableName, "", querybuilding.UsersTableColumns, 0, false, filter)
}

// BuildTestUserCreationQuery builds a query and arguments that creates a test user.
func (b *Sqlite) BuildTestUserCreationQuery(ctx context.Context, testUserConfig *types.TestUserCreationConfig) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	perms := 0
	if testUserConfig.IsServiceAdmin {
		perms = math.MaxInt64
	}

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.UsersTableName).
			Columns(
				querybuilding.ExternalIDColumn,
				querybuilding.UsersTableUsernameColumn,
				querybuilding.UsersTableHashedPasswordColumn,
				querybuilding.UsersTableTwoFactorSekretColumn,
				querybuilding.UsersTableReputationColumn,
				querybuilding.UsersTableAdminPermissionsColumn,
				querybuilding.UsersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				b.externalIDGenerator.NewExternalID(),
				testUserConfig.Username,
				testUserConfig.HashedPassword,
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
func (b *Sqlite) BuildCreateUserQuery(ctx context.Context, input *types.UserDataStoreCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.UsersTableName).
			Columns(
				querybuilding.ExternalIDColumn,
				querybuilding.UsersTableUsernameColumn,
				querybuilding.UsersTableHashedPasswordColumn,
				querybuilding.UsersTableTwoFactorSekretColumn,
				querybuilding.UsersTableReputationColumn,
				querybuilding.UsersTableAdminPermissionsColumn,
			).
			Values(
				b.externalIDGenerator.NewExternalID(),
				input.Username,
				input.HashedPassword,
				input.TwoFactorSecret,
				types.UnverifiedAccountStatus,
				0,
			),
	)
}

// BuildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (b *Sqlite) BuildUpdateUserQuery(ctx context.Context, input *types.User) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
			Set(querybuilding.UsersTableUsernameColumn, input.Username).
			Set(querybuilding.UsersTableHashedPasswordColumn, input.HashedPassword).
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

// BuildSetUserStatusQuery returns a SQL query (and arguments) that would change a user's account status.
func (b *Sqlite) BuildSetUserStatusQuery(ctx context.Context, input *types.UserReputationUpdateInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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

// BuildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's passwords.
func (b *Sqlite) BuildUpdateUserPasswordQuery(ctx context.Context, userID uint64, newHash string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildUpdateUserTwoFactorSecretQuery(ctx context.Context, userID uint64, newSecret string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildVerifyUserTwoFactorSecretQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildArchiveUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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
func (b *Sqlite) BuildGetAuditLogEntriesForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

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

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Or{
				squirrel.Eq{userIDKey: userID},
				squirrel.Eq{performedByIDKey: userID},
			}).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}
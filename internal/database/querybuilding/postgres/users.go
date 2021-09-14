package postgres

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ querybuilding.UserSQLQueryBuilder = (*Postgres)(nil)
)

// BuildUserHasStatusQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (b *Postgres) BuildUserHasStatusQuery(ctx context.Context, userID string, statuses ...string) (query string, args []interface{}) {
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
			Suffix(querybuilding.ExistenceSuffix),
	)
}

// BuildGetUserQuery returns a SQL query (and argument) for retrieving a user by their database ID.
func (b *Postgres) BuildGetUserQuery(ctx context.Context, userID string) (query string, args []interface{}) {
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
func (b *Postgres) BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx context.Context, userID string) (query string, args []interface{}) {
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
func (b *Postgres) BuildGetUserByUsernameQuery(ctx context.Context, username string) (query string, args []interface{}) {
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
func (b *Postgres) BuildSearchForUserByUsernameQuery(ctx context.Context, usernameQuery string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachSearchQueryToSpan(span, usernameQuery)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.UsersTableColumns...).
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
func (b *Postgres) BuildGetAllUsersCountQuery(ctx context.Context) (query string) {
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
// to a given filter's criteria. It is assumed that this is only accessible to site administrators.
func (b *Postgres) BuildGetUsersQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	return b.buildListQuery(
		ctx,
		querybuilding.UsersTableName,
		nil,
		nil,
		"",
		querybuilding.UsersTableColumns,
		"",
		false,
		filter,
	)
}

// BuildTestUserCreationQuery builds a query and arguments that creates a test user.
func (b *Postgres) BuildTestUserCreationQuery(ctx context.Context, testUserConfig *types.TestUserCreationConfig) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, testUserConfig.Username)

	serviceRole := authorization.ServiceUserRole
	if testUserConfig.IsServiceAdmin {
		serviceRole = authorization.ServiceAdminRole
	}

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.UsersTableName).
			Columns(
				querybuilding.IDColumn,
				querybuilding.UsersTableUsernameColumn,
				querybuilding.UsersTableHashedPasswordColumn,
				querybuilding.UsersTableTwoFactorSekretColumn,
				querybuilding.UsersTableReputationColumn,
				querybuilding.UsersTableServiceRolesColumn,
				querybuilding.UsersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				testUserConfig.ID,
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				querybuilding.DefaultTestUserTwoFactorSecret,
				types.GoodStandingAccountStatus,
				serviceRole.String(),
				currentUnixTimeQuery,
			),
	)
}

// BuildUserCreationQuery returns a SQL query (and arguments) that would create a given Requester.
// NOTE: we always default is_admin to false, on the assumption that
// admins have DB access and will change that value via SQL query.
// There should be no way to update a user via this structure
// such that they would have admin privileges.
func (b *Postgres) BuildUserCreationQuery(ctx context.Context, input *types.UserDataStoreCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, input.Username)

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.UsersTableName).
			Columns(
				querybuilding.IDColumn,
				querybuilding.UsersTableUsernameColumn,
				querybuilding.UsersTableHashedPasswordColumn,
				querybuilding.UsersTableTwoFactorSekretColumn,
				querybuilding.UsersTableReputationColumn,
				querybuilding.UsersTableServiceRolesColumn,
			).
			Values(
				input.ID,
				input.Username,
				input.HashedPassword,
				input.TwoFactorSecret,
				types.UnverifiedAccountStatus,
				authorization.ServiceUserRole.String(),
			),
	)
}

// BuildUpdateUserQuery returns a SQL query (and arguments) that would update the given user's row.
func (b *Postgres) BuildUpdateUserQuery(ctx context.Context, input *types.User) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.ID)
	tracing.AttachUsernameToSpan(span, input.Username)

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
func (b *Postgres) BuildSetUserStatusQuery(ctx context.Context, input *types.UserReputationUpdateInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.TargetUserID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.UsersTableName).
			Set(querybuilding.UsersTableReputationColumn, input.NewReputation).
			Set(querybuilding.UsersTableReputationExplanationColumn, input.Reason).
			Where(squirrel.Eq{
				querybuilding.IDColumn:         input.TargetUserID,
				querybuilding.ArchivedOnColumn: nil,
			}),
	)
}

// BuildUpdateUserPasswordQuery returns a SQL query (and arguments) that would update the given user's passwords.
func (b *Postgres) BuildUpdateUserPasswordQuery(ctx context.Context, userID, newHash string) (query string, args []interface{}) {
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
func (b *Postgres) BuildUpdateUserTwoFactorSecretQuery(ctx context.Context, userID, newSecret string) (query string, args []interface{}) {
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
func (b *Postgres) BuildVerifyUserTwoFactorSecretQuery(ctx context.Context, userID string) (query string, args []interface{}) {
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
func (b *Postgres) BuildArchiveUserQuery(ctx context.Context, userID string) (query string, args []interface{}) {
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

// BuildGetAuditLogEntriesForUserQuery constructs a SQL query for fetching audit log entries belong to a user with a given ID.
func (b *Postgres) BuildGetAuditLogEntriesForUserQuery(ctx context.Context, userID string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

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

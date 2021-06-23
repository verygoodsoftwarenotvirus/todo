package postgres

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ querybuilding.AccountUserMembershipSQLQueryBuilder = (*Postgres)(nil)
)

const (
	accountMemberRolesSeparator = ","
)

// BuildGetDefaultAccountIDForUserQuery does .
func (b *Postgres) BuildGetDefaultAccountIDForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.
			Select(fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.IDColumn)).
			From(querybuilding.AccountsTableName).
			Join(fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				querybuilding.AccountsUserMembershipTableName,
				querybuilding.AccountsUserMembershipTableName,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
				querybuilding.AccountsTableName,
				querybuilding.IDColumn,
			)).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn):      userID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableDefaultUserAccountColumn): true,
			}),
	)
}

// BuildGetAccountMembershipsForUserQuery does .
func (b *Postgres) BuildGetAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.
			Select(querybuilding.AccountsUserMembershipTableColumns...).
			Join(fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				querybuilding.AccountsTableName,
				querybuilding.AccountsTableName,
				querybuilding.IDColumn,
				querybuilding.AccountsUserMembershipTableName,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
			)).
			From(querybuilding.AccountsUserMembershipTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                               nil,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn): userID,
			}),
	)
}

// BuildUserIsMemberOfAccountQuery builds a query that checks to see if the user is the member of a given account.
func (b *Postgres) BuildUserIsMemberOfAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.
			Select(fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.IDColumn)).
			Prefix(querybuilding.ExistencePrefix).
			From(querybuilding.AccountsUserMembershipTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableAccountOwnershipColumn): accountID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn):    userID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                                  nil,
			}).
			Suffix(querybuilding.ExistenceSuffix),
	)
}

// BuildAddUserToAccountQuery builds a query that adds a user to an account.
func (b *Postgres) BuildAddUserToAccountQuery(ctx context.Context, input *types.AddUserToAccountInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.UserID)
	tracing.AttachAccountIDToSpan(span, input.AccountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.AccountsUserMembershipTableName).
			Columns(
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
				querybuilding.AccountsUserMembershipTableAccountRolesColumn,
			).
			Values(
				input.UserID,
				input.AccountID,
				strings.Join(input.AccountRoles, accountMemberRolesSeparator),
			),
	)
}

// BuildMarkAccountAsUserDefaultQuery builds a query that marks a user's account as their primary.
func (b *Postgres) BuildMarkAccountAsUserDefaultQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsUserMembershipTableName).
			Set(
				querybuilding.AccountsUserMembershipTableDefaultUserAccountColumn,
				squirrel.And{
					squirrel.Eq{querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID},
					squirrel.Eq{querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID},
				},
			).
			Where(squirrel.Eq{
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
				querybuilding.ArchivedOnColumn:                               nil,
			}),
	)
}

// BuildModifyUserPermissionsQuery builds.
func (b *Postgres) BuildModifyUserPermissionsQuery(ctx context.Context, userID, accountID uint64, newRoles []string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsUserMembershipTableName).
			Set(querybuilding.AccountsUserMembershipTableAccountRolesColumn, strings.Join(newRoles, accountMemberRolesSeparator)).
			Where(squirrel.Eq{
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn:    userID,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID,
			}),
	)
}

// BuildTransferAccountOwnershipQuery does .
func (b *Postgres) BuildTransferAccountOwnershipQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, newOwnerID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsTableName).
			Set(querybuilding.AccountsTableUserOwnershipColumn, newOwnerID).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                         accountID,
				querybuilding.AccountsTableUserOwnershipColumn: currentOwnerID,
				querybuilding.ArchivedOnColumn:                 nil,
			}),
	)
}

// BuildTransferAccountMembershipsQuery does .
func (b *Postgres) BuildTransferAccountMembershipsQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, newOwnerID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsUserMembershipTableName).
			Set(querybuilding.AccountsUserMembershipTableUserOwnershipColumn, newOwnerID).
			Where(squirrel.Eq{
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID,
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn:    currentOwnerID,
				querybuilding.ArchivedOnColumn:                                  nil,
			}),
	)
}

// BuildCreateMembershipForNewUserQuery builds a query that .
func (b *Postgres) BuildCreateMembershipForNewUserQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.AccountsUserMembershipTableName).
			Columns(
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
				querybuilding.AccountsUserMembershipTableDefaultUserAccountColumn,
				querybuilding.AccountsUserMembershipTableAccountRolesColumn,
			).
			Values(
				userID,
				accountID,
				true,
				strings.Join([]string{authorization.AccountAdminRole.String()}, accountMemberRolesSeparator),
			),
	)
}

// BuildRemoveUserFromAccountQuery builds a query that removes a user from an account.
func (b *Postgres) BuildRemoveUserFromAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Delete(querybuilding.AccountsUserMembershipTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableAccountOwnershipColumn): accountID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn):    userID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                                  nil,
			}),
	)
}

// BuildArchiveAccountMembershipsForUserQuery does .
func (b *Postgres) BuildArchiveAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsUserMembershipTableName).
			Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
				querybuilding.ArchivedOnColumn:                               nil,
			}),
	)
}

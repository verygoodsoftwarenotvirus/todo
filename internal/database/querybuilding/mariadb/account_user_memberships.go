package mariadb

import (
	"context"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ querybuilding.AccountUserMembershipSQLQueryBuilder = (*MariaDB)(nil)

// BuildGetDefaultAccountIDForUserQuery does .
func (b *MariaDB) BuildGetDefaultAccountIDForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(span, b.sqlBuilder.
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

// BuildMarkAccountAsUserDefaultQuery does .
func (b *MariaDB) BuildMarkAccountAsUserDefaultQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsUserMembershipTableName).
			Set(querybuilding.AccountsUserMembershipTableDefaultUserAccountColumn, squirrel.And{
				squirrel.Eq{querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID},
				squirrel.Eq{querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID},
			}).
			Where(squirrel.Eq{
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
				querybuilding.ArchivedOnColumn:                               nil,
			}),
	)
}

// BuildTransferAccountOwnershipQuery does .
func (b *MariaDB) BuildTransferAccountOwnershipQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
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
func (b *MariaDB) BuildTransferAccountMembershipsQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
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

// BuildModifyUserPermissionsQuery builds.
func (b *MariaDB) BuildModifyUserPermissionsQuery(ctx context.Context, userID, accountID uint64, perms permissions.ServiceUserPermission, newRole string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsUserMembershipTableName).
			Set(querybuilding.AccountsUserMembershipTableUserPermissionsColumn, perms).
			Set(querybuilding.AccountsUserMembershipTableAccountRoleColumn, newRole).
			Where(squirrel.Eq{
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn:    userID,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID,
			}),
	)
}

// BuildArchiveAccountMembershipsForUserQuery does .
func (b *MariaDB) BuildArchiveAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
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

// BuildGetAccountMembershipsForUserQuery does .
func (b *MariaDB) BuildGetAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	columns := append(querybuilding.AccountsUserMembershipTableColumns, fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.AccountsTableNameColumn))

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(columns...).
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

// BuildCreateMembershipForNewUserQuery builds a query that .
func (b *MariaDB) BuildCreateMembershipForNewUserQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
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
				querybuilding.AccountsUserMembershipTableAccountRoleColumn,
				querybuilding.AccountsUserMembershipTableUserPermissionsColumn,
			).
			Values(
				userID,
				accountID,
				true,
				authorization.AccountAdminRole.String(),
				math.MaxInt64,
			),
	)
}

// BuildUserIsMemberOfAccountQuery builds a query that checks to see if the user is the member of a given account.
func (b *MariaDB) BuildUserIsMemberOfAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.IDColumn)).
			Prefix(querybuilding.ExistencePrefix).
			From(querybuilding.AccountsUserMembershipTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn): accountID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn): userID,
				fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                               nil,
			}).
			Suffix(querybuilding.ExistenceSuffix))
}

// BuildAddUserToAccountQuery builds a query that adds a user to an account.
func (b *MariaDB) BuildAddUserToAccountQuery(ctx context.Context, input *types.AddUserToAccountInput) (query string, args []interface{}) {
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
				querybuilding.AccountsUserMembershipTableAccountRoleColumn,
				querybuilding.AccountsUserMembershipTableUserPermissionsColumn,
			).
			Values(
				input.UserID,
				input.AccountID,
				input.AccountRoles,
				input.UserAccountPermissions,
			),
	)
}

// BuildRemoveUserFromAccountQuery builds a query that removes a user from an account.
func (b *MariaDB) BuildRemoveUserFromAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
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

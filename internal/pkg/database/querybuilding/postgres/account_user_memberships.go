package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ querybuilding.AccountUserMembershipSQLQueryBuilder = (*Postgres)(nil)

// BuildGetDefaultAccountIDForUserQuery does .
func (b *Postgres) BuildGetDefaultAccountIDForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
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

// BuildGetAccountMembershipsForUserQuery does .
func (b *Postgres) BuildGetAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
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
				querybuilding.AccountsUserMembershipTableUserPermissionsColumn,
			).
			Values(
				userID,
				accountID,
				true,
				math.MaxUint32,
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
func (b *Postgres) BuildModifyUserPermissionsQuery(ctx context.Context, userID, accountID uint64, perms permissions.ServiceUserPermissions) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.AccountsUserMembershipTableName).
			Set(querybuilding.AccountsUserMembershipTableUserPermissionsColumn, perms).
			Where(squirrel.Eq{
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn:    userID,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID,
			}),
	)
}

// BuildTransferAccountOwnershipQuery builds.
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

// BuildUserIsMemberOfAccountQuery builds a query that checks to see if the user is the member of a given account.
func (b *Postgres) BuildUserIsMemberOfAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
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
func (b *Postgres) BuildAddUserToAccountQuery(ctx context.Context, accountID uint64, input *types.AddUserToAccountInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.UserID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.AccountsUserMembershipTableName).
			Columns(
				querybuilding.AccountsUserMembershipTableUserOwnershipColumn,
				querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
				querybuilding.AccountsUserMembershipTableUserPermissionsColumn,
			).
			Values(
				input.UserID,
				accountID,
				input.UserAccountPermissions,
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

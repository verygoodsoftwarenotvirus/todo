package querier

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountUserMembershipDataManager = (*SQLQuerier)(nil)
)

// scanAccountUserMembership takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountUserMembership struct.
func (q *SQLQuerier) scanAccountUserMembership(ctx context.Context, scan database.Scanner) (x *types.AccountUserMembership, accountName string, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.AccountUserMembership{}

	var rawPerms int64

	targetVars := []interface{}{
		&x.ID,
		&x.BelongsToUser,
		&x.BelongsToAccount,
		&rawPerms,
		&x.DefaultAccount,
		&x.CreatedOn,
		&x.ArchivedOn,
		&accountName,
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, "", observability.PrepareError(err, q.logger, span, "scanning account user memberships")
	}

	newPerms := permissions.NewServiceUserPermissions(uint32(rawPerms))
	x.UserAccountPermissions = newPerms

	return x, accountName, nil
}

// scanAccountUserMemberships takes some database rows and turns them into a slice of memberships.
func (q *SQLQuerier) scanAccountUserMemberships(ctx context.Context, rows database.ResultIterator) (defaultAccount uint64, membershipMap map[uint64]types.UserAccountMembershipInfo, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	membershipMap = map[uint64]types.UserAccountMembershipInfo{}
	logger := q.logger

	for rows.Next() {
		x, accountName, scanErr := q.scanAccountUserMembership(ctx, rows)
		if scanErr != nil {
			return 0, nil, scanErr
		}

		if x.DefaultAccount && defaultAccount == 0 {
			defaultAccount = x.BelongsToAccount
		}

		membershipMap[x.BelongsToAccount] = types.UserAccountMembershipInfo{
			AccountName: accountName,
			Permissions: x.UserAccountPermissions,
		}
	}

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return 0, nil, observability.PrepareError(err, logger, span, "handling rows")
	}

	return defaultAccount, membershipMap, nil
}

// BuildRequestContextForUser does .
func (q *SQLQuerier) BuildRequestContextForUser(ctx context.Context, userID uint64) (*types.RequestContext, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	user, err := q.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching user from database")
	}

	getAccountMembershipsQuery, getAccountMembershipsArgs := q.sqlQueryBuilder.BuildGetAccountMembershipsForUserQuery(ctx, userID)

	membershipRows, err := q.performReadQuery(ctx, "account memberships for user", getAccountMembershipsQuery, getAccountMembershipsArgs...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching user's memberships from database")
	}

	defaultAccountID, membershipMap, err := q.scanAccountUserMemberships(ctx, membershipRows)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user's memberships from database")
	}

	reqCtx := &types.RequestContext{
		Requester: types.RequesterInfo{
			ID:                      user.ID,
			Reputation:              user.Reputation,
			ReputationExplanation:   user.ReputationExplanation,
			ServiceAdminPermissions: user.ServiceAdminPermissions,
		},
		AccountPermissionsMap: membershipMap,
		ActiveAccountID:       defaultAccountID,
	}

	return reqCtx, nil
}

// GetDefaultAccountIDForUser does .
func (q *SQLQuerier) GetDefaultAccountIDForUser(ctx context.Context, userID uint64) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return 0, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	query, args := q.sqlQueryBuilder.BuildGetDefaultAccountIDForUserQuery(ctx, userID)

	var id uint64
	if err := q.getOneRow(ctx, q.db, "default account ID query", query, args...).Scan(&id); err != nil {
		return 0, observability.PrepareError(err, logger, span, "executing id query")
	}

	return id, nil
}

// MarkAccountAsUserDefault does a thing.
func (q *SQLQuerier) MarkAccountAsUserDefault(ctx context.Context, userID, accountID, changedByUser uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 || accountID == 0 {
		return ErrInvalidIDProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:      userID,
		keys.AccountIDKey:   accountID,
		keys.RequesterIDKey: changedByUser,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachRequestingUserIDToSpan(span, changedByUser)

	query, args := q.sqlQueryBuilder.BuildMarkAccountAsUserDefaultQuery(ctx, userID, accountID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the account.
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user default account assignment", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "assigning user default account")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserMarkedAccountAsDefaultEventEntry(userID, accountID, changedByUser)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "account not found for user")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("account marked as default")

	return nil
}

// UserIsMemberOfAccount does a thing.
func (q *SQLQuerier) UserIsMemberOfAccount(ctx context.Context, userID, accountID uint64) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 || accountID == 0 {
		return false, ErrInvalidIDProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	query, args := q.sqlQueryBuilder.BuildUserIsMemberOfAccountQuery(ctx, userID, accountID)

	result, err := q.performBooleanQuery(ctx, q.db, query, args)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "performing user membership check query")
	}

	return result, nil
}

// ModifyUserPermissions does a thing.
func (q *SQLQuerier) ModifyUserPermissions(ctx context.Context, accountID, userID, changedByUser uint64, input *types.ModifyUserPermissionsInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 || userID == 0 {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey:   accountID,
		keys.UserIDKey:      userID,
		keys.RequesterIDKey: changedByUser,
		"new_permissions":   input.UserAccountPermissions,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachRequestingUserIDToSpan(span, changedByUser)

	query, args := q.sqlQueryBuilder.BuildModifyUserPermissionsQuery(ctx, userID, accountID, input.UserAccountPermissions)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the membership.
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user account permissions modification", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "modifying user account permissions")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildModifyUserPermissionsEventEntry(userID, accountID, changedByUser, input.UserAccountPermissions, input.Reason)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user account membership permission modification audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("user permissions modified")

	return nil
}

// TransferAccountOwnership does a thing.
func (q *SQLQuerier) TransferAccountOwnership(ctx context.Context, accountID, transferredBy uint64, input *types.TransferAccountOwnershipInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey:   accountID,
		keys.RequesterIDKey: transferredBy,
		"current_owner":     input.CurrentOwner,
		"new_owner":         input.NewOwner,
	})

	tracing.AttachUserIDToSpan(span, input.NewOwner)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachRequestingUserIDToSpan(span, transferredBy)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	transferAccountOwnershipQuery, transferAccountOwnershipArgs := q.sqlQueryBuilder.BuildTransferAccountOwnershipQuery(ctx, input.CurrentOwner, input.NewOwner, accountID)

	// create the membership.
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user ownership transfer", transferAccountOwnershipQuery, transferAccountOwnershipArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "transferring account to new owner")
	}

	transferAccountMembershipQuery, transferAccountMembershipArgs := q.sqlQueryBuilder.BuildTransferAccountMembershipsQuery(ctx, input.CurrentOwner, input.NewOwner, accountID)

	// create the membership.
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user memberships transfer", transferAccountMembershipQuery, transferAccountMembershipArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "transferring account memberships")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildTransferAccountOwnershipEventEntry(accountID, transferredBy, input)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account ownership transfer audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("account transferred to new owner")

	return nil
}

// AddUserToAccount does a thing.
func (q *SQLQuerier) AddUserToAccount(ctx context.Context, input *types.AddUserToAccountInput, accountID, addedByUser uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.RequesterIDKey: addedByUser,
		keys.UserIDKey:      input.UserID,
		keys.AccountIDKey:   accountID,
		keys.PermissionsKey: input.UserAccountPermissions,
	})

	tracing.AttachUserIDToSpan(span, input.UserID)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachRequestingUserIDToSpan(span, addedByUser)

	query, args := q.sqlQueryBuilder.BuildAddUserToAccountQuery(ctx, accountID, input)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the membership.
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user account membership creation", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "creating user account membership")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(addedByUser, accountID, input)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user added to account audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("user added to account")

	return nil
}

// RemoveUserFromAccount does a thing.
func (q *SQLQuerier) RemoveUserFromAccount(ctx context.Context, userID, accountID, removedByUser uint64, reason string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 || accountID == 0 || removedByUser == 0 {
		return ErrInvalidIDProvided
	}

	if reason == "" {
		return ErrEmptyInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:      userID,
		keys.AccountIDKey:   accountID,
		keys.ReasonKey:      reason,
		keys.RequesterIDKey: removedByUser,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachRequestingUserIDToSpan(span, removedByUser)

	query, args := q.sqlQueryBuilder.BuildRemoveUserFromAccountQuery(ctx, userID, accountID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the membership.
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "user membership removal", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "removing user from account")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserRemovedFromAccountEventEntry(userID, accountID, removedByUser, reason)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing remove user from account audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("user removed from account")

	return nil
}

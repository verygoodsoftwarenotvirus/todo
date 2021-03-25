package querier

import (
	"context"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountUserMembershipDataManager = (*Client)(nil)

	errDefaultAccountNotFoundForUser = errors.New("default account not found for user")
)

// scanAccountUserMembership takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountUserMembership struct.
func (c *Client) scanAccountUserMembership(ctx context.Context, scan database.Scanner) (x *types.AccountUserMembership, err error) {
	_, span := c.tracer.StartSpan(ctx)
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
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, observability.PrepareError(err, c.logger, span, "scanning account user memberships")
	}

	newPerms := permissions.NewServiceUserPermissions(uint32(rawPerms))
	x.UserAccountPermissions = newPerms

	return x, nil
}

// scanAccountUserMemberships takes some database rows and turns them into a slice of memberships.
func (c *Client) scanAccountUserMemberships(ctx context.Context, rows database.ResultIterator) (memberships []*types.AccountUserMembership, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	for rows.Next() {
		x, scanErr := c.scanAccountUserMembership(ctx, rows)
		if scanErr != nil {
			return nil, scanErr
		}

		memberships = append(memberships, x)
	}

	if err = c.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, observability.PrepareError(err, logger, span, "handling rows")
	}

	return memberships, nil
}

// GetRequestContextForUser does a thing.
func (c *Client) GetRequestContextForUser(ctx context.Context, userID uint64) (reqCtx *types.RequestContext, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("GetRequestContextForUser called")

	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching user from database")
	}

	getAccountMembershipsQuery, getAccountMembershipsArgs := c.sqlQueryBuilder.BuildGetAccountMembershipsForUserQuery(userID)

	membershipRows, err := c.db.QueryContext(ctx, getAccountMembershipsQuery, getAccountMembershipsArgs...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching user's memberships from database")
	}

	memberships, err := c.scanAccountUserMemberships(ctx, membershipRows)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user's memberships from database")
	}

	activeAccountID := uint64(0)
	accountPermissionsMap := map[uint64]permissions.ServiceUserPermissions{}

	for _, membership := range memberships {
		if membership.BelongsToAccount != 0 {
			accountPermissionsMap[membership.BelongsToAccount] = membership.UserAccountPermissions

			if membership.DefaultAccount && activeAccountID == 0 {
				activeAccountID = membership.BelongsToAccount
			}
		}
	}

	if activeAccountID == 0 {
		return nil, observability.PrepareError(errDefaultAccountNotFoundForUser, logger, span, "default account not found for user #%d", userID)
	}

	reqCtx = &types.RequestContext{
		User: types.UserRequestContext{
			ID:                      user.ID,
			Status:                  user.Reputation,
			ServiceAdminPermissions: user.ServiceAdminPermissions,
		},
		AccountPermissionsMap: accountPermissionsMap,
		ActiveAccountID:       activeAccountID,
	}

	return reqCtx, nil
}

// GetMembershipsForUser does a thing.
func (c *Client) GetMembershipsForUser(ctx context.Context, userID uint64) (defaultAccount uint64, permissionsMap map[uint64]permissions.ServiceUserPermissions, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.UserIDKey, userID)
	getAccountMembershipsQuery, getAccountMembershipsArgs := c.sqlQueryBuilder.BuildGetAccountMembershipsForUserQuery(userID)

	membershipRows, err := c.db.QueryContext(ctx, getAccountMembershipsQuery, getAccountMembershipsArgs...)
	if err != nil {
		return 0, nil, observability.PrepareError(err, logger, span, "fetching memberships from database")
	}

	memberships, err := c.scanAccountUserMemberships(ctx, membershipRows)
	if err != nil {
		return 0, nil, observability.PrepareError(err, logger, span, "scanning memberships from database")
	}

	permissionsMap = map[uint64]permissions.ServiceUserPermissions{}

	for _, membership := range memberships {
		if membership.BelongsToAccount != 0 {
			permissionsMap[membership.BelongsToAccount] = membership.UserAccountPermissions

			if membership.DefaultAccount && defaultAccount == 0 {
				defaultAccount = membership.BelongsToAccount
			}
		}
	}

	if defaultAccount == 0 {
		return 0, nil, observability.PrepareError(errDefaultAccountNotFoundForUser, logger, span, "account not found for user #%d", userID)
	}

	return defaultAccount, permissionsMap, nil
}

// MarkAccountAsUserDefault does a thing.
func (c *Client) MarkAccountAsUserDefault(ctx context.Context, userID, accountID, changedByUser uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
		keys.RequesterKey: changedByUser,
	})

	logger.Debug("MarkAccountAsUserDefault called")

	query, args := c.sqlQueryBuilder.BuildMarkAccountAsUserDefaultQuery(userID, accountID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the account.
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user default account assignment", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "assigning user default account")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserMarkedAccountAsDefaultEventEntry(userID, accountID, changedByUser)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "account not found for user")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// UserIsMemberOfAccount does a thing.
func (c *Client) UserIsMemberOfAccount(ctx context.Context, userID, accountID uint64) (bool, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	logger.Debug("UserIsMemberOfAccount called")

	query, args := c.sqlQueryBuilder.BuildUserIsMemberOfAccountQuery(userID, accountID)

	return c.performBooleanQuery(ctx, c.db, query, args)
}

// ModifyUserPermissions does a thing.
func (c *Client) ModifyUserPermissions(ctx context.Context, accountID, userID, changedByUser uint64, input *types.ModifyUserPermissionsInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
		keys.RequesterKey: changedByUser,
		"new_permissions": input.UserAccountPermissions,
	})

	query, args := c.sqlQueryBuilder.BuildModifyUserPermissionsQuery(userID, accountID, input.UserAccountPermissions)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the membership.
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user account permissions modification", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "modifying user account permissions")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildModifyUserPermissionsEventEntry(userID, accountID, changedByUser, input.UserAccountPermissions, input.Reason)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user account membership permission modification audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Debug("user permissions modified")

	return nil
}

// TransferAccountOwnership does a thing.
func (c *Client) TransferAccountOwnership(ctx context.Context, accountID, transferredBy uint64, input *types.TransferAccountOwnershipInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.RequesterKey: transferredBy,
		"current_owner":   input.CurrentOwner,
		"new_owner":       input.NewOwner,
	})

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	transferAccountOwnershipQuery, transferAccountOwnershipArgs := c.sqlQueryBuilder.BuildTransferAccountOwnershipQuery(input.CurrentOwner, input.NewOwner, accountID)

	// create the membership.
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user ownership transfer", transferAccountOwnershipQuery, transferAccountOwnershipArgs); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "transferring account to new owner")
	}

	transferAccountMembershipQuery, transferAccountMembershipArgs := c.sqlQueryBuilder.BuildTransferAccountMembershipsQuery(input.CurrentOwner, input.NewOwner, accountID)

	// create the membership.
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user memberships transfer", transferAccountMembershipQuery, transferAccountMembershipArgs); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "transferring account memberships")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildTransferAccountOwnershipEventEntry(accountID, transferredBy, input)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account ownership transfer audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Debug("TransferAccountOwnership called")

	return nil
}

// AddUserToAccount does a thing.
func (c *Client) AddUserToAccount(ctx context.Context, input *types.AddUserToAccountInput, accountID, addedByUser uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.RequesterKey:   addedByUser,
		keys.UserIDKey:      input.UserID,
		keys.AccountIDKey:   accountID,
		keys.PermissionsKey: input.UserAccountPermissions,
	})

	query, args := c.sqlQueryBuilder.BuildAddUserToAccountQuery(accountID, input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the membership.
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user account membership creation", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "creating user account membership")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(addedByUser, accountID, input)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing user added to account audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Debug("user added to account")

	return nil
}

// RemoveUserFromAccount does a thing.
func (c *Client) RemoveUserFromAccount(ctx context.Context, userID, accountID, removedByUser uint64, reason string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
		keys.ReasonKey:    reason,
		keys.RequesterKey: removedByUser,
	})

	logger.Debug("RemoveUserFromAccount called")

	query, args := c.sqlQueryBuilder.BuildRemoveUserFromAccountQuery(userID, accountID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the membership.
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user membership removal", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "removing user from account")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserRemovedFromAccountEventEntry(userID, accountID, removedByUser, reason)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing remove user from account audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

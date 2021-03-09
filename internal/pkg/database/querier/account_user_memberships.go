package querier

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountUserMembershipDataManager = (*Client)(nil)
)

// scanAccountUserMembership takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountUserMembership struct.
func (c *Client) scanAccountUserMembership(scan database.Scanner) (x *types.AccountUserMembership, err error) {
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

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, scanErr
	}

	newPerms := permissions.NewServiceUserPermissions(uint32(rawPerms))
	x.UserAccountPermissions = newPerms

	return x, nil
}

// scanAccountUserMemberships takes some database rows and turns them into a slice of memberships.
func (c *Client) scanAccountUserMemberships(rows database.ResultIterator) (memberships []*types.AccountUserMembership, err error) {
	for rows.Next() {
		x, scanErr := c.scanAccountUserMembership(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		memberships = append(memberships, x)
	}

	if handleErr := c.handleRows(rows); handleErr != nil {
		return nil, handleErr
	}

	return memberships, nil
}

// GetRequestContextForUser does a thing.
func (c *Client) GetRequestContextForUser(ctx context.Context, userID uint64) (reqCtx *types.RequestContext, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("GetRequestContextForUser called")

	user, fetchUserErr := c.GetUser(ctx, userID)
	if fetchUserErr != nil {
		return nil, fetchUserErr
	}

	reqCtx = &types.RequestContext{
		User: types.UserRequestContext{
			Username:                user.Username,
			ID:                      user.ID,
			Status:                  user.Reputation,
			ServiceAdminPermissions: user.ServiceAdminPermissions,
		},
	}

	getAccountMembershipsQuery, getAccountMembershipsArgs := c.sqlQueryBuilder.BuildGetAccountMembershipsForUserQuery(userID)

	membershipRows, getMembershipsErr := c.db.QueryContext(ctx, getAccountMembershipsQuery, getAccountMembershipsArgs...)
	if getMembershipsErr != nil {
		logger.Error(getMembershipsErr, "fetching memberships from database")
		return nil, getMembershipsErr
	}

	memberships, scanErr := c.scanAccountUserMemberships(membershipRows)
	if scanErr != nil {
		logger.Error(scanErr, "scanning memberships from database")
		return nil, scanErr
	}

	reqCtx.User.AccountPermissionsMap = map[uint64]permissions.ServiceUserPermissions{}

	for _, membership := range memberships {
		if membership.BelongsToAccount != 0 {
			reqCtx.User.AccountPermissionsMap[membership.BelongsToAccount] = membership.UserAccountPermissions

			if membership.DefaultAccount && reqCtx.User.ActiveAccountID == 0 {
				reqCtx.User.ActiveAccountID = membership.BelongsToAccount
			}
		} else {
			logger.WithValue("membership", membership).Info("WTF WTF WTF WTF WTF")
		}
	}

	if reqCtx.User.ActiveAccountID == 0 {
		return nil, fmt.Errorf("default account not found for user %d", userID)
	}

	return reqCtx, nil
}

// GetMembershipsForUser does a thing.
func (c *Client) GetMembershipsForUser(ctx context.Context, userID uint64) (defaultAccount uint64, permissionsMap map[uint64]permissions.ServiceUserPermissions, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: userID,
	})

	logger.Debug("GetMembershipsForUser called")

	getAccountMembershipsQuery, getAccountMembershipsArgs := c.sqlQueryBuilder.BuildGetAccountMembershipsForUserQuery(userID)

	membershipRows, getMembershipsErr := c.db.QueryContext(ctx, getAccountMembershipsQuery, getAccountMembershipsArgs...)
	if getMembershipsErr != nil {
		logger.Error(getMembershipsErr, "fetching memberships from database")
		return 0, nil, getMembershipsErr
	}

	memberships, scanErr := c.scanAccountUserMemberships(membershipRows)
	if scanErr != nil {
		logger.Error(scanErr, "scanning memberships from database")
		return 0, nil, scanErr
	}

	permissionsMap = map[uint64]permissions.ServiceUserPermissions{}

	for _, membership := range memberships {
		if membership.BelongsToAccount != 0 {
			permissionsMap[membership.BelongsToAccount] = membership.UserAccountPermissions

			if membership.DefaultAccount && defaultAccount == 0 {
				defaultAccount = membership.BelongsToAccount
			}
		} else {
			logger.WithValue("membership", membership).Info("WTF WTF WTF WTF WTF")
		}
	}

	if defaultAccount == 0 {
		return 0, nil, fmt.Errorf("default account not found for user %d", userID)
	}

	logger.WithValue("permissions_map", permissionsMap).Info("returning permission map")

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
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the account.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user default account assignment", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserMarkedAccountAsDefaultEventEntry(userID, accountID, changedByUser)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
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
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the membership.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user account permissions modification", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildModifyUserPermissionsEventEntry(userID, accountID, changedByUser, input.UserAccountPermissions, input.Reason)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
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
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	transferAccountOwnershipQuery, transferAccountOwnershipArgs := c.sqlQueryBuilder.BuildTransferAccountOwnershipQuery(input.CurrentOwner, input.NewOwner, accountID)

	// create the membership.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user ownership transfer", transferAccountOwnershipQuery, transferAccountOwnershipArgs); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	transferAccountMembershipQuery, transferAccountMembershipArgs := c.sqlQueryBuilder.BuildTransferAccountMembershipsQuery(input.CurrentOwner, input.NewOwner, accountID)

	// create the membership.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user memberships transfer", transferAccountMembershipQuery, transferAccountMembershipArgs); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildTransferAccountOwnershipEventEntry(accountID, transferredBy, input)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
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
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the membership.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user account membership creation", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(addedByUser, accountID, input)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
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
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the membership.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user membership removal", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserRemovedFromAccountEventEntry(userID, accountID, removedByUser, reason)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

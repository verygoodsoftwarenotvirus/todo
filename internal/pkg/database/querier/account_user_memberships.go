package querier

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountUserMembershipDataManager = (*Client)(nil)
)

// scanAccountUserMembership takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountUserMembership struct.
func (c *Client) scanAccountUserMembership(scan database.Scanner) (x *types.AccountUserMembership, err error) {
	x = &types.AccountUserMembership{}

	targetVars := []interface{}{
		&x.ID,
		&x.BelongsToAccount,
		&x.BelongsToUser,
		&x.UserPermissions,
		&x.DefaultAccount,
		&x.CreatedOn,
		&x.ArchivedOn,
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, scanErr
	}

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

// GetMembershipsForUser does a thing.
func (c *Client) GetMembershipsForUser(ctx context.Context, userID uint64) (defaultAccount uint64, permissionsMap map[uint64]bitmask.ServiceUserPermissions, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: userID,
	})

	logger.Debug("GetMembershipsForUser called")

	getAccountMembershipsQuery, getAccountMembershipsArgs := c.sqlQueryBuilder.BuildGetAccountMembershipsForUserQuery(userID)

	rows, getMembershipsErr := c.db.QueryContext(ctx, getAccountMembershipsQuery, getAccountMembershipsArgs...)
	if getMembershipsErr != nil {
		logger.Error(getMembershipsErr, "fetching memberships from database")
		return 0, nil, getMembershipsErr
	}

	memberships, scanErr := c.scanAccountUserMemberships(rows)
	if scanErr != nil {
		logger.Error(scanErr, "scanning memberships from database")
		return 0, nil, scanErr
	}

	permissionsMap = map[uint64]bitmask.ServiceUserPermissions{}

	for _, membership := range memberships {
		permissionsMap[membership.ID] = membership.UserPermissions

		if membership.DefaultAccount && defaultAccount == 0 {
			defaultAccount = membership.ID
		}
	}

	if defaultAccount == 0 {
		return 0, nil, fmt.Errorf("default account not found for user %d", userID)
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
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the item.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user default account assignment", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserMarkedAccountAsDefaultEventEntry(userID, accountID, changedByUser))

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
func (c *Client) ModifyUserPermissions(ctx context.Context, accountID, changedByUser uint64, input *types.ModifyUserPermissionsInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.RequesterKey: changedByUser,
		"input":           input,
	})

	logger.Debug("ModifyUserPermissions called")

	query, args := c.sqlQueryBuilder.BuildModifyUserPermissionsQuery(accountID, input.UserID, input.UserPermissions)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the item.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user membership removal", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildModifyUserPermissionsEventEntry(input.UserID, accountID, changedByUser, input.UserPermissions, input.Reason))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// TransferAccountOwnership does a thing.
func (c *Client) TransferAccountOwnership(ctx context.Context, accountID, transferredBy uint64, input *types.TransferAccountOwnershipInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.RequesterKey: transferredBy,
	})

	logger.Debug("TransferAccountOwnership called")

	query, args := c.sqlQueryBuilder.BuildTransferAccountOwnershipQuery(input.CurrentOwner, input.NewOwner, accountID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the item.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user membership removal", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildTransferAccountOwnershipEventEntry(input.CurrentOwner, input.NewOwner, transferredBy, accountID, input.Reason))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// AddUserToAccount does a thing.
func (c *Client) AddUserToAccount(ctx context.Context, userID, accountID, addedByUser uint64, reason string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
		keys.RequesterKey: addedByUser,
		keys.ReasonKey:    reason,
	})

	logger.Debug("AddUserToAccount called")

	query, args := c.sqlQueryBuilder.BuildAddUserToAccountQuery(userID, accountID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the item.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user account membership creation", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(userID, accountID, addedByUser, reason))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

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

	// create the item.
	if writeErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user membership removal", query, args); writeErr != nil {
		c.rollbackTransaction(tx)
		return writeErr
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserRemovedFromAccountEventEntry(userID, accountID, removedByUser, reason))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

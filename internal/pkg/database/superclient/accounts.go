package superclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountDataManager = (*Client)(nil)

// GetAccount fetches an account from the database.
func (c *Client) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		"account_id": accountID,
		"user_id":    userID,
	}).Debug("GetAccount called")

	return c.querier.GetAccount(ctx, accountID, userID)
}

// GetAllAccountsCount fetches the count of accounts from the database that meet a particular filter.
func (c *Client) GetAllAccountsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccountsCount called")

	return c.querier.GetAllAccountsCount(ctx)
}

// GetAllAccounts fetches a list of all accounts in the database.
func (c *Client) GetAllAccounts(ctx context.Context, results chan []types.Account, bucketSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccounts called")

	return c.querier.GetAllAccounts(ctx, results, bucketSize)
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (c *Client) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.AccountList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
	}

	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		"user_id": userID,
	}).Debug("GetAccounts called")

	return c.querier.GetAccounts(ctx, userID, filter)
}

// GetAccountsForAdmin fetches a list of accounts from the database that meet a particular filter for all users.
func (c *Client) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.AccountList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
	}

	c.logger.Debug("GetAccountsForAdmin called")

	accountList, err := c.querier.GetAccountsForAdmin(ctx, filter)

	return accountList, err
}

// CreateAccount creates an account in the database.
func (c *Client) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreateAccount called")

	return c.querier.CreateAccount(ctx, input)
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the
// provided input to have a valid ID.
func (c *Client) UpdateAccount(ctx context.Context, updated *types.Account) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, updated.ID)
	c.logger.WithValue(keys.AccountIDKey, updated.ID).Debug("UpdateAccount called")

	return c.querier.UpdateAccount(ctx, updated)
}

// ArchiveAccount archives an account from the database by its ID.
func (c *Client) ArchiveAccount(ctx context.Context, accountID, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	c.logger.WithValues(map[string]interface{}{
		"account_id": accountID,
		"user_id":    userID,
	}).Debug("ArchiveAccount called")

	return c.querier.ArchiveAccount(ctx, accountID, userID)
}

// LogAccountCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountCreationEvent(ctx context.Context, account *types.Account) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, account.BelongsToUser).Debug("LogAccountCreationEvent called")

	c.querier.LogAccountCreationEvent(ctx, account)
}

// LogAccountUpdateEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountUpdateEvent(ctx context.Context, userID, accountID uint64, changes []types.FieldChangeSummary) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogAccountUpdateEvent called")

	c.querier.LogAccountUpdateEvent(ctx, userID, accountID, changes)
}

// LogAccountArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountArchiveEvent(ctx context.Context, userID, accountID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogAccountArchiveEvent called")

	c.querier.LogAccountArchiveEvent(ctx, userID, accountID)
}

// GetAuditLogEntriesForAccount fetches a list of audit log entries from the database that relate to a given account.
func (c *Client) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForAccount called")

	return c.querier.GetAuditLogEntriesForAccount(ctx, accountID)
}

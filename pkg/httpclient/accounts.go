package httpclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// SwitchActiveAccount will, when provided the correct credentials, fetch a login cookie.
func (c *Client) SwitchActiveAccount(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	if c.authMethod == cookieAuthMethod {
		req, err := c.BuildSwitchActiveAccountRequest(ctx, accountID)
		if err != nil {
			return fmt.Errorf("building login request: %w", err)
		}

		res, err := c.authedClient.Do(req)
		if err != nil {
			return fmt.Errorf("encountered error executing login request: %w", err)
		}

		c.closeResponseBody(res)
	}

	c.accountID = accountID

	return nil
}

// GetAccount retrieves an account.
func (c *Client) GetAccount(ctx context.Context, accountID uint64) (account *types.Account, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.BuildGetAccountRequest(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &account); retrieveErr != nil {
		return nil, retrieveErr
	}

	return account, nil
}

// GetAccounts retrieves a list of accounts.
func (c *Client) GetAccounts(ctx context.Context, filter *types.QueryFilter) (accounts *types.AccountList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAccountsRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &accounts); retrieveErr != nil {
		return nil, retrieveErr
	}

	return accounts, nil
}

// CreateAccount creates an account.
func (c *Client) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (account *types.Account, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildCreateAccountRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &account)

	return account, err
}

// UpdateAccount updates an account.
func (c *Client) UpdateAccount(ctx context.Context, account *types.Account) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if account == nil {
		return ErrNilInputProvided
	}

	req, err := c.BuildUpdateAccountRequest(ctx, account)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &account)
}

// ArchiveAccount archives an account.
func (c *Client) ArchiveAccount(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.BuildArchiveAccountRequest(ctx, accountID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// AddUserToAccount adds a user from an account.
func (c *Client) AddUserToAccount(ctx context.Context, accountID uint64, input *types.AddUserToAccountInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildAddUserRequest(ctx, accountID, input)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// MarkAsDefault marks a given account as the default for a given user.
func (c *Client) MarkAsDefault(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.BuildMarkAsDefaultRequest(ctx, accountID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// RemoveUser removes a user from an account.
func (c *Client) RemoveUser(ctx context.Context, accountID, userID uint64, reason string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	if userID == 0 {
		return fmt.Errorf("userID: %w", ErrInvalidIDProvided)
	}

	req, err := c.BuildRemoveUserRequest(ctx, accountID, userID, reason)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// ModifyMemberPermissions modifies a given user's permissions for a given account.
func (c *Client) ModifyMemberPermissions(ctx context.Context, accountID, userID uint64, input *types.ModifyUserPermissionsInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	if userID == 0 {
		return fmt.Errorf("userID: %w", ErrInvalidIDProvided)
	}

	if input == nil {
		return ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildModifyMemberPermissionsRequest(ctx, accountID, userID, input)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// TransferAccountOwnership transfers ownership of an account to a given user.
func (c *Client) TransferAccountOwnership(ctx context.Context, accountID uint64, input *types.TransferAccountOwnershipInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	if input == nil {
		return ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildTransferAccountOwnershipRequest(ctx, accountID, input)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// GetAuditLogForAccount retrieves a list of audit log entries pertaining to an account.
func (c *Client) GetAuditLogForAccount(ctx context.Context, accountID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	req, err := c.BuildGetAuditLogForAccountRequest(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

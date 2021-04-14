package http

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// SwitchActiveAccount will switch the account on whose behalf requests are made.
func (c *Client) SwitchActiveAccount(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	if c.authMethod == cookieAuthMethod {
		input := &types.ChangeActiveAccountInput{}

		req, err := c.requestBuilder.BuildSwitchActiveAccountRequest(ctx, accountID)
		if err != nil {
			return observability.PrepareError(err, logger, span, "building account switch request")
		}

		if err = c.executeAndUnmarshal(ctx, req, c.authedClient, nil); err != nil {
			return observability.PrepareError(err, logger, span, "executing account switch request")
		}
	}

	c.accountID = accountID

	return nil
}

// GetAccount retrieves an account.
func (c *Client) GetAccount(ctx context.Context, accountID uint64) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	req, err := c.requestBuilder.BuildGetAccountRequest(ctx, accountID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building account retrieval request")
	}

	var account *types.Account
	if err = c.fetchAndUnmarshal(ctx, req, &account); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving account")
	}

	return account, nil
}

// GetAccounts retrieves a list of accounts.
func (c *Client) GetAccounts(ctx context.Context, filter *types.QueryFilter) (*types.AccountList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.loggerWithFilter(filter)

	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetAccountsRequest(ctx, filter)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building account list request")
	}

	var accounts *types.AccountList
	if err = c.fetchAndUnmarshal(ctx, req, &accounts); err != nil {
		return nil, observability.PrepareError(err, logger, span, "retrieving accounts")
	}

	return accounts, nil
}

// CreateAccount creates an account.
func (c *Client) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := c.logger.WithValue("account_name", input.Name)

	if err := input.Validate(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildCreateAccountRequest(ctx, input)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building account creation request")
	}

	var account *types.Account
	if err = c.fetchAndUnmarshal(ctx, req, &account); err != nil {
		return nil, observability.PrepareError(err, logger, span, "creating account")
	}

	return account, nil
}

// UpdateAccount updates an account.
func (c *Client) UpdateAccount(ctx context.Context, account *types.Account) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if account == nil {
		return ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, account.ID)
	tracing.AttachAccountIDToSpan(span, account.ID)

	req, err := c.requestBuilder.BuildUpdateAccountRequest(ctx, account)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building account update request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, &account); err != nil {
		return observability.PrepareError(err, logger, span, "updating account")
	}

	return nil
}

// ArchiveAccount archives an account.
func (c *Client) ArchiveAccount(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	req, err := c.requestBuilder.BuildArchiveAccountRequest(ctx, accountID)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building account archive request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "archiving account")
	}

	return nil
}

// AddUserToAccount adds a user to an account.
func (c *Client) AddUserToAccount(ctx context.Context, input *types.AddUserToAccountInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, input.AccountID).WithValue(keys.UserIDKey, input.UserID)
	tracing.AttachAccountIDToSpan(span, input.AccountID)
	tracing.AttachUserIDToSpan(span, input.UserID)

	if err := input.Validate(ctx); err != nil {
		return observability.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildAddUserRequest(ctx, input)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building add user to account request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "adding user to account")
	}

	return nil
}

// MarkAsDefault marks a given account as the default for a given user.
func (c *Client) MarkAsDefault(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	req, err := c.requestBuilder.BuildMarkAsDefaultRequest(ctx, accountID)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building mark account as default request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "marking account as default")
	}

	return nil
}

// RemoveUserFromAccount removes a user from an account.
func (c *Client) RemoveUserFromAccount(ctx context.Context, accountID, userID uint64, reason string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	if reason == "" {
		return ErrEmptyInputProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID).WithValue(keys.UserIDKey, userID)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, userID)

	req, err := c.requestBuilder.BuildRemoveUserRequest(ctx, accountID, userID, reason)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building remove user from account request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "removing user from account")
	}

	return nil
}

// ModifyMemberPermissions modifies a given user's permissions for a given account.
func (c *Client) ModifyMemberPermissions(ctx context.Context, accountID, userID uint64, input *types.ModifyUserPermissionsInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID).WithValue(keys.UserIDKey, userID)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, userID)

	if err := input.Validate(ctx); err != nil {
		return observability.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildModifyMemberPermissionsRequest(ctx, accountID, userID, input)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building modify account member permissions request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "modifying user account permissions")
	}

	return nil
}

// TransferAccountOwnership transfers ownership of an account to a given user.
func (c *Client) TransferAccountOwnership(ctx context.Context, accountID uint64, input *types.TransferAccountOwnershipInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID).
		WithValue("old_owner", input.CurrentOwner).
		WithValue("new_owner", input.NewOwner)

	tracing.AttachToSpan(span, "old_owner", input.CurrentOwner)
	tracing.AttachToSpan(span, "new_owner", input.NewOwner)

	if err := input.Validate(ctx); err != nil {
		return observability.PrepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildTransferAccountOwnershipRequest(ctx, accountID, input)
	if err != nil {
		return observability.PrepareError(err, logger, span, "building transfer account ownership request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return observability.PrepareError(err, logger, span, "transferring account to user")
	}

	return nil
}

// GetAuditLogForAccount retrieves a list of audit log entries pertaining to an account.
func (c *Client) GetAuditLogForAccount(ctx context.Context, accountID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	req, err := c.requestBuilder.BuildGetAuditLogForAccountRequest(ctx, accountID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building fetch audit log entries for account request")
	}

	var entries []*types.AuditLogEntry
	if err = c.fetchAndUnmarshal(ctx, req, &entries); err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching audit log entries for account")
	}

	return entries, nil
}

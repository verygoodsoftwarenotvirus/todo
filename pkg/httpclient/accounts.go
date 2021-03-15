package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	accountsBasePath = "accounts"
)

// BuildSwitchActiveAccountRequest builds an HTTP request for fetching an account.
func (c *Client) BuildSwitchActiveAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrZeroIDProvided
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "account", "select")

	input := &types.ChangeActiveAccountInput{
		AccountID: accountID,
	}

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// SwitchActiveAccount will, when provided the correct credentials, fetch a login cookie.
func (c *Client) SwitchActiveAccount(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrZeroIDProvided
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

// BuildGetAccountRequest builds an HTTP request for fetching an account.
func (c *Client) BuildGetAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrZeroIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		accountsBasePath,
		strconv.FormatUint(accountID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAccount retrieves an account.
func (c *Client) GetAccount(ctx context.Context, accountID uint64) (account *types.Account, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrZeroIDProvided
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

// BuildGetAccountsRequest builds an HTTP request for fetching accounts.
func (c *Client) BuildGetAccountsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), accountsBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
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

// BuildCreateAccountRequest builds an HTTP request for creating an account.
func (c *Client) BuildCreateAccountRequest(ctx context.Context, input *types.AccountCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
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

// BuildUpdateAccountRequest builds an HTTP request for updating an account.
func (c *Client) BuildUpdateAccountRequest(ctx context.Context, account *types.Account) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if account == nil {
		return nil, ErrNilInputProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		accountsBasePath,
		strconv.FormatUint(account.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPut, uri, account)
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

// BuildArchiveAccountRequest builds an HTTP request for updating an account.
func (c *Client) BuildArchiveAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrZeroIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		accountsBasePath,
		strconv.FormatUint(accountID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveAccount archives an account.
func (c *Client) ArchiveAccount(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrZeroIDProvided
	}

	req, err := c.BuildArchiveAccountRequest(ctx, accountID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildAddUserRequest builds a request that adds a user from an account.
func (c *Client) BuildAddUserRequest(ctx context.Context, accountID uint64, input *types.AddUserToAccountInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "member")
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
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

// BuildMarkAsDefaultRequest builds a request that marks a given account as the default for a given user.
func (c *Client) BuildMarkAsDefaultRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrZeroIDProvided
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "default")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
}

// MarkAsDefault marks a given account as the default for a given user.
func (c *Client) MarkAsDefault(ctx context.Context, accountID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return ErrZeroIDProvided
	}

	req, err := c.BuildMarkAsDefaultRequest(ctx, accountID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildRemoveUserRequest builds a request that removes a user from an account.
func (c *Client) BuildRemoveUserRequest(ctx context.Context, accountID, userID uint64, reason string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrZeroIDProvided)
	}

	if userID == 0 {
		return nil, fmt.Errorf("userID: %w", ErrZeroIDProvided)
	}

	u := c.buildRawURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "members", strconv.FormatUint(userID, 10))

	if reason != "" {
		u.Query().Set("reason", reason)
	}

	tracing.AttachURLToSpan(span, u)

	return http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
}

// RemoveUser removes a user from an account.
func (c *Client) RemoveUser(ctx context.Context, accountID, userID uint64, reason string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return fmt.Errorf("accountID: %w", ErrZeroIDProvided)
	}

	if userID == 0 {
		return fmt.Errorf("userID: %w", ErrZeroIDProvided)
	}

	req, err := c.BuildRemoveUserRequest(ctx, accountID, userID, reason)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildModifyMemberPermissionsRequest builds a request that modifies a given user's permissions for a given account.
func (c *Client) BuildModifyMemberPermissionsRequest(ctx context.Context, accountID, userID uint64, input *types.ModifyUserPermissionsInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrZeroIDProvided)
	}

	if userID == 0 {
		return nil, fmt.Errorf("userID: %w", ErrZeroIDProvided)
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "members", strconv.FormatUint(userID, 10), "permissions")
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPatch, uri, input)
}

// ModifyMemberPermissions modifies a given user's permissions for a given account.
func (c *Client) ModifyMemberPermissions(ctx context.Context, accountID, userID uint64, input *types.ModifyUserPermissionsInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return fmt.Errorf("accountID: %w", ErrZeroIDProvided)
	}

	if userID == 0 {
		return fmt.Errorf("userID: %w", ErrZeroIDProvided)
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

// BuildTransferAccountOwnershipRequest builds a request that transfers ownership of an account to a given user.
func (c *Client) BuildTransferAccountOwnershipRequest(ctx context.Context, accountID uint64, input *types.TransferAccountOwnershipInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrZeroIDProvided)
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "transfer")
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// TransferAccountOwnership transfers ownership of an account to a given user.
func (c *Client) TransferAccountOwnership(ctx context.Context, accountID uint64, input *types.TransferAccountOwnershipInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return fmt.Errorf("accountID: %w", ErrZeroIDProvided)
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

// BuildGetAuditLogForAccountRequest builds an HTTP request for fetching a list of audit log entries pertaining to an account.
func (c *Client) BuildGetAuditLogForAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrZeroIDProvided)
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogForAccount retrieves a list of audit log entries pertaining to an account.
func (c *Client) GetAuditLogForAccount(ctx context.Context, accountID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrZeroIDProvided)
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

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
		return nil, ErrInvalidIDProvided
	}

	uri := c.buildVersionlessURL(ctx, nil, usersBasePath, "account", "select")

	input := &types.ChangeActiveAccountInput{
		AccountID: accountID,
	}

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildGetAccountRequest builds an HTTP request for fetching an account.
func (c *Client) BuildGetAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
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

// BuildGetAccountsRequest builds an HTTP request for fetching accounts.
func (c *Client) BuildGetAccountsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), accountsBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
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

// BuildArchiveAccountRequest builds an HTTP request for updating an account.
func (c *Client) BuildArchiveAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
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

// BuildMarkAsDefaultRequest builds a request that marks a given account as the default for a given user.
func (c *Client) BuildMarkAsDefaultRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "default")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
}

// BuildRemoveUserRequest builds a request that removes a user from an account.
func (c *Client) BuildRemoveUserRequest(ctx context.Context, accountID, userID uint64, reason string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	if userID == 0 {
		return nil, fmt.Errorf("userID: %w", ErrInvalidIDProvided)
	}

	u := c.buildRawURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "members", strconv.FormatUint(userID, 10))

	if reason != "" {
		u.Query().Set("reason", reason)
	}

	tracing.AttachURLToSpan(span, u)

	return http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
}

// BuildModifyMemberPermissionsRequest builds a request that modifies a given user's permissions for a given account.
func (c *Client) BuildModifyMemberPermissionsRequest(ctx context.Context, accountID, userID uint64, input *types.ModifyUserPermissionsInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	if userID == 0 {
		return nil, fmt.Errorf("userID: %w", ErrInvalidIDProvided)
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

// BuildTransferAccountOwnershipRequest builds a request that transfers ownership of an account to a given user.
func (c *Client) BuildTransferAccountOwnershipRequest(ctx context.Context, accountID uint64, input *types.TransferAccountOwnershipInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
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

// BuildGetAuditLogForAccountRequest builds an HTTP request for fetching a list of audit log entries pertaining to an account.
func (c *Client) BuildGetAuditLogForAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	uri := c.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

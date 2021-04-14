package requests

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	accountsBasePath = "accounts"
)

// BuildSwitchActiveAccountRequest builds an HTTP request for switching active accounts.
func (b *Builder) BuildSwitchActiveAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)

	uri := b.buildVersionlessURL(ctx, nil, usersBasePath, "account", "select")

	input := &types.ChangeActiveAccountInput{
		AccountID: accountID,
	}

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildGetAccountRequest builds an HTTP request for fetching an account.
func (b *Builder) BuildGetAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)

	uri := b.BuildURL(
		ctx,
		nil,
		accountsBasePath,
		strconv.FormatUint(accountID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetAccountsRequest builds an HTTP request for fetching a list of accounts.
func (b *Builder) BuildGetAccountsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.BuildURL(ctx, filter.ToValues(), accountsBasePath)

	tracing.AttachRequestURIToSpan(span, uri)
	tracing.AttachQueryFilterToSpan(span, filter)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateAccountRequest builds an HTTP request for creating an account.
func (b *Builder) BuildCreateAccountRequest(ctx context.Context, input *types.AccountCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger.WithValue(keys.NameKey, input.Name)

	if err := input.Validate(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	uri := b.BuildURL(ctx, nil, accountsBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildUpdateAccountRequest builds an HTTP request for updating an account.
func (b *Builder) BuildUpdateAccountRequest(ctx context.Context, account *types.Account) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if account == nil {
		return nil, ErrNilInputProvided
	}

	uri := b.BuildURL(
		ctx,
		nil,
		accountsBasePath,
		strconv.FormatUint(account.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPut, uri, account)
}

// BuildArchiveAccountRequest builds an HTTP request for archiving an account.
func (b *Builder) BuildArchiveAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(
		ctx,
		nil,
		accountsBasePath,
		strconv.FormatUint(accountID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildAddUserRequest builds a request that adds a user to an account.
func (b *Builder) BuildAddUserRequest(ctx context.Context, input *types.AddUserToAccountInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger.WithValue(keys.UserIDKey, input.UserID)

	if err := input.Validate(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	uri := b.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(input.AccountID, 10), "member")
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildMarkAsDefaultRequest builds a request that marks a given account as the default for a given user.
func (b *Builder) BuildMarkAsDefaultRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "default")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
}

// BuildRemoveUserRequest builds a request that removes a user from an account.
func (b *Builder) BuildRemoveUserRequest(ctx context.Context, accountID, userID uint64, reason string) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	if userID == 0 {
		return nil, fmt.Errorf("userID: %w", ErrInvalidIDProvided)
	}

	u := b.buildAPIV1URL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "members", strconv.FormatUint(userID, 10))

	if reason != "" {
		q := u.Query()
		q.Set("reason", reason)
		u.RawQuery = q.Encode()
	}

	tracing.AttachURLToSpan(span, u)

	return http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
}

// BuildModifyMemberPermissionsRequest builds a request that modifies a given user's permissions for a given account.
func (b *Builder) BuildModifyMemberPermissionsRequest(ctx context.Context, accountID, userID uint64, input *types.ModifyUserPermissionsInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger.WithValue(keys.UserIDKey, userID).WithValue(keys.AccountIDKey, accountID)

	if err := input.Validate(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	uri := b.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "members", strconv.FormatUint(userID, 10), "permissions")
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPatch, uri, input)
}

// BuildTransferAccountOwnershipRequest builds a request that transfers ownership of an account to a given user.
func (b *Builder) BuildTransferAccountOwnershipRequest(ctx context.Context, accountID uint64, input *types.TransferAccountOwnershipInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, fmt.Errorf("accountID: %w", ErrInvalidIDProvided)
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger.WithValue(keys.AccountIDKey, accountID)

	if err := input.Validate(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	uri := b.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "transfer")
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildGetAuditLogForAccountRequest builds an HTTP request for fetching a list of audit log entries pertaining to an account.
func (b *Builder) BuildGetAuditLogForAccountRequest(ctx context.Context, accountID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(ctx, nil, accountsBasePath, strconv.FormatUint(accountID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

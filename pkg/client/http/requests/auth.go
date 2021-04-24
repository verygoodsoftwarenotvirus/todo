package requests

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	authBasePath = "auth"
)

// BuildUserStatusRequest builds an HTTP request that fetches a user's status.
func (b *Builder) BuildUserStatusRequest(ctx context.Context) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := b.logger
	uri := b.buildUnversionedURL(ctx, nil, authBasePath, "status")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildLoginRequest builds an HTTP request that fetches a login cookie.
func (b *Builder) BuildLoginRequest(ctx context.Context, input *types.UserLoginInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachUsernameToSpan(span, input.Username)

	// validating here requires settings knowledge, so we do not do it

	uri := b.buildUnversionedURL(ctx, nil, usersBasePath, "login")

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildLogoutRequest builds an HTTP request that clears the user's session.
func (b *Builder) BuildLogoutRequest(ctx context.Context) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.buildUnversionedURL(ctx, nil, usersBasePath, "logout")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, b.logger, span, "building user status request")
	}

	return req, nil
}

// BuildChangePasswordRequest builds a request to change a user's password.
func (b *Builder) BuildChangePasswordRequest(ctx context.Context, cookie *http.Cookie, input *types.PasswordUpdateInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// validating here requires settings knowledge, so we do not do it.

	uri := b.buildUnversionedURL(ctx, nil, usersBasePath, "password", "new")

	req, err := b.buildDataRequest(ctx, http.MethodPut, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildCycleTwoFactorSecretRequest builds a request to change a user's 2FA secret.
func (b *Builder) BuildCycleTwoFactorSecretRequest(ctx context.Context, cookie *http.Cookie, input *types.TOTPSecretRefreshInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
	}

	uri := b.buildUnversionedURL(ctx, nil, usersBasePath, "totp_secret", "new")

	req, err := b.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildVerifyTOTPSecretRequest builds a request to validate a 2FA secret.
func (b *Builder) BuildVerifyTOTPSecretRequest(ctx context.Context, userID uint64, token string) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := b.logger.WithValue(keys.UserIDKey, userID)

	if _, err := strconv.ParseUint(token, 10, 64); token == "" || err != nil {
		return nil, observability.PrepareError(err, logger, span, "invalid token provided")
	}

	uri := b.buildUnversionedURL(ctx, nil, usersBasePath, "totp_secret", "verify")

	return b.buildDataRequest(ctx, http.MethodPost, uri, &types.TOTPSecretVerificationInput{
		TOTPToken: token,
		UserID:    userID,
	})
}

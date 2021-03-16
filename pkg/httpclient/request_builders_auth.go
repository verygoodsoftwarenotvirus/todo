package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	authBasePath = "auth"
)

// BuildStatusRequest builds an HTTP request that fetches a user's status.
func (c *Client) BuildStatusRequest(ctx context.Context, cookie *http.Cookie) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	uri := c.buildVersionlessURL(ctx, nil, authBasePath, "status")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildLoginRequest builds an authenticating HTTP request.
func (c *Client) BuildLoginRequest(ctx context.Context, input *types.UserLoginInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// validating here requires settings knowledge, so we do not do it

	uri := c.buildVersionlessURL(ctx, nil, usersBasePath, "login")

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildLogoutRequest builds a de-authorizing HTTP request.
func (c *Client) BuildLogoutRequest(ctx context.Context) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(ctx, nil, usersBasePath, "logout")

	return http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
}

// BuildChangePasswordRequest builds a request to change a user's authentication.
func (c *Client) BuildChangePasswordRequest(ctx context.Context, cookie *http.Cookie, input *types.PasswordUpdateInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// validating here requires settings knowledge so we do not do it.

	uri := c.buildVersionlessURL(ctx, nil, usersBasePath, "password", "new")

	req, err := c.buildDataRequest(ctx, http.MethodPut, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildCycleTwoFactorSecretRequest builds a request to change a user's 2FA secret.
func (c *Client) BuildCycleTwoFactorSecretRequest(ctx context.Context, cookie *http.Cookie, input *types.TOTPSecretRefreshInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.buildVersionlessURL(ctx, nil, usersBasePath, "totp_secret", "new")

	req, err := c.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildVerifyTOTPSecretRequest builds a request to validate a TOTP secret.
func (c *Client) BuildVerifyTOTPSecretRequest(ctx context.Context, userID uint64, token string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	if _, err := strconv.ParseUint(token, 10, 64); token == "" || err != nil {
		return nil, fmt.Errorf("invalid token provided: %q", token)
	}

	uri := c.buildVersionlessURL(ctx, nil, usersBasePath, "totp_secret", "verify")

	return c.buildDataRequest(ctx, http.MethodPost, uri, &types.TOTPSecretVerificationInput{
		TOTPToken: token,
		UserID:    userID,
	})
}

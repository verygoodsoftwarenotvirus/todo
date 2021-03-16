package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// Status executes an HTTP request that fetches a user's status.
func (c *Client) Status(ctx context.Context, cookie *http.Cookie) (*types.UserStatusResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	req, err := c.BuildStatusRequest(ctx, cookie)
	if err != nil {
		return nil, fmt.Errorf("building login request: %w", err)
	}

	var output *types.UserStatusResponse

	if err := c.retrieve(ctx, req, &output); err != nil {
		return nil, err
	}

	return output, nil
}

// Login will, when provided the correct credentials, fetch a login cookie.
func (c *Client) Login(ctx context.Context, input *types.UserLoginInput) (*http.Cookie, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// validating here requires settings knowledge, so we do not do it

	req, err := c.BuildLoginRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building login request: %w", err)
	}

	res, err := c.plainClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("encountered error executing login request: %w", err)
	}

	c.closeResponseBody(res)

	cookies := res.Cookies()
	if len(cookies) > 0 {
		return cookies[0], nil
	}

	return nil, ErrNoCookiesReturned
}

// Logout logs a user out.
func (c *Client) Logout(ctx context.Context) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildLogoutRequest(ctx)
	if err != nil {
		return fmt.Errorf("building login request: %w", err)
	}

	res, err := c.authedClient.Do(req)
	if err != nil {
		return fmt.Errorf("encountered error executing login request: %w", err)
	}

	c.closeResponseBody(res)

	// should I be doing something here to undo the auth state in the client?

	return nil
}

// ChangePassword executes a request to change a user's authentication.
func (c *Client) ChangePassword(ctx context.Context, cookie *http.Cookie, input *types.PasswordUpdateInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return ErrCookieRequired
	}

	if input == nil {
		return ErrNilInputProvided
	}

	// validating here requires settings knowledge so we do not do it.

	req, err := c.BuildChangePasswordRequest(ctx, cookie, input)
	if err != nil {
		return fmt.Errorf("building authentication change request: %w", err)
	}

	res, err := c.executeRawRequest(ctx, c.plainClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	c.closeResponseBody(res)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("erroneous response code when changing authentication: %d", res.StatusCode)
	}

	return nil
}

// CycleTwoFactorSecret executes a request to change a user's 2FA secret.
func (c *Client) CycleTwoFactorSecret(ctx context.Context, cookie *http.Cookie, input *types.TOTPSecretRefreshInput) (*types.TOTPSecretRefreshResponse, error) {
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

	req, err := c.BuildCycleTwoFactorSecretRequest(ctx, cookie, input)
	if err != nil {
		return nil, fmt.Errorf("building authentication change request: %w", err)
	}

	var output *types.TOTPSecretRefreshResponse
	err = c.executeRequest(ctx, req, &output)

	return output, err
}

// VerifyTOTPSecret executes a request to verify a TOTP secret.
func (c *Client) VerifyTOTPSecret(ctx context.Context, userID uint64, token string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	if _, err := strconv.ParseUint(token, 10, 64); token == "" || err != nil {
		return fmt.Errorf("invalid token provided: %q", token)
	}

	req, err := c.BuildVerifyTOTPSecretRequest(ctx, userID, token)
	if err != nil {
		return fmt.Errorf("building TOTP validation request: %w", err)
	}

	res, err := c.executeRawRequest(ctx, c.plainClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	c.closeResponseBody(res)

	if res.StatusCode == http.StatusBadRequest {
		return ErrInvalidTOTPToken
	} else if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("erroneous response code when validating TOTP secret: %d", res.StatusCode)
	}

	return nil
}

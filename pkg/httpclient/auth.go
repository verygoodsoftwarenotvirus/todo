package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	authBasePath = "auth"
)

// BuildStatusRequest builds an HTTP request that fetches a user's status.
func (c *Client) BuildStatusRequest(ctx context.Context, cookie *http.Cookie) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(nil, authBasePath, "status")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// Status executes an HTTP request that fetches a user's status.
func (c *Client) Status(ctx context.Context, cookie *http.Cookie) (*types.UserStatusResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

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

// BuildLoginRequest builds an authenticating HTTP request.
func (c *Client) BuildLoginRequest(ctx context.Context, input *types.UserLoginInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "login")

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// Login will, when provided the correct credentials, fetch a login cookie.
func (c *Client) Login(ctx context.Context, input *types.UserLoginInput) (*http.Cookie, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

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

	return nil, errors.New("no cookies returned from request")
}

// BuildLogoutRequest builds a de-authorizing HTTP request.
func (c *Client) BuildLogoutRequest(ctx context.Context) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath, "logout")

	return http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
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

	return nil
}

// BuildChangePasswordRequest builds a request to change a user's authentication.
func (c *Client) BuildChangePasswordRequest(ctx context.Context, cookie *http.Cookie, input *types.PasswordUpdateInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "password", "new")

	req, err := c.buildDataRequest(ctx, http.MethodPut, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// ChangePassword executes a request to change a user's authentication.
func (c *Client) ChangePassword(ctx context.Context, cookie *http.Cookie, input *types.PasswordUpdateInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

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

// BuildCycleTwoFactorSecretRequest builds a request to change a user's 2FA secret.
func (c *Client) BuildCycleTwoFactorSecretRequest(ctx context.Context, cookie *http.Cookie, input *types.TOTPSecretRefreshInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "totp_secret", "new")

	req, err := c.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// CycleTwoFactorSecret executes a request to change a user's 2FA secret.
func (c *Client) CycleTwoFactorSecret(ctx context.Context, cookie *http.Cookie, input *types.TOTPSecretRefreshInput) (*types.TOTPSecretRefreshResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildCycleTwoFactorSecretRequest(ctx, cookie, input)
	if err != nil {
		return nil, fmt.Errorf("building authentication change request: %w", err)
	}

	var output *types.TOTPSecretRefreshResponse
	err = c.executeRequest(ctx, req, &output)

	return output, err
}

// BuildVerifyTOTPSecretRequest builds a request to validate a TOTP secret.
func (c *Client) BuildVerifyTOTPSecretRequest(ctx context.Context, userID uint64, token string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath, "totp_secret", "verify")

	return c.buildDataRequest(ctx, http.MethodPost, uri, &types.TOTPSecretVerificationInput{
		TOTPToken: token,
		UserID:    userID,
	})
}

// VerifyTOTPSecret executes a request to verify a TOTP secret.
func (c *Client) VerifyTOTPSecret(ctx context.Context, userID uint64, token string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

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

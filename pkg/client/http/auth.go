package http

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// UserStatus fetches a user's status.
func (c *Client) UserStatus(ctx context.Context, cookie *http.Cookie) (*types.UserStatusResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	logger := c.logger

	req, err := c.requestBuilder.BuildStatusRequest(ctx, cookie)
	if err != nil {
		return nil, prepareError(err, logger, span, "building user status request")
	}

	var output *types.UserStatusResponse

	if err = c.fetchAndUnmarshal(ctx, req, &output); err != nil {
		return nil, prepareError(err, logger, span, "retrieving plan")
	}

	return output, nil
}

// Login fetches a login cookie.
func (c *Client) Login(ctx context.Context, input *types.UserLoginInput) (*http.Cookie, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// validating here requires settings knowledge, so we do not do it
	logger := c.logger.WithValue(keys.UsernameKey, input.Username)

	req, err := c.requestBuilder.BuildLoginRequest(ctx, input)
	if err != nil {
		return nil, prepareError(err, logger, span, "building login request")
	}

	res, err := c.fetchResponseToRequest(ctx, c.unauthenticatedClient, req)
	if err != nil {
		return nil, prepareError(err, logger, span, "executing login request")
	}

	c.closeResponseBody(ctx, res)

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

	logger := c.logger

	req, err := c.requestBuilder.BuildLogoutRequest(ctx)
	if err != nil {
		return prepareError(err, logger, span, "building logout request")
	}

	res, err := c.fetchResponseToRequest(ctx, c.authedClient, req)
	if err != nil {
		return prepareError(err, logger, span, "executing logout request")
	}

	c.closeResponseBody(ctx, res)

	c.authedClient.Transport = newDefaultRoundTripper(c.authedClient.Timeout)

	return nil
}

// ChangePassword changes a user's password.
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

	logger := c.logger

	req, err := c.requestBuilder.BuildChangePasswordRequest(ctx, cookie, input)
	if err != nil {
		return prepareError(err, logger, span, "building change password request")
	}

	res, err := c.fetchResponseToRequest(ctx, c.unauthenticatedClient, req)
	if err != nil {
		return prepareError(err, logger, span, "changing password")
	}

	c.closeResponseBody(ctx, res)

	if res.StatusCode != http.StatusOK {
		return prepareError(err, logger, span, "invalid response code: %d", res.StatusCode)
	}

	return nil
}

// CycleTwoFactorSecret cycles a user's 2FA secret.
func (c *Client) CycleTwoFactorSecret(ctx context.Context, cookie *http.Cookie, input *types.TOTPSecretRefreshInput) (*types.TOTPSecretRefreshResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := c.logger

	if err := input.Validate(ctx); err != nil {
		return nil, prepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildCycleTwoFactorSecretRequest(ctx, cookie, input)
	if err != nil {
		return nil, prepareError(err, logger, span, "building cycle two factor secret request")
	}

	var output *types.TOTPSecretRefreshResponse
	if err = c.fetchAndUnmarshal(ctx, req, &output); err != nil {
		return nil, prepareError(err, logger, span, "cycling two factor secret")
	}

	return output, nil
}

// VerifyTOTPSecret verifies a 2FA secret.
func (c *Client) VerifyTOTPSecret(ctx context.Context, userID uint64, token string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	if _, err := strconv.ParseUint(token, 10, 64); token == "" || err != nil {
		return prepareError(err, logger, span, "invalid token provided: %q", token)
	}

	req, err := c.requestBuilder.BuildVerifyTOTPSecretRequest(ctx, userID, token)
	if err != nil {
		return prepareError(err, logger, span, "building verify two factor secret request")
	}

	res, err := c.fetchResponseToRequest(ctx, c.unauthenticatedClient, req)
	if err != nil {
		return prepareError(err, logger, span, "verifying two factor secret")
	}

	c.closeResponseBody(ctx, res)

	if res.StatusCode == http.StatusBadRequest {
		return ErrInvalidTOTPToken
	} else if res.StatusCode != http.StatusAccepted {
		return prepareError(err, logger, span, "erroneous response code when validating TOTP secret: %d", res.StatusCode)
	}

	return nil
}

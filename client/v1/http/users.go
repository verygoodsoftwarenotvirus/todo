package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const usersBasePath = "users"

// BuildGetUserRequest builds an HTTP request for fetching a user.
func (c *V1Client) BuildGetUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildGetUserRequest")
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetUser retrieves a user.
func (c *V1Client) GetUser(ctx context.Context, userID uint64) (user *models.User, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetUser")
	defer span.End()

	req, err := c.BuildGetUserRequest(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &user)
	return user, err
}

// BuildGetUsersRequest builds an HTTP request for fetching a user.
func (c *V1Client) BuildGetUsersRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildGetUsersRequest")
	defer span.End()

	uri := c.buildVersionlessURL(filter.ToValues(), usersBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetUsers retrieves a list of users.
func (c *V1Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetUsers")
	defer span.End()

	users := &models.UserList{}

	req, err := c.BuildGetUsersRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &users)
	return users, err
}

// BuildCreateUserRequest builds an HTTP request for creating a user.
func (c *V1Client) BuildCreateUserRequest(ctx context.Context, body *models.UserCreationInput) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildCreateUserRequest")
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath)

	return c.buildDataRequest(ctx, http.MethodPost, uri, body)
}

// CreateUser creates a new user.
func (c *V1Client) CreateUser(ctx context.Context, input *models.UserCreationInput) (*models.UserCreationResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "CreateUser")
	defer span.End()

	user := &models.UserCreationResponse{}

	req, err := c.BuildCreateUserRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeUnauthenticatedDataRequest(ctx, req, &user)
	return user, err
}

// BuildArchiveUserRequest builds an HTTP request for updating a user.
func (c *V1Client) BuildArchiveUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildArchiveUserRequest")
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveUser archives a user.
func (c *V1Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx, "ArchiveUser")
	defer span.End()

	req, err := c.BuildArchiveUserRequest(ctx, userID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildLoginRequest builds an authenticating HTTP request.
func (c *V1Client) BuildLoginRequest(ctx context.Context, input *models.UserLoginInput) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildLoginRequest")
	defer span.End()

	if input == nil {
		return nil, errors.New("nil input provided")
	}

	body, err := createBodyFromStruct(&input)
	if err != nil {
		return nil, fmt.Errorf("building request body: %w", err)
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "login")
	return c.buildDataRequest(ctx, http.MethodPost, uri, body)
}

// Login will, when provided the correct credentials, fetch a login cookie.
func (c *V1Client) Login(ctx context.Context, input *models.UserLoginInput) (*http.Cookie, error) {
	ctx, span := tracing.StartSpan(ctx, "Login")
	defer span.End()

	if input == nil {
		return nil, errors.New("nil input provided")
	}

	req, err := c.BuildLoginRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error building login request: %w", err)
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

// BuildVerifyTOTPSecretRequest builds a request to validate a TOTP secret.
func (c *V1Client) BuildVerifyTOTPSecretRequest(ctx context.Context, userID uint64, token string) (*http.Request, error) {
	ctx, span := tracing.StartSpan(ctx, "BuildVerifyTOTPSecretRequest")
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath, "totp_secret", "verify")

	return c.buildDataRequest(ctx, http.MethodPost, uri, &models.TOTPSecretVerificationInput{
		TOTPToken: token,
		UserID:    userID,
	})
}

// VerifyTOTPSecret executes a request to verify a TOTP secret.
func (c *V1Client) VerifyTOTPSecret(ctx context.Context, userID uint64, token string) error {
	ctx, span := tracing.StartSpan(ctx, "VerifyTOTPSecret")
	defer span.End()

	req, err := c.BuildVerifyTOTPSecretRequest(ctx, userID, token)
	if err != nil {
		return fmt.Errorf("error building TOTP validation request: %w", err)
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

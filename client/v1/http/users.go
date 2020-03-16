package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

const usersBasePath = "users"

// BuildGetUserRequest builds an HTTP request for fetching a user
func (c *V1Client) BuildGetUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	_, span := trace.StartSpan(ctx, "BuildGetUserRequest")
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetUser retrieves a user
func (c *V1Client) GetUser(ctx context.Context, userID uint64) (user *models.User, err error) {
	ctx, span := trace.StartSpan(ctx, "GetUser")
	defer span.End()

	req, err := c.BuildGetUserRequest(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &user)
	return user, err
}

// BuildGetUsersRequest builds an HTTP request for fetching a user
func (c *V1Client) BuildGetUsersRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	_, span := trace.StartSpan(ctx, "BuildGetUsersRequest")
	defer span.End()

	uri := c.buildVersionlessURL(filter.ToValues(), usersBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetUsers retrieves a list of users
func (c *V1Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	ctx, span := trace.StartSpan(ctx, "GetUsers")
	defer span.End()

	users := &models.UserList{}

	req, err := c.BuildGetUsersRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &users)
	return users, err
}

// BuildCreateUserRequest builds an HTTP request for creating a user
func (c *V1Client) BuildCreateUserRequest(ctx context.Context, body *models.UserInput) (*http.Request, error) {
	_, span := trace.StartSpan(ctx, "BuildCreateUserRequest")
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath)

	return c.buildDataRequest(ctx, http.MethodPost, uri, body)
}

// CreateUser creates a new user
func (c *V1Client) CreateUser(ctx context.Context, input *models.UserInput) (*models.UserCreationResponse, error) {
	ctx, span := trace.StartSpan(ctx, "CreateUser")
	defer span.End()

	user := &models.UserCreationResponse{}

	req, err := c.BuildCreateUserRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeUnathenticatedDataRequest(ctx, req, &user)
	return user, err
}

// BuildArchiveUserRequest builds an HTTP request for updating a user
func (c *V1Client) BuildArchiveUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	_, span := trace.StartSpan(ctx, "BuildArchiveUserRequest")
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequest(http.MethodDelete, uri, nil)
}

// ArchiveUser archives a user
func (c *V1Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "ArchiveUser")
	defer span.End()

	req, err := c.BuildArchiveUserRequest(ctx, userID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildLoginRequest builds an authenticating HTTP request
func (c *V1Client) BuildLoginRequest(ctx context.Context, username, password, totpToken string) (*http.Request, error) {
	_, span := trace.StartSpan(ctx, "BuildLoginRequest")
	defer span.End()

	body, err := createBodyFromStruct(&models.UserLoginInput{
		Username:  username,
		Password:  password,
		TOTPToken: totpToken,
	})

	if err != nil {
		return nil, fmt.Errorf("creating body from struct: %w", err)
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "login")
	return c.buildDataRequest(ctx, http.MethodPost, uri, body)
}

// Login will, when provided the correct credentials, fetch a login cookie
func (c *V1Client) Login(ctx context.Context, username, password, totpToken string) (*http.Cookie, error) {
	ctx, span := trace.StartSpan(ctx, "Login")
	defer span.End()

	req, err := c.BuildLoginRequest(ctx, username, password, totpToken)
	if err != nil {
		return nil, err
	}

	res, err := c.plainClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("encountered error executing login request: %w", err)
	}

	if c.Debug {
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			c.logger.Error(err, "dumping response")
		}
		c.logger.WithValue("response", string(b)).Debug("login response received")
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			c.logger.Error(err, "closing response body")
		}
	}()

	cookies := res.Cookies()
	if len(cookies) > 0 {
		return cookies[0], nil
	}

	return nil, errors.New("no cookies returned from request")
}

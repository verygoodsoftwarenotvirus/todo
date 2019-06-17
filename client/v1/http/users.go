package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const usersBasePath = "users"

func (c *V1Client) buildVersionlessURL(qp url.Values, parts ...string) string {
	tu := *c.URL

	u, err := url.Parse(path.Join(parts...))
	if err != nil {
		panic(fmt.Sprintf("user tried to build an invalid URL: %v", err))
	}

	if qp != nil {
		u.RawQuery = qp.Encode()
	}

	return tu.ResolveReference(u).String()
}

// BuildGetUserRequest builds an http Request for fetching a user
func (c *V1Client) BuildGetUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	uri := c.buildVersionlessURL(nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetUser gets a user
func (c *V1Client) GetUser(ctx context.Context, userID uint64) (user *models.User, err error) {
	req, err := c.BuildGetUserRequest(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &user)
	return user, err
}

// BuildGetUsersRequest builds an http Request for fetching a user
func (c *V1Client) BuildGetUsersRequest(ctx context.Context, filter *models.QueryFilter) (*http.Request, error) {
	uri := c.buildVersionlessURL(filter.ToValues(), usersBasePath)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// GetUsers gets a list of users
func (c *V1Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	users := &models.UserList{}

	req, err := c.BuildGetUsersRequest(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.retrieve(ctx, req, &users)
	return users, err
}

// BuildCreateUserRequest builds an http Request for creating a user
func (c *V1Client) BuildCreateUserRequest(ctx context.Context, body *models.UserInput) (*http.Request, error) {
	uri := c.buildVersionlessURL(nil, usersBasePath)

	return c.buildDataRequest(http.MethodPost, uri, body)
}

// CreateUser creates a new user
func (c *V1Client) CreateUser(ctx context.Context, input *models.UserInput) (*models.UserCreationResponse, error) {
	user := &models.UserCreationResponse{}

	req, err := c.BuildCreateUserRequest(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	err = c.makeUnauthedDataRequest(ctx, req, &user)
	return user, err
}

// BuildDeleteUserRequest builds an http Request for updating a user
func (c *V1Client) BuildDeleteUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	uri := c.buildVersionlessURL(nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequest(http.MethodDelete, uri, nil)
}

// DeleteUser deletes a user
func (c *V1Client) DeleteUser(ctx context.Context, userID uint64) error {
	req, err := c.BuildDeleteUserRequest(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	return c.makeRequest(ctx, req, nil)
}

// BuildLoginRequest builds an authenticating HTTP request
func (c *V1Client) BuildLoginRequest(username, password, totpToken string) (*http.Request, error) {
	body, err := createBodyFromStruct(&models.UserLoginInput{
		Username:  username,
		Password:  password,
		TOTPToken: totpToken,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating body from struct")
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "login")
	return c.buildDataRequest(http.MethodPost, uri, body)
}

// Login logs a user in
func (c *V1Client) Login(ctx context.Context, username, password, totpToken string) (*http.Cookie, error) {
	req, err := c.BuildLoginRequest(username, password, totpToken)
	if err != nil {
		return nil, err
	}

	res, err := c.plainClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "encountered error executing login request")
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

package client

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

const usersBasePath = "users"

func (c *V1Client) buildVersionlessURL(qp url.Values, parts ...string) string {
	tu := *c.URL

	u, _ := url.Parse(strings.Join(parts, "/"))
	if qp != nil {
		u.RawQuery = qp.Encode()
	}

	return tu.ResolveReference(u).String()
}

// GetUser gets a user
func (c *V1Client) GetUser(ctx context.Context, id string) (user *models.User, err error) {
	logger := c.logger.WithValue("user_id", id)
	logger.Debug("GetUser called")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUser")
	span.SetTag("itemID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(nil, usersBasePath, id)
	return user, c.get(ctx, uri, &user)
}

// GetUsers gets a list of users
func (c *V1Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	users := &models.UserList{}
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUsers")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(filter.ToValues(), usersBasePath)
	err := c.get(ctx, uri, &users)
	return users, err
}

// CreateUser creates a user
func (c *V1Client) CreateUser(ctx context.Context, input *models.UserInput) (*models.UserCreationResponse, error) {
	user := &models.UserCreationResponse{}
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateUser")
	span.SetTag("username", input.Username)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(nil, usersBasePath)
	err := c.post(ctx, uri, input, &user)
	return user, err
}

// CreateNewUser creates a new user
func (c *V1Client) CreateNewUser(ctx context.Context, input *models.UserInput) (*models.UserCreationResponse, error) {
	user := &models.UserCreationResponse{}
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateUser")
	span.SetTag("username", input.Username)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(nil, usersBasePath)
	err := c.postPlain(ctx, uri, input, &user)
	return user, err
}

// DeleteUser deletes a user
func (c *V1Client) DeleteUser(ctx context.Context, username string) error {
	logger := c.logger.WithValue("username", username)
	logger.Debug("")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "DeleteUser")
	span.SetTag("username", username)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(nil, usersBasePath, username)
	return c.delete(ctx, uri)
}

// Login logs a user in
func (c *V1Client) Login(ctx context.Context, username, password, TOTPToken string) (*http.Cookie, error) {
	logger := c.logger.WithValue("username", username)
	logger.Debug("login called")

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "login")
	span.SetTag("username", username)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	if c.currentUserCookie != nil {
		logger.Debug("returning user cookie from cache")
		return c.currentUserCookie, nil
	}

	body, err := createBodyFromStruct(&models.UserLoginInput{
		Username:  username,
		Password:  password,
		TOTPToken: TOTPToken,
	})

	if err != nil {
		logger.Error(err, "")
		return nil, err
	}

	uri := c.buildVersionlessURL(nil, usersBasePath, "login")
	req, _ := c.buildDataRequest(http.MethodPost, uri, body)

	res, err := c.plainClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "encountered error executing request")
	}

	b, _ := httputil.DumpResponse(res, true)
	logger.WithValue("response", string(b)).Debug("login response received")

	cookies := res.Cookies()
	if len(cookies) > 0 {
		c.currentUserCookie = cookies[0]
		return c.currentUserCookie, nil
	}

	return nil, errors.New("no cookies returned from request")
}

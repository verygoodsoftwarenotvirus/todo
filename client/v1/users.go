package client

import (
	"context"
	"net/http"
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
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUser")
	span.SetTag("itemID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(nil, usersBasePath, id)
	return user, c.get(ctx, uri, &user)
}

// GetUsers gets a list of users
func (c *V1Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (users *models.UserList, err error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUsers")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(filter.ToValues(), usersBasePath)
	return users, c.get(ctx, uri, &users)
}

// CreateUser creates a user
func (c *V1Client) CreateUser(ctx context.Context, input *models.UserInput) (user *models.UserCreationResponse, err error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateUser")
	span.SetTag("username", input.Username)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(nil, usersBasePath)
	return user, c.post(ctx, uri, input, &user)
}

// DeleteUser deletes a user
func (c *V1Client) DeleteUser(ctx context.Context, username string) error {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "DeleteUser")
	span.SetTag("username", username)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.buildVersionlessURL(nil, usersBasePath, username)
	return c.delete(ctx, uri)
}

// Login logs a user in
func (c *V1Client) Login(ctx context.Context, username, password, totpToken string) (*http.Cookie, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "Login")
	span.SetTag("username", username)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	body, err := createBodyFromStruct(&models.UserLoginInput{
		Username:  username,
		Password:  password,
		TOTPToken: totpToken,
	})
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodPost, c.buildVersionlessURL(nil, usersBasePath, "login"), body)
	res, err := c.plainClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "encountered error executing request")
	}
	cookies := res.Cookies()
	if len(cookies) > 0 {
		return cookies[0], nil
	}

	return nil, errors.New("no cookies returned from request")
}

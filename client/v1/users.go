package client

import (
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

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
func (c *V1Client) GetUser(id string) (user *models.User, err error) {
	return user, c.get(c.buildVersionlessURL(nil, usersBasePath, id), &user)
}

// GetUsers gets a list of users
func (c *V1Client) GetUsers(filter *models.QueryFilter) (users *models.UserList, err error) {
	return users, c.get(c.buildVersionlessURL(filter.ToValues(), usersBasePath), &users)
}

// CreateUser creates a user
func (c *V1Client) CreateUser(input *models.UserInput) (user *models.UserCreationResponse, err error) {
	return user, c.post(c.buildVersionlessURL(nil, usersBasePath), input, &user)
}

// DeleteUser deletes a user
func (c *V1Client) DeleteUser(username string) error {
	return c.delete(c.buildVersionlessURL(nil, usersBasePath, username))
}

// Login logs a user in
func (c *V1Client) Login(username, password, totpToken string) (*http.Cookie, error) {
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

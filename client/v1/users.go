package client

import (
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const usersBasePath = "users"

func (c *V1Client) buildVersionlessURL(qp Valuer, parts ...string) string {
	tu := *c.URL

	u, _ := url.Parse(strings.Join(parts, "/"))
	if qp != nil {
		u.RawQuery = qp.ToValues().Encode()
	}

	return tu.ResolveReference(u).String()
}

func (c *V1Client) GetUser(id string) (user *models.User, err error) {
	return user, c.get(c.buildVersionlessURL(nil, usersBasePath, id), &user)
}

func (c *V1Client) GetUsers(filter *models.QueryFilter) (users *models.UserList, err error) {
	return users, c.get(c.buildVersionlessURL(filter, usersBasePath), &users)
}

func (c *V1Client) CreateUser(input *models.UserInput) (user *models.User, err error) {
	return user, c.post(c.buildVersionlessURL(nil, usersBasePath), input, &user)
}

func (c *V1Client) DeleteUser(username string) error {
	return c.delete(c.buildVersionlessURL(nil, usersBasePath, username))
}

package client

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const oauth2ClientsBasePath = "clients"

func (c *V1Client) GetOauth2Client(id string) (oauth2Client *models.Oauth2Client, err error) {
	return oauth2Client, c.get(c.BuildURL(nil, oauth2ClientsBasePath, id), &oauth2Client)
}

func (c *V1Client) GetOauth2Clients(filter *models.QueryFilter) (oauth2Clients *models.Oauth2ClientList, err error) {
	return oauth2Clients, c.get(c.BuildURL(filter, oauth2ClientsBasePath), &oauth2Clients)
}

func (c *V1Client) CreateOauth2Client(input *models.Oauth2ClientInput) (oauth2Client *models.Oauth2Client, err error) {
	return oauth2Client, c.post(c.BuildURL(nil, oauth2ClientsBasePath), input, &oauth2Client)
}

func (c *V1Client) UpdateOauth2Client(updated *models.Oauth2Client) (err error) {
	return c.put(c.BuildURL(nil, oauth2ClientsBasePath, updated.ID), updated, &updated)
}

func (c *V1Client) DeleteOauth2Client(id string) error {
	return c.delete(c.BuildURL(nil, oauth2ClientsBasePath, id))
}

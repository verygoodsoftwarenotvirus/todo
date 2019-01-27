package client

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const oauth2ClientsBasePath = "oauth2/clients"

// GetOauth2Client gets an OAuth2 client
func (c *V1Client) GetOauth2Client(id string) (oauth2Client *models.Oauth2Client, err error) {
	c.logger.Debugf("GetOauth2Client called on %s", id)
	return oauth2Client, c.get(c.BuildURL(nil, oauth2ClientsBasePath, id), &oauth2Client)
}

// GetOauth2Clients gets a list of OAuth2 clients
func (c *V1Client) GetOauth2Clients(filter *models.QueryFilter) (oauth2Clients *models.Oauth2ClientList, err error) {
	c.logger.Debugln("GetOauth2Clients called")
	return oauth2Clients, c.get(c.BuildURL(filter.ToValues(), oauth2ClientsBasePath), &oauth2Clients)
}

// CreateOauth2Client creates an OAuth2 client
func (c *V1Client) CreateOauth2Client(input *models.Oauth2ClientCreationInput) (oauth2Client *models.Oauth2Client, err error) {
	c.logger.Debugln("CreateOauth2Client called")
	return oauth2Client, c.post(c.BuildURL(nil, oauth2ClientsBasePath), input, &oauth2Client)
}

// UpdateOauth2Client updates an OAuth2 client
func (c *V1Client) UpdateOauth2Client(updated *models.Oauth2Client) (err error) {
	c.logger.Debugf("UpdateOauth2Client called on %s", updated.ID)
	return c.put(c.BuildURL(nil, oauth2ClientsBasePath, updated.ID), updated, &updated)
}

// DeleteOauth2Client deletes an OAuth2 client
func (c *V1Client) DeleteOauth2Client(id string) error {
	c.logger.Debugf("DeleteOauth2Client called on %s", id)
	return c.delete(c.BuildURL(nil, oauth2ClientsBasePath, id))
}

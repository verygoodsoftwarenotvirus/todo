package client

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const oauth2ClientsBasePath = "oauth2/clients"

// GetOauth2Client gets an OAuth2 client
func (c *V1Client) GetOauth2Client(ctx context.Context, id string) (oauth2Client *models.OAuth2Client, err error) {
	c.logger.Debugf("GetOauth2Client called on %s", id)

	span := c.tracer.StartSpan("GetOauth2Client")
	span.SetTag("OAuth2ClientID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath, id)
	return oauth2Client, c.get(ctx, uri, &oauth2Client)
}

// GetOauth2Clients gets a list of OAuth2 clients
func (c *V1Client) GetOauth2Clients(ctx context.Context, filter *models.QueryFilter) (oauth2Clients *models.Oauth2ClientList, err error) {
	c.logger.Debugln("GetOauth2Clients called")

	span := c.tracer.StartSpan("GetOauth2Clients")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(filter.ToValues(), oauth2ClientsBasePath)
	return oauth2Clients, c.get(ctx, uri, &oauth2Clients)
}

// CreateOauth2Client creates an OAuth2 client
func (c *V1Client) CreateOauth2Client(ctx context.Context, input *models.Oauth2ClientCreationInput) (oauth2Client *models.OAuth2Client, err error) {
	c.logger.Debugln("CreateOauth2Client called")

	span := c.tracer.StartSpan("CreateOauth2Client")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath)
	return oauth2Client, c.post(ctx, uri, input, &oauth2Client)
}

// UpdateOauth2Client updates an OAuth2 client
func (c *V1Client) UpdateOauth2Client(ctx context.Context, updated *models.OAuth2Client) (err error) {
	c.logger.Debugf("UpdateOauth2Client called on %s", updated.ID)

	span := c.tracer.StartSpan("UpdateOauth2Client")
	span.SetTag("OAuth2ClientID", updated.ID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath, updated.ID)
	return c.put(ctx, uri, updated, &updated)
}

// DeleteOauth2Client deletes an OAuth2 client
func (c *V1Client) DeleteOauth2Client(ctx context.Context, id string) error {
	c.logger.Debugf("DeleteOauth2Client called on %s", id)

	span := c.tracer.StartSpan("DeleteOauth2Client")
	span.SetTag("OAuth2ClientID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath, id)
	return c.delete(ctx, uri)
}

package client

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const oauth2ClientsBasePath = "oauth2/clients"

// GetOAuth2Client gets an OAuth2 client
func (c *V1Client) GetOAuth2Client(ctx context.Context, id string) (oauth2Client *models.OAuth2Client, err error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetOAuth2Client")
	span.SetTag("OAuth2ClientID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath, id)
	return oauth2Client, c.get(ctx, uri, &oauth2Client)
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (c *V1Client) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter) (oauth2Clients *models.OAuth2ClientList, err error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetOAuth2Clients")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(filter.ToValues(), oauth2ClientsBasePath)
	return oauth2Clients, c.get(ctx, uri, &oauth2Clients)
}

// CreateOAuth2Client creates an OAuth2 client
func (c *V1Client) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (oauth2Client *models.OAuth2Client, err error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateOAuth2Client")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath)
	return oauth2Client, c.post(ctx, uri, input, &oauth2Client)
}

// UpdateOAuth2Client updates an OAuth2 client
func (c *V1Client) UpdateOAuth2Client(ctx context.Context, updated *models.OAuth2Client) error {
	logger := c.logger.WithValues(map[string]interface{}{
		"id":        updated.ID,
		"client_id": updated.ClientID,
	})

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "UpdateOAuth2Client")
	span.SetTag("OAuth2ClientID", updated.ID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath, updated.ID)
	if err := c.put(ctx, uri, updated, &updated); err != nil {
		logger.Error(err, "error encountered updating OAuth2 client")
		return err
	}
	return nil
}

// DeleteOAuth2Client deletes an OAuth2 client
func (c *V1Client) DeleteOAuth2Client(ctx context.Context, id string) error {
	logger := c.logger.WithValue("oauth2client_id", id)

	span := tracing.FetchSpanFromContext(ctx, c.tracer, "DeleteOAuth2Client")
	span.SetTag("OAuth2ClientID", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	uri := c.BuildURL(nil, oauth2ClientsBasePath, id)
	if err := c.delete(ctx, uri); err != nil {
		logger.Error(err, "error encountered deleting OAuth2 client")
		return err
	}
	return nil
}

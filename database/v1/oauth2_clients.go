package database

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	scopesSeparator = `,`
)

var _ models.OAuth2ClientHandler = (*Client)(nil)

// GetOAuth2Client gets an OAuth2 client
func (c *Client) GetOAuth2Client(ctx context.Context, clientID string, userID uint64) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetOAuth2Client")
	defer span.Finish()

	logger := c.logger.WithValue("oauth2_client_id", clientID)
	logger.Debug("Postgres.GetOAuth2Client called")

	return c.database.GetOAuth2Client(ctx, clientID, userID)
}

// GetOAuth2ClientByClientID fetches any OAuth2 client by client ID, regardless of ownershic. This is used by
// authenticating middleware to fetch client information it needs to validate
func (c *Client) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetOAuth2Client")
	defer span.Finish()

	c.logger.WithValue("client_id", clientID).Debug("GetOAuth2Client called")

	return c.database.GetOAuth2ClientByClientID(ctx, clientID)
}

// GetOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (c *Client) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetOAuth2ClientCount")
	defer span.Finish()

	logger := c.logger.WithValue("filter", filter)
	logger.Debug("Postgres.GetOAuth2ClientCount called")

	return c.database.GetOAuth2ClientCount(ctx, filter, userID)
}

// GetAllOAuth2Clients returns all OAuth2 clients, irrespective of ownershic. It is called on startup to populate
// the OAuth2 Client handler
func (c *Client) GetAllOAuth2Clients(ctx context.Context) ([]models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetAllOAuth2Clients")
	defer span.Finish()
	c.logger.Debug("Postgres.GetAllOAuth2Clients called")

	return c.database.GetAllOAuth2Clients(ctx)
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (c *Client) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetOAuth2Clients")
	defer span.Finish()

	logger := c.logger.WithValue("filter", filter)
	logger.Debug("Postgres.GetOAuth2Clients called")

	if filter == nil {
		logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}

	filter.SetPage(filter.Page)

	return c.database.GetOAuth2Clients(ctx, filter, userID)
}

// CreateOAuth2Client creates an OAuth2 client
func (c *Client) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateOAuth2Client")
	defer span.Finish()

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":    input.ClientID,
		"redirect_uri": input.RedirectURI,
		"scopes":       input.Scopes,
		"belongs_to":   input.BelongsTo,
	})
	logger.Debug("Postgres.CreateOAuth2Client called.")

	return c.database.CreateOAuth2Client(ctx, input)
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (c *Client) UpdateOAuth2Client(ctx context.Context, updated *models.OAuth2Client) error {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "UpdateOAuth2Client")
	defer span.Finish()

	logger := c.logger.WithValues(map[string]interface{}{
		"redirect_uri": updated.RedirectURI,
		"scopes":       updated.Scopes,
		"belongs_to":   updated.BelongsTo,
	})
	logger.Debug("Postgres.UpdateOAuth2Client called.")

	return c.database.UpdateOAuth2Client(ctx, updated)
}

// DeleteOAuth2Client deletes an OAuth2 client
func (c *Client) DeleteOAuth2Client(ctx context.Context, clientID string, userID uint64) error {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "DeleteOAuth2Client")
	defer span.Finish()

	logger := c.logger.WithValues(map[string]interface{}{
		"oauth2_client_id": clientID,
		"user_id":          userID,
	})
	logger.Debug("Postgres.DeleteOAuth2Client called")

	return c.database.DeleteOAuth2Client(ctx, clientID, userID)
}

package dbclient

import (
	"context"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.OAuth2ClientDataManager = (*Client)(nil)

// GetOAuth2Client gets an OAuth2 client
func (c *Client) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2Client")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("client_id", strconv.FormatUint(clientID, 10)))
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValues(map[string]interface{}{
		"client_id": clientID,
		"user_id":   userID,
	}).Debug("looking for an oauth2 client by this ID, for this user")

	client, err := c.querier.GetOAuth2Client(ctx, clientID, userID)
	if err != nil {
		c.logger.WithValues(map[string]interface{}{
			"error":     err,
			"user_id":   userID,
			"client_id": clientID,
		}).Debug("error fetching oauth2 client from the querier")

		return nil, err
	}

	return client, nil
}

// GetOAuth2ClientByClientID fetches any OAuth2 client by client ID, regardless of ownershic. This is used by
// authenticating middleware to fetch client information it needs to validate
func (c *Client) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2ClientByClientID")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("client_id", clientID))

	c.logger.WithValue("oauth2_client_id", clientID).Debug("GetOAuth2ClientByClientID called")

	client, err := c.querier.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		c.logger.WithValues(map[string]interface{}{
			"error":            err,
			"oauth2_client_id": clientID,
		}).Debug("error fetching oauth2 client from the querier")

		return nil, err
	}

	return client, nil
}

// GetOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (c *Client) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2ClientCount")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter()
	}

	logger := c.logger.WithValues(map[string]interface{}{
		"user_id": userID,
		"filter":  filter,
	})
	logger.Debug("GetOAuth2ClientCount called")

	return c.querier.GetOAuth2ClientCount(ctx, filter, userID)
}

// GetAllOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (c *Client) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllOAuth2ClientCount")
	defer span.End()

	c.logger.Debug("GetAllOAuth2ClientCount called")

	return c.querier.GetAllOAuth2ClientCount(ctx)
}

// GetAllOAuth2Clients returns all OAuth2 clients, irrespective of ownershic. It is called on startup to populate
// the OAuth2 Client handler
func (c *Client) GetAllOAuth2Clients(ctx context.Context) ([]*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllOAuth2Clients")
	defer span.End()

	c.logger.Debug("GetAllOAuth2Clients called")

	return c.querier.GetAllOAuth2Clients(ctx)
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (c *Client) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2Clients")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	logger := c.logger.WithValues(map[string]interface{}{
		"user_id": userID,
		"filter":  filter,
	})
	logger.Debug("GetOAuth2Clients called")

	if filter == nil {
		logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter()
	}

	filter.SetPage(filter.Page)

	return c.querier.GetOAuth2Clients(ctx, filter, userID)
}

// CreateOAuth2Client creates an OAuth2 client
func (c *Client) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "CreateOAuth2Client")
	defer span.End()

	client, err := c.querier.CreateOAuth2Client(ctx, input)
	if err != nil {
		c.logger.
			WithValues(map[string]interface{}{
				"client_id":  input.ClientID,
				"belongs_to": input.BelongsTo,
			}).
			WithError(err).
			Debug("error writing oauth2 client to the querier")
		return nil, err
	}

	c.logger.WithValues(map[string]interface{}{
		"client_id":  client.ID,
		"belongs_to": client.BelongsTo,
	}).Debug("new oauth2 client created successfully")

	return client, nil
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (c *Client) UpdateOAuth2Client(ctx context.Context, updated *models.OAuth2Client) error {
	ctx, span := trace.StartSpan(ctx, "UpdateOAuth2Client")
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		"redirect_uri": updated.RedirectURI,
		"scopes":       updated.Scopes,
		"belongs_to":   updated.BelongsTo,
	})
	logger.Debug("UpdateOAuth2Client called.")

	return c.querier.UpdateOAuth2Client(ctx, updated)
}

// DeleteOAuth2Client deletes an OAuth2 client
func (c *Client) DeleteOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "DeleteOAuth2Client")
	defer span.End()

	err := c.querier.DeleteOAuth2Client(ctx, clientID, userID)
	if err != nil {
		c.logger.
			WithError(err).
			Debug("error deleting oauth2 client to the querier")
		return err
	}

	c.logger.WithValues(map[string]interface{}{
		"client_id":  clientID,
		"belongs_to": userID,
	}).Debug("removed oauth2 client successfully")

	return nil
}

package dbclient

import (
	"context"
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.OAuth2ClientDataManager = (*Client)(nil)

// attachOAuth2ClientDatabaseIDToSpan is a consistent way to attach an oauth2 client's ID to a span
func attachOAuth2ClientDatabaseIDToSpan(span *trace.Span, oauth2ClientID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("oauth2client_id", strconv.FormatUint(oauth2ClientID, 10)))
	}
}

// attachOAuth2ClientIDToSpan is a consistent way to attach an oauth2 client's Client ID to a span
func attachOAuth2ClientIDToSpan(span *trace.Span, clientID string) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("client_id", clientID))
	}
}

// GetOAuth2Client gets an OAuth2 client from the database
func (c *Client) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2Client")
	defer span.End()

	attachUserIDToSpan(span, userID)
	attachOAuth2ClientDatabaseIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id": clientID,
		"user_id":   userID,
	})
	logger.Debug("GetOAuth2Client called")

	client, err := c.querier.GetOAuth2Client(ctx, clientID, userID)
	if err != nil {
		logger.Error(err, "error fetching oauth2 client from the querier")
		return nil, err
	}

	return client, nil
}

// GetOAuth2ClientByClientID fetches any OAuth2 client by client ID, regardless of ownership.
// This is used by authenticating middleware to fetch client information it needs to validate.
func (c *Client) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2ClientByClientID")
	defer span.End()

	attachOAuth2ClientIDToSpan(span, clientID)
	logger := c.logger.WithValue("oauth2client_client_id", clientID)
	logger.Debug("GetOAuth2ClientByClientID called")

	client, err := c.querier.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "error fetching oauth2 client from the querier")
		return nil, err
	}

	return client, nil
}

// GetOAuth2ClientCount gets the count of OAuth2 clients in the database that match the current filter
func (c *Client) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2ClientCount")
	defer span.End()

	attachUserIDToSpan(span, userID)
	attachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetOAuth2ClientCount called")

	return c.querier.GetOAuth2ClientCount(ctx, filter, userID)
}

// GetAllOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (c *Client) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllOAuth2ClientCount")
	defer span.End()

	c.logger.Debug("GetAllOAuth2ClientCount called")

	return c.querier.GetAllOAuth2ClientCount(ctx)
}

// GetAllOAuth2ClientsForUser returns all OAuth2 clients belonging to a given user
func (c *Client) GetAllOAuth2ClientsForUser(ctx context.Context, userID uint64) ([]*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllOAuth2ClientsForUser")
	defer span.End()

	attachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("GetAllOAuth2ClientsForUser called")

	return c.querier.GetAllOAuth2ClientsForUser(ctx, userID)
}

// GetAllOAuth2Clients returns all OAuth2 clients, irrespective of ownership.
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

	attachUserIDToSpan(span, userID)
	attachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetOAuth2Clients called")

	return c.querier.GetOAuth2Clients(ctx, filter, userID)
}

// CreateOAuth2Client creates an OAuth2 client
func (c *Client) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "CreateOAuth2Client")
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":  input.ClientID,
		"belongs_to": input.BelongsTo,
	})

	client, err := c.querier.CreateOAuth2Client(ctx, input)
	if err != nil {
		logger.WithError(err).Debug("error writing oauth2 client to the querier")
		return nil, err
	}

	logger.Debug("new oauth2 client created successfully")

	return client, nil
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (c *Client) UpdateOAuth2Client(ctx context.Context, updated *models.OAuth2Client) error {
	ctx, span := trace.StartSpan(ctx, "UpdateOAuth2Client")
	defer span.End()

	return c.querier.UpdateOAuth2Client(ctx, updated)
}

// ArchiveOAuth2Client archives an OAuth2 client
func (c *Client) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "ArchiveOAuth2Client")
	defer span.End()

	attachUserIDToSpan(span, userID)
	attachOAuth2ClientDatabaseIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":  clientID,
		"belongs_to": userID,
	})

	err := c.querier.ArchiveOAuth2Client(ctx, clientID, userID)
	if err != nil {
		logger.WithError(err).Debug("error deleting oauth2 client to the querier")
		return err
	}
	logger.Debug("removed oauth2 client successfully")

	return nil
}

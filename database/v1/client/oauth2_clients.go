package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var _ models.OAuth2ClientDataManager = (*Client)(nil)

// GetOAuth2Client gets an OAuth2 client from the database
func (c *Client) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2Client")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, clientID)

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

	tracing.AttachOAuth2ClientIDToSpan(span, clientID)
	logger := c.logger.WithValue("oauth2client_client_id", clientID)
	logger.Debug("GetOAuth2ClientByClientID called")

	client, err := c.querier.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "error fetching oauth2 client from the querier")
		return nil, err
	}

	return client, nil
}

// GetAllOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (c *Client) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "GetAllOAuth2ClientCount")
	defer span.End()

	c.logger.Debug("GetAllOAuth2ClientCount called")

	return c.querier.GetAllOAuth2ClientCount(ctx)
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (c *Client) GetOAuth2Clients(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.OAuth2ClientList, error) {
	ctx, span := trace.StartSpan(ctx, "GetOAuth2Clients")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetOAuth2Clients called")

	return c.querier.GetOAuth2Clients(ctx, userID, filter)
}

// CreateOAuth2Client creates an OAuth2 client
func (c *Client) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "CreateOAuth2Client")
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":       input.ClientID,
		"belongs_to_user": input.BelongsToUser,
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

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":       clientID,
		"belongs_to_user": userID,
	})

	err := c.querier.ArchiveOAuth2Client(ctx, clientID, userID)
	if err != nil {
		logger.WithError(err).Debug("error deleting oauth2 client to the querier")
		return err
	}
	logger.Debug("removed oauth2 client successfully")

	return nil
}

package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.OAuth2ClientDataManager = (*Client)(nil)

// GetOAuth2Client gets an OAuth2 client from the database.
func (c *Client) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*types.OAuth2Client, error) {
	ctx, span := tracing.StartSpan(ctx)
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
func (c *Client) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*types.OAuth2Client, error) {
	ctx, span := tracing.StartSpan(ctx)
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

// GetTotalOAuth2ClientCount gets the count of OAuth2 clients that match the current filter.
func (c *Client) GetTotalOAuth2ClientCount(ctx context.Context) (uint64, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetTotalOAuth2ClientCount called")

	return c.querier.GetTotalOAuth2ClientCount(ctx)
}

// GetAllOAuth2Clients loads all OAuth2 clients into a channel.
func (c *Client) GetAllOAuth2Clients(ctx context.Context, results chan []types.OAuth2Client) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllItems called")

	return c.querier.GetAllOAuth2Clients(ctx, results)
}

// GetOAuth2Clients gets a list of OAuth2 clients.
func (c *Client) GetOAuth2Clients(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.OAuth2ClientList, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachFilterToSpan(span, filter)

	c.logger.WithValue("user_id", userID).Debug("GetOAuth2Clients called")

	return c.querier.GetOAuth2Clients(ctx, userID, filter)
}

// CreateOAuth2Client creates an OAuth2 client.
func (c *Client) CreateOAuth2Client(ctx context.Context, input *types.OAuth2ClientCreationInput) (*types.OAuth2Client, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		"oauth2client_client_id": input.ClientID,
		"belongs_to_user":        input.BelongsToUser,
	})

	client, err := c.querier.CreateOAuth2Client(ctx, input)
	if err != nil {
		logger.WithError(err).Debug("error writing oauth2 client to the querier")
		return nil, err
	}

	logger.WithValue("client_id", client.ID).Debug("new oauth2 client created successfully")

	return client, nil
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (c *Client) UpdateOAuth2Client(ctx context.Context, updated *types.OAuth2Client) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	return c.querier.UpdateOAuth2Client(ctx, updated)
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (c *Client) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":       clientID,
		"belongs_to_user": userID,
	})

	if err := c.querier.ArchiveOAuth2Client(ctx, clientID, userID); err != nil {
		logger.WithError(err).Debug("error deleting oauth2 client to the querier")
		return err
	}

	logger.Debug("removed oauth2 client successfully")

	return nil
}

// GetAuditLogEntriesForOAuth2Client fetches a list of audit log entries from the database that relate to a given client.
func (c *Client) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]types.AuditLogEntry, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForOAuth2Client called")

	return c.querier.GetAuditLogEntriesForOAuth2Client(ctx, clientID)
}

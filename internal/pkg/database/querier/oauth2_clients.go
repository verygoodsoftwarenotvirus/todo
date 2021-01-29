package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.OAuth2ClientDataManager  = (*Client)(nil)
	_ types.OAuth2ClientAuditManager = (*Client)(nil)
)

// scanOAuth2Client takes a Scanner (i.e. *sql.Row) and scans its results into an OAuth2Client struct.
func (c *Client) scanOAuth2Client(scan database.Scanner, includeCounts bool) (client *types.OAuth2Client, filteredCount, totalCount uint64, err error) {
	client = &types.OAuth2Client{}

	var rawScopes string

	targetVars := []interface{}{
		&client.ID,
		&client.Name,
		&client.ClientID,
		&rawScopes,
		&client.RedirectURI,
		&client.ClientSecret,
		&client.CreatedOn,
		&client.LastUpdatedOn,
		&client.ArchivedOn,
		&client.BelongsToUser,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	if scopes := strings.Split(rawScopes, queriers.OAuth2ClientsTableScopeSeparator); len(scopes) >= 1 && scopes[0] != "" {
		client.Scopes = scopes
	}

	return client, filteredCount, totalCount, nil
}

// scanOAuth2Clients takes sql rows and turns them into a slice of OAuth2Clients.
func (c *Client) scanOAuth2Clients(rows database.ResultIterator, includeCounts bool) (clients []*types.OAuth2Client, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		client, fc, tc, scanErr := c.scanOAuth2Client(rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, scanErr
		}

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}

		clients = append(clients, client)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing rows")
		return nil, 0, 0, closeErr
	}

	return clients, filteredCount, totalCount, nil
}

// GetOAuth2Client gets an OAuth2 client from the database.
func (c *Client) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*types.OAuth2Client, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":    clientID,
		keys.UserIDKey: userID,
	})
	logger.Debug("GetOAuth2Client called")

	query, args := c.sqlQueryBuilder.BuildGetOAuth2ClientQuery(clientID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := c.scanOAuth2Client(row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 client: %w", err)
	}

	return client, nil
}

// GetOAuth2ClientByClientID fetches any OAuth2 client by client ID, regardless of ownership.
// This is used by authenticating middleware to fetch client information it needs to validate.
func (c *Client) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*types.OAuth2Client, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachOAuth2ClientIDToSpan(span, clientID)
	logger := c.logger.WithValue(keys.OAuth2ClientIDKey, clientID)
	logger.Debug("GetOAuth2ClientByClientID called")

	query, args := c.sqlQueryBuilder.BuildGetOAuth2ClientByClientIDQuery(clientID)
	row := c.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := c.scanOAuth2Client(row, false)
	if err != nil {
		return nil, fmt.Errorf("scanning oauth2 client: %w", err)
	}

	return client, nil
}

// GetTotalOAuth2ClientCount gets the count of OAuth2 clients that match the current filter.
func (c *Client) GetTotalOAuth2ClientCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetTotalOAuth2ClientCount called")

	if err = c.db.QueryRowContext(ctx, c.sqlQueryBuilder.BuildGetAllOAuth2ClientsCountQuery()).Scan(&count); err != nil {
		return 0, fmt.Errorf("executing account subscription plans count query: %w", err)
	}

	return count, nil
}

// GetAllOAuth2Clients loads all OAuth2 clients into a channel.
func (c *Client) GetAllOAuth2Clients(ctx context.Context, results chan []*types.OAuth2Client, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllItems called")

	count, countErr := c.GetTotalOAuth2ClientCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of oauth2 clients: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfOAuth2ClientsQuery(begin, end)
			logger := c.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			clients, _, _, scanErr := c.scanOAuth2Clients(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			results <- clients
		}(beginID, endID)
	}

	return nil
}

// GetOAuth2Clients gets a list of OAuth2 clients.
func (c *Client) GetOAuth2Clients(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.OAuth2ClientList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.OAuth2ClientList{}

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("GetOAuth2Clients called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetOAuth2ClientsQuery(userID, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 clients: %w", err)
	}

	if x.Clients, x.FilteredCount, x.TotalCount, err = c.scanOAuth2Clients(rows, true); err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return x, nil
}

// CreateOAuth2Client creates an OAuth2 client.
func (c *Client) CreateOAuth2Client(ctx context.Context, input *types.OAuth2ClientCreationInput) (*types.OAuth2Client, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValues(map[string]interface{}{
		keys.OAuth2ClientDatabaseIDKey: input.ClientID,
		keys.UserIDKey:                 input.BelongsToUser,
	}).Debug("CreateOAuth2Client called")

	query, args := c.sqlQueryBuilder.BuildCreateOAuth2ClientQuery(input)

	id, err := c.performWriteQuery(ctx, "oauth2 client creation", query, args)
	if err != nil {
		return nil, err
	}

	x := &types.OAuth2Client{
		ID:            id,
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		RedirectURI:   input.RedirectURI,
		Scopes:        input.Scopes,
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     c.currentTime(),
	}

	return x, nil
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (c *Client) UpdateOAuth2Client(ctx context.Context, updated *types.OAuth2Client) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	query, args := c.sqlQueryBuilder.BuildUpdateOAuth2ClientQuery(updated)

	return c.execContext(ctx, "oauth2 client update", query, args)
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (c *Client) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, clientID)

	c.logger.WithValues(map[string]interface{}{
		keys.OAuth2ClientDatabaseIDKey: clientID,
		keys.UserIDKey:                 userID,
	}).Debug("ArchiveOAuth2Client called")

	query, args := c.sqlQueryBuilder.BuildArchiveOAuth2ClientQuery(clientID, userID)

	return c.execContext(ctx, "oauth2 client archive", query, args)
}

// LogOAuth2ClientCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, client.BelongsToUser).Debug("LogOAuth2ClientCreationEvent called")

	c.createAuditLogEntry(ctx, audit.BuildOAuth2ClientCreationEventEntry(client))
}

// LogOAuth2ClientArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogOAuth2ClientArchiveEvent called")

	c.createAuditLogEntry(ctx, audit.BuildOAuth2ClientArchiveEventEntry(userID, clientID))
}

// GetAuditLogEntriesForOAuth2Client fetches a list of audit log entries from the database that relate to a given client.
func (c *Client) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForOAuth2Client called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForOAuth2ClientQuery(clientID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.DelegatedClientDataManager  = (*Client)(nil)
	_ types.DelegatedClientAuditManager = (*Client)(nil)
)

// scanDelegatedClient takes a Scanner (i.e. *sql.Row) and scans its results into an DelegatedClient struct.
func (c *Client) scanDelegatedClient(scan database.Scanner, includeCounts bool) (client *types.DelegatedClient, filteredCount, totalCount uint64, err error) {
	client = &types.DelegatedClient{}

	targetVars := []interface{}{
		&client.ID,
		&client.ExternalID,
		&client.Name,
		&client.ClientID,
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

	return client, filteredCount, totalCount, nil
}

// scanDelegatedClients takes sql rows and turns them into a slice of DelegatedClients.
func (c *Client) scanDelegatedClients(rows database.ResultIterator, includeCounts bool) (clients []*types.DelegatedClient, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		client, fc, tc, scanErr := c.scanDelegatedClient(rows, includeCounts)
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

// GetDelegatedClient gets an Delegated client from the database.
func (c *Client) GetDelegatedClient(ctx context.Context, clientID, userID uint64) (*types.DelegatedClient, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachDelegatedClientIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id":    clientID,
		keys.UserIDKey: userID,
	})
	logger.Debug("GetDelegatedClient called")

	query, args := c.sqlQueryBuilder.BuildGetDelegatedClientQuery(clientID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := c.scanDelegatedClient(row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 client: %w", err)
	}

	return client, nil
}

// GetTotalDelegatedClientCount gets the count of Delegated clients that match the current filter.
func (c *Client) GetTotalDelegatedClientCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetTotalDelegatedClientCount called")

	if err = c.db.QueryRowContext(ctx, c.sqlQueryBuilder.BuildGetAllDelegatedClientsCountQuery()).Scan(&count); err != nil {
		return 0, fmt.Errorf("executing account subscription accountsubscriptionplans count query: %w", err)
	}

	return count, nil
}

// GetAllDelegatedClients loads all Delegated clients into a channel.
func (c *Client) GetAllDelegatedClients(ctx context.Context, results chan []*types.DelegatedClient, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllItems called")

	count, countErr := c.GetTotalDelegatedClientCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of oauth2 clients: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfDelegatedClientsQuery(begin, end)
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

			clients, _, _, scanErr := c.scanDelegatedClients(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			results <- clients
		}(beginID, endID)
	}

	return nil
}

// GetDelegatedClients gets a list of Delegated clients.
func (c *Client) GetDelegatedClients(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.DelegatedClientList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.DelegatedClientList{}

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("GetDelegatedClients called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetDelegatedClientsQuery(userID, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 clients: %w", err)
	}

	if x.Clients, x.FilteredCount, x.TotalCount, err = c.scanDelegatedClients(rows, true); err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return x, nil
}

// CreateDelegatedClient creates an Delegated client.
func (c *Client) CreateDelegatedClient(ctx context.Context, input *types.DelegatedClientCreationInput) (*types.DelegatedClient, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValues(map[string]interface{}{
		keys.DelegatedClientIDKey: input.ClientID,
		keys.UserIDKey:            input.BelongsToUser,
	}).Debug("CreateDelegatedClient called")

	query, args := c.sqlQueryBuilder.BuildCreateDelegatedClientQuery(input)

	id, err := c.performCreateQuery(ctx, false, "oauth2 client creation", query, args)
	if err != nil {
		return nil, err
	}

	x := &types.DelegatedClient{
		ID:            id,
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     c.currentTime(),
	}

	return x, nil
}

// UpdateDelegatedClient updates a Delegated client. Note that this function expects the input's
// ID field to be valid.
func (c *Client) UpdateDelegatedClient(ctx context.Context, updated *types.DelegatedClient) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	query, args := c.sqlQueryBuilder.BuildUpdateDelegatedClientQuery(updated)

	return c.performCreateQueryIgnoringReturn(ctx, "oauth2 client update", query, args)
}

// ArchiveDelegatedClient archives an Delegated client.
func (c *Client) ArchiveDelegatedClient(ctx context.Context, clientID, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachDelegatedClientIDToSpan(span, clientID)

	c.logger.WithValues(map[string]interface{}{
		keys.DelegatedClientIDKey: clientID,
		keys.UserIDKey:            userID,
	}).Debug("ArchiveDelegatedClient called")

	query, args := c.sqlQueryBuilder.BuildArchiveDelegatedClientQuery(clientID, userID)

	return c.performCreateQueryIgnoringReturn(ctx, "oauth2 client archive", query, args)
}

// LogDelegatedClientCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogDelegatedClientCreationEvent(ctx context.Context, client *types.DelegatedClient) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, client.BelongsToUser).Debug("LogDelegatedClientCreationEvent called")

	c.createAuditLogEntry(ctx, audit.BuildDelegatedClientCreationEventEntry(client))
}

// LogDelegatedClientArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogDelegatedClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogDelegatedClientArchiveEvent called")

	c.createAuditLogEntry(ctx, audit.BuildDelegatedClientArchiveEventEntry(userID, clientID))
}

// GetAuditLogEntriesForDelegatedClient fetches a list of audit log entries from the database that relate to a given client.
func (c *Client) GetAuditLogEntriesForDelegatedClient(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForDelegatedClient called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForDelegatedClientQuery(clientID)

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

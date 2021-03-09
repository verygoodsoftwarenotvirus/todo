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
	_ types.APIClientDataManager = (*Client)(nil)
)

// scanAPIClient takes a Scanner (i.e. *sql.Row) and scans its results into an APIClient struct.
func (c *Client) scanAPIClient(scan database.Scanner, includeCounts bool) (client *types.APIClient, filteredCount, totalCount uint64, err error) {
	client = &types.APIClient{}

	targetVars := []interface{}{
		&client.ID,
		&client.ExternalID,
		&client.Name,
		&client.ClientID,
		&client.ClientSecret,
		&client.CreatedOn,
		&client.LastUpdatedOn,
		&client.ArchivedOn,
		&client.BelongsToAccount,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	return client, filteredCount, totalCount, nil
}

// scanAPIClients takes sql rows and turns them into a slice of API Clients.
func (c *Client) scanAPIClients(rows database.ResultIterator, includeCounts bool) (clients []*types.APIClient, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		client, fc, tc, scanErr := c.scanAPIClient(rows, includeCounts)
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

// GetAPIClientByClientID gets an API client from the database.
func (c *Client) GetAPIClientByClientID(ctx context.Context, clientID string) (*types.APIClient, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAPIClientClientIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id": clientID,
	})
	logger.Debug("GetAPIClientByClientID called")

	query, args := c.sqlQueryBuilder.BuildGetAPIClientByClientIDQuery(clientID)
	row := c.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := c.scanAPIClient(row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for API client: %w", err)
	}

	return client, nil
}

// GetAPIClientByDatabaseID gets an API client from the database.
func (c *Client) GetAPIClientByDatabaseID(ctx context.Context, clientID, userID uint64) (*types.APIClient, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAPIClientDatabaseIDToSpan(span, clientID)
	tracing.AttachUserIDToSpan(span, userID)

	logger := c.logger.WithValues(map[string]interface{}{
		"client_id": clientID,
		"user_id":   userID,
	})
	logger.Debug("GetAPIClientByDatabaseID called")

	query, args := c.sqlQueryBuilder.BuildGetAPIClientByDatabaseIDQuery(clientID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := c.scanAPIClient(row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for API client: %w", err)
	}

	return client, nil
}

// GetTotalAPIClientCount gets the count of API clients that match the current filter.
func (c *Client) GetTotalAPIClientCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetTotalAPIClientCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllAPIClientsCountQuery())
}

// GetAllAPIClients loads all API clients into a channel.
func (c *Client) GetAllAPIClients(ctx context.Context, results chan []*types.APIClient, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAPIClients called")

	count, countErr := c.GetTotalAPIClientCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of API clients: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfAPIClientsQuery(begin, end)
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

			clients, _, _, scanErr := c.scanAPIClients(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			results <- clients
		}(beginID, endID)
	}

	return nil
}

// GetAPIClients gets a list of API clients.
func (c *Client) GetAPIClients(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.APIClientList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.APIClientList{}

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetAPIClientsQuery(userID, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for API clients: %w", err)
	}

	if x.Clients, x.FilteredCount, x.TotalCount, err = c.scanAPIClients(rows, true); err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	logger.WithValue("result_count", len(x.Clients)).Debug("GetAPIClients called")

	return x, nil
}

// CreateAPIClient creates an API client.
func (c *Client) CreateAPIClient(ctx context.Context, input *types.APICientCreationInput, createdByUser uint64) (*types.APIClient, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValues(map[string]interface{}{
		keys.APIClientClientIDKey: input.ClientID,
		keys.UserIDKey:            input.BelongsToAccount,
	})

	logger.Debug("CreateAPIClient called")

	query, args := c.sqlQueryBuilder.BuildCreateAPIClientQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}

	id, err := c.performWriteQuery(ctx, tx, false, "API client creation", query, args)
	if err != nil {
		c.rollbackTransaction(tx)
		return nil, err
	}

	x := &types.APIClient{
		ID:               id,
		Name:             input.Name,
		ClientID:         input.ClientID,
		ClientSecret:     input.ClientSecret,
		BelongsToAccount: input.BelongsToAccount,
		CreatedOn:        c.currentTime(),
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAPIClientCreationEventEntry(x)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(tx)

		return nil, fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return x, nil
}

// ArchiveAPIClient archives an API client.
func (c *Client) ArchiveAPIClient(ctx context.Context, clientID, accountID, archivedByUser uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, archivedByUser)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachAPIClientDatabaseIDToSpan(span, clientID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.APIClientDatabaseIDKey: clientID,
		keys.AccountIDKey:           accountID,
		keys.UserIDKey:              archivedByUser,
	})

	logger.Debug("ArchiveAPIClient called")

	query, args := c.sqlQueryBuilder.BuildArchiveAPIClientQuery(clientID, accountID)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("error beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "API client archive", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error updating API client: %w", execErr)
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAPIClientArchiveEventEntry(accountID, clientID, archivedByUser)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", commitErr)
	}

	return nil
}

// GetAuditLogEntriesForAPIClient fetches a list of audit log entries from the database that relate to a given client.
func (c *Client) GetAuditLogEntriesForAPIClient(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForAPIClient called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAPIClientQuery(clientID)

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

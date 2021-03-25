package querier

import (
	"context"
	"database/sql"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.APIClientDataManager = (*Client)(nil)
)

// scanAPIClient takes a Scanner (i.e. *sql.Row) and scans its results into an APIClient struct.
func (c *Client) scanAPIClient(ctx context.Context, scan database.Scanner, includeCounts bool) (client *types.APIClient, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

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
		&client.BelongsToUser,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "scanning API client database result")
	}

	return client, filteredCount, totalCount, nil
}

// scanAPIClients takes sql rows and turns them into a slice of API Clients.
func (c *Client) scanAPIClients(ctx context.Context, rows database.ResultIterator, includeCounts bool) (clients []*types.APIClient, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		client, fc, tc, scanErr := c.scanAPIClient(ctx, rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, observability.PrepareError(scanErr, logger, span, "scanning API client")
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

	if err = c.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
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

	client, _, _, err := c.scanAPIClient(ctx, row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying for API client")
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

	client, _, _, err := c.scanAPIClient(ctx, row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying for API client")
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

	logger := c.logger.WithValue("batch_size", batchSize)

	count, err := c.GetTotalAPIClientCount(ctx)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching count of API clients")
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfAPIClientsQuery(begin, end)
			logger = logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if queryErr != nil {
				if !errors.Is(queryErr, sql.ErrNoRows) {
					logger.Error(queryErr, "querying for database rows")
				}
				return
			}

			clients, _, _, scanErr := c.scanAPIClients(ctx, rows, false)
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

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.APIClientList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetAPIClientsQuery(userID, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying for API clients")
	}

	if x.Clients, x.FilteredCount, x.TotalCount, err = c.scanAPIClients(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
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
		keys.UserIDKey:            input.BelongsToUser,
	})

	logger.Debug("CreateAPIClient called")

	query, args := c.sqlQueryBuilder.BuildCreateAPIClientQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	id, err := c.performWriteQuery(ctx, tx, false, "API client creation", query, args)
	if err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating API client")
	}

	client := &types.APIClient{
		ID:            id,
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     c.currentTime(),
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAPIClientCreationEventEntry(client, createdByUser)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing API client creation audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	return client, nil
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

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "API client archive", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating API client")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAPIClientArchiveEventEntry(accountID, clientID, archivedByUser)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing API client archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForAPIClient fetches a list of audit log entries from the database that relate to a given client.
func (c *Client) GetAuditLogEntriesForAPIClient(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.APIClientDatabaseIDKey, clientID)

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAPIClientQuery(clientID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return auditLogEntries, nil
}

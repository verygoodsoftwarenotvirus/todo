package queriers

import (
	"context"
	"database/sql"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	_ types.APIClientDataManager = (*SQLQuerier)(nil)
)

// scanAPIClient takes a Scanner (i.e. *sql.Row) and scans its results into an APIClient struct.
func (q *SQLQuerier) scanAPIClient(ctx context.Context, scan database.Scanner, includeCounts bool) (client *types.APIClient, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	client = &types.APIClient{}

	targetVars := []interface{}{
		&client.ID,
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
func (q *SQLQuerier) scanAPIClients(ctx context.Context, rows database.ResultIterator, includeCounts bool) (clients []*types.APIClient, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		client, fc, tc, scanErr := q.scanAPIClient(ctx, rows, includeCounts)
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

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return clients, filteredCount, totalCount, nil
}

// GetAPIClientByClientID gets an API client from the database.
func (q *SQLQuerier) GetAPIClientByClientID(ctx context.Context, clientID string) (*types.APIClient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == "" {
		return nil, ErrEmptyInputProvided
	}

	tracing.AttachAPIClientClientIDToSpan(span, clientID)
	logger := q.logger.WithValue(keys.APIClientClientIDKey, clientID)

	query, args := q.sqlQueryBuilder.BuildGetAPIClientByClientIDQuery(ctx, clientID)
	row := q.getOneRow(ctx, q.db, "API client", query, args...)

	client, _, _, err := q.scanAPIClient(ctx, row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying for API client")
	}

	return client, nil
}

// GetAPIClientByDatabaseID gets an API client from the database.
func (q *SQLQuerier) GetAPIClientByDatabaseID(ctx context.Context, clientID, userID string) (*types.APIClient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == "" || userID == "" {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAPIClientDatabaseIDToSpan(span, clientID)
	tracing.AttachUserIDToSpan(span, userID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.APIClientDatabaseIDKey: clientID,
		keys.UserIDKey:              userID,
	})

	query, args := q.sqlQueryBuilder.BuildGetAPIClientByDatabaseIDQuery(ctx, clientID, userID)
	row := q.getOneRow(ctx, q.db, "API client", query, args...)

	client, _, _, err := q.scanAPIClient(ctx, row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying for API client")
	}

	return client, nil
}

// GetTotalAPIClientCount gets the count of API clients that match the current filter.
func (q *SQLQuerier) GetTotalAPIClientCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetAllAPIClientsCountQuery(ctx), "fetching count of API clients")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of API clients")
	}

	return count, nil
}

// GetAPIClients gets a list of API clients.
func (q *SQLQuerier) GetAPIClients(ctx context.Context, userID string, filter *types.QueryFilter) (x *types.APIClientList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := filter.AttachToLogger(q.logger).WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.APIClientList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetAPIClientsQuery(ctx, userID, filter)

	rows, err := q.performReadQuery(ctx, q.db, "API clients", query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying for API clients")
	}

	if x.Clients, x.FilteredCount, x.TotalCount, err = q.scanAPIClients(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return x, nil
}

// CreateAPIClient creates an API client.
func (q *SQLQuerier) CreateAPIClient(ctx context.Context, input *types.APIClientCreationInput) (*types.APIClient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.APIClientClientIDKey: input.ClientID,
		keys.UserIDKey:            input.BelongsToUser,
	})

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildCreateAPIClientQuery(ctx, input)

	if writeErr := q.performWriteQueryIgnoringReturn(ctx, tx, "API client creation", query, args); writeErr != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(writeErr, logger, span, "creating API client")
	}

	tracing.AttachAPIClientDatabaseIDToSpan(span, input.ID)

	client := &types.APIClient{
		ID:            input.ID,
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     q.currentTime(),
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("API client created")

	return client, nil
}

// ArchiveAPIClient archives an API client.
func (q *SQLQuerier) ArchiveAPIClient(ctx context.Context, clientID, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == "" || userID == "" {
		return ErrNilInputProvided
	}

	tracing.AttachAccountIDToSpan(span, userID)
	tracing.AttachAPIClientDatabaseIDToSpan(span, clientID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.APIClientDatabaseIDKey: clientID,
		keys.UserIDKey:              userID,
	})

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildArchiveAPIClientQuery(ctx, clientID, userID)

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "API client archive", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating API client")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("API client archived")

	return nil
}

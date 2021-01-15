package postgres

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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.OAuth2ClientDataManager = (*Postgres)(nil)

// scanOAuth2Client takes a Scanner (i.e. *sql.Row) and scans its results into an OAuth2Client struct.
func (q *Postgres) scanOAuth2Client(scan database.Scanner, includeCounts bool) (client *types.OAuth2Client, filteredCount, totalCount uint64, err error) {
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
func (q *Postgres) scanOAuth2Clients(rows database.ResultIterator, includeCounts bool) (clients []types.OAuth2Client, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		client, fc, tc, scanErr := q.scanOAuth2Client(rows, includeCounts)
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

		clients = append(clients, *client)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		q.logger.Error(closeErr, "closing rows")
		return nil, 0, 0, closeErr
	}

	return clients, filteredCount, totalCount, nil
}

// BuildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID.
func (q *Postgres) BuildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	var err error

	// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
	// care about ownership. It does still care about archived status
	query, args, err = q.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableClientIDColumn): clientID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                 nil,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientByClientID gets an OAuth2 client.
func (q *Postgres) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*types.OAuth2Client, error) {
	query, args := q.BuildGetOAuth2ClientByClientIDQuery(clientID)
	row := q.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := q.scanOAuth2Client(row, false)

	return client, err
}

// BuildGetBatchOfOAuth2ClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Postgres) BuildGetBatchOfOAuth2ClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAllOAuth2Clients fetches every item from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (q *Postgres) GetAllOAuth2Clients(ctx context.Context, resultChannel chan []types.OAuth2Client, bucketSize uint16) error {
	count, countErr := q.GetTotalOAuth2ClientCount(ctx)
	if countErr != nil {
		return fmt.Errorf("error fetching count of webhooks: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(bucketSize) {
		endID := beginID + uint64(bucketSize)
		go func(begin, end uint64) {
			query, args := q.BuildGetBatchOfOAuth2ClientsQuery(begin, end)
			logger := q.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := q.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			clients, _, _, scanErr := q.scanOAuth2Clients(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			resultChannel <- clients
		}(beginID, endID)
	}

	return nil
}

// BuildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (q *Postgres) BuildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn):                          clientID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                  nil,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2Client retrieves an OAuth2 client from the database.
func (q *Postgres) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*types.OAuth2Client, error) {
	query, args := q.BuildGetOAuth2ClientQuery(clientID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := q.scanOAuth2Client(row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 client: %w", err)
	}

	return client, nil
}

// BuildGetAllOAuth2ClientsCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (q *Postgres) BuildGetAllOAuth2ClientsCountQuery() string {
	var err error

	getAllOAuth2ClientCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.OAuth2ClientsTableName)).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return getAllOAuth2ClientCountQuery
}

// GetTotalOAuth2ClientCount will get the count of OAuth2 clients that match the current filter.
func (q *Postgres) GetTotalOAuth2ClientCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := q.db.QueryRowContext(ctx, q.BuildGetAllOAuth2ClientsCountQuery()).Scan(&count)

	return count, err
}

// BuildGetOAuth2ClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (q *Postgres) BuildGetOAuth2ClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		queriers.OAuth2ClientsTableName,
		queriers.OAuth2ClientsTableOwnershipColumn,
		queriers.OAuth2ClientsTableColumns,
		userID,
		false,
		filter,
	)
}

// GetOAuth2Clients gets a list of OAuth2 clients.
func (q *Postgres) GetOAuth2Clients(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.OAuth2ClientList, error) {
	query, args := q.BuildGetOAuth2ClientsQuery(userID, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 clients: %w", err)
	}

	list, filteredCount, totalCount, err := q.scanOAuth2Clients(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	ocl := &types.OAuth2ClientList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: filteredCount,
			TotalCount:    totalCount,
		},
		Clients: list,
	}

	return ocl, nil
}

// BuildCreateOAuth2ClientQuery returns a SQL query (and args) that will create the given OAuth2Client in the database.
func (q *Postgres) BuildCreateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(queriers.OAuth2ClientsTableName).
		Columns(
			queriers.OAuth2ClientsTableNameColumn,
			queriers.OAuth2ClientsTableClientIDColumn,
			queriers.OAuth2ClientsTableClientSecretColumn,
			queriers.OAuth2ClientsTableScopesColumn,
			queriers.OAuth2ClientsTableRedirectURIColumn,
			queriers.OAuth2ClientsTableOwnershipColumn,
		).
		Values(
			input.Name,
			input.ClientID,
			input.ClientSecret,
			strings.Join(input.Scopes, queriers.OAuth2ClientsTableScopeSeparator),
			input.RedirectURI,
			input.BelongsToUser,
		).
		Suffix(fmt.Sprintf("RETURNING %s, %s", queriers.IDColumn, queriers.CreatedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// CreateOAuth2Client creates an OAuth2 client.
func (q *Postgres) CreateOAuth2Client(ctx context.Context, input *types.OAuth2ClientCreationInput) (*types.OAuth2Client, error) {
	x := &types.OAuth2Client{
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		RedirectURI:   input.RedirectURI,
		Scopes:        input.Scopes,
		BelongsToUser: input.BelongsToUser,
	}
	query, args := q.BuildCreateOAuth2ClientQuery(x)

	err := q.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, fmt.Errorf("error executing client creation query: %w", err)
	}

	return x, nil
}

// buildUpdateOAuth2ClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (q *Postgres) buildUpdateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.OAuth2ClientsTableName).
		Set(queriers.OAuth2ClientsTableClientIDColumn, input.ClientID).
		Set(queriers.OAuth2ClientsTableClientSecretColumn, input.ClientSecret).
		Set(queriers.OAuth2ClientsTableScopesColumn, strings.Join(input.Scopes, queriers.OAuth2ClientsTableScopeSeparator)).
		Set(queriers.OAuth2ClientsTableRedirectURIColumn, input.RedirectURI).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                          input.ID,
			queriers.OAuth2ClientsTableOwnershipColumn: input.BelongsToUser,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.LastUpdatedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateOAuth2Client updates a OAuth2 client.
// NOTE: this function expects the input's ID field to be valid and non-zero.
func (q *Postgres) UpdateOAuth2Client(ctx context.Context, input *types.OAuth2Client) error {
	query, args := q.buildUpdateOAuth2ClientQuery(input)
	return q.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// BuildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *Postgres) BuildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.OAuth2ClientsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                          clientID,
			queriers.OAuth2ClientsTableOwnershipColumn: userID,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.ArchivedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (q *Postgres) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	query, args := q.BuildArchiveOAuth2ClientQuery(clientID, userID)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// LogOAuth2ClientCreationEvent saves a OAuth2ClientCreationEvent in the audit log table.
func (q *Postgres) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	q.CreateAuditLogEntry(ctx, audit.BuildOAuth2ClientCreationEventEntry(client))
}

// LogOAuth2ClientArchiveEvent saves a OAuth2ClientArchiveEvent in the audit log table.
func (q *Postgres) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	q.CreateAuditLogEntry(ctx, audit.BuildOAuth2ClientArchiveEventEntry(userID, clientID))
}

// BuildGetAuditLogEntriesForOAuth2ClientQuery constructs a SQL query for fetching audit log entries
// associated with a given oauth2 client.
func (q *Postgres) BuildGetAuditLogEntriesForOAuth2ClientQuery(clientID uint64) (query string, args []interface{}) {
	clientIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.OAuth2ClientAssignmentKey)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{clientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	return q.buildQuery(builder)
}

// GetAuditLogEntriesForOAuth2Client fetches a audit log entries for a given oauth2 client from the database.
func (q *Postgres) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]types.AuditLogEntry, error) {
	query, args := q.BuildGetAuditLogEntriesForOAuth2ClientQuery(clientID)

	q.logger.WithValue(keys.OAuth2ClientIDKey, clientID).Debug("GetAuditLogEntriesForOAuth2Client invoked")

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

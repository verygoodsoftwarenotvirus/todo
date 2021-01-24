package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.OAuth2ClientDataManager  = (*Sqlite)(nil)
	_ types.OAuth2ClientAuditManager = (*Sqlite)(nil)
)

// scanOAuth2Client takes a Scanner (i.e. *sql.Row) and scans its results into an OAuth2Client struct.
func (c *Sqlite) scanOAuth2Client(scan database.Scanner, includeCounts bool) (client *types.OAuth2Client, filteredCount, totalCount uint64, err error) {
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
func (c *Sqlite) scanOAuth2Clients(rows database.ResultIterator, includeCounts bool) (clients []*types.OAuth2Client, filteredCount, totalCount uint64, err error) {
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

// buildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID.
func (c *Sqlite) buildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	var err error

	// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
	// care about ownership. It does still care about archived status
	query, args, err = c.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableClientIDColumn): clientID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                 nil,
		}).ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientByClientID gets an OAuth2 client.
func (c *Sqlite) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*types.OAuth2Client, error) {
	query, args := c.buildGetOAuth2ClientByClientIDQuery(clientID)
	row := c.db.QueryRowContext(ctx, query, args...)

	client, _, _, err := c.scanOAuth2Client(row, false)

	return client, err
}

// buildGetBatchOfOAuth2ClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (c *Sqlite) buildGetBatchOfOAuth2ClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := c.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetAllOAuth2Clients fetches every item from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (c *Sqlite) GetAllOAuth2Clients(ctx context.Context, resultChannel chan []*types.OAuth2Client, batchSize uint16) error {
	count, countErr := c.GetTotalOAuth2ClientCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of oauth2 clients: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.buildGetBatchOfOAuth2ClientsQuery(begin, end)
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

			resultChannel <- clients
		}(beginID, endID)
	}

	return nil
}

// buildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (c *Sqlite) buildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn):                          clientID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                  nil,
		}).ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2Client retrieves an OAuth2 client from the database.
func (c *Sqlite) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*types.OAuth2Client, error) {
	query, args := c.buildGetOAuth2ClientQuery(clientID, userID)
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

// buildGetAllOAuth2ClientsCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (c *Sqlite) buildGetAllOAuth2ClientsCountQuery() string {
	var err error

	getAllOAuth2ClientCountQuery, _, err := c.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.OAuth2ClientsTableName)).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return getAllOAuth2ClientCountQuery
}

// GetTotalOAuth2ClientCount will get the count of OAuth2 clients that match the current filter.
func (c *Sqlite) GetTotalOAuth2ClientCount(ctx context.Context) (count uint64, err error) {
	err = c.db.QueryRowContext(ctx, c.buildGetAllOAuth2ClientsCountQuery()).Scan(&count)
	return count, err
}

// buildGetOAuth2ClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (c *Sqlite) buildGetOAuth2ClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	return c.buildListQuery(
		queriers.OAuth2ClientsTableName,
		queriers.OAuth2ClientsTableOwnershipColumn,
		queriers.OAuth2ClientsTableColumns,
		userID,
		false,
		filter,
	)
}

// GetOAuth2Clients gets a list of OAuth2 clients.
func (c *Sqlite) GetOAuth2Clients(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.OAuth2ClientList, error) {
	query, args := c.buildGetOAuth2ClientsQuery(userID, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 clients: %w", err)
	}

	list, filteredCount, totalCount, err := c.scanOAuth2Clients(rows, true)
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

// buildCreateOAuth2ClientQuery returns a SQL query (and args) that will create the given OAuth2Client in the database.
func (c *Sqlite) buildCreateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
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
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// CreateOAuth2Client creates an OAuth2 client.
func (c *Sqlite) CreateOAuth2Client(ctx context.Context, input *types.OAuth2ClientCreationInput) (*types.OAuth2Client, error) {
	x := &types.OAuth2Client{
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		RedirectURI:   input.RedirectURI,
		Scopes:        input.Scopes,
		BelongsToUser: input.BelongsToUser,
	}
	query, args := c.buildCreateOAuth2ClientQuery(x)

	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing client creation query: %w", err)
	}

	x.ID = c.getIDFromResult(res)
	x.CreatedOn = c.timeTeller.Now()

	return x, nil
}

// buildUpdateOAuth2ClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (c *Sqlite) buildUpdateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
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
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// UpdateOAuth2Client updates a OAuth2 client.
// NOTE: this function expects the input's ID field to be valid and non-zero.
func (c *Sqlite) UpdateOAuth2Client(ctx context.Context, input *types.OAuth2Client) error {
	query, args := c.buildUpdateOAuth2ClientQuery(input)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (c *Sqlite) buildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Update(queriers.OAuth2ClientsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                          clientID,
			queriers.OAuth2ClientsTableOwnershipColumn: userID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (c *Sqlite) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	query, args := c.buildArchiveOAuth2ClientQuery(clientID, userID)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// LogOAuth2ClientCreationEvent saves a OAuth2ClientCreationEvent in the audit log table.
func (c *Sqlite) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	c.createAuditLogEntry(ctx, audit.BuildOAuth2ClientCreationEventEntry(client))
}

// LogOAuth2ClientArchiveEvent saves a OAuth2ClientArchiveEvent in the audit log table.
func (c *Sqlite) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	c.createAuditLogEntry(ctx, audit.BuildOAuth2ClientArchiveEventEntry(userID, clientID))
}

// buildGetAuditLogEntriesForOAuth2ClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (c *Sqlite) buildGetAuditLogEntriesForOAuth2ClientQuery(clientID uint64) (query string, args []interface{}) {
	var err error

	oauth2ClientIDKey := fmt.Sprintf(
		jsonPluckQuery,
		queriers.AuditLogEntriesTableName,
		queriers.AuditLogEntriesTableContextColumn,
		audit.OAuth2ClientAssignmentKey,
	)
	builder := c.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{oauth2ClientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	query, args, err = builder.ToSql()
	c.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForOAuth2Client fetches an audit log entry from the database.
func (c *Sqlite) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	query, args := c.buildGetAuditLogEntriesForOAuth2ClientQuery(clientID)

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

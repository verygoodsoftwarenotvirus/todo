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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// scanOAuth2Client takes a Scanner (i.e. *sql.Row) and scans its results into an OAuth2Client struct.
func (q *Postgres) scanOAuth2Client(scan database.Scanner, includeCount bool) (*types.OAuth2Client, uint64, error) {
	var (
		x         = &types.OAuth2Client{}
		rawScopes string
		count     uint64
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.ClientID,
		&rawScopes,
		&x.RedirectURI,
		&x.ClientSecret,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	if scopes := strings.Split(rawScopes, queriers.OAuth2ClientsTableScopeSeparator); len(scopes) >= 1 && scopes[0] != "" {
		x.Scopes = scopes
	}

	return x, count, nil
}

// scanOAuth2Clients takes sql rows and turns them into a slice of OAuth2Clients.
func (q *Postgres) scanOAuth2Clients(rows database.ResultIterator, includeCount bool) ([]*types.OAuth2Client, uint64, error) {
	var (
		list  []*types.OAuth2Client
		count uint64
	)

	for rows.Next() {
		client, c, err := q.scanOAuth2Client(rows, includeCount)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 && includeCount {
			count = c
		}

		list = append(list, client)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := rows.Close(); err != nil {
		q.logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID.
func (q *Postgres) buildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
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
	query, args := q.buildGetOAuth2ClientByClientIDQuery(clientID)
	row := q.db.QueryRowContext(ctx, query, args...)

	client, _, err := q.scanOAuth2Client(row, false)

	return client, err
}

// buildGetAllOAuth2ClientsQuery builds a SQL query.
func (q *Postgres) buildGetAllOAuth2ClientsQuery() (query string) {
	var err error

	getAllOAuth2ClientsQuery, _, err := q.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return getAllOAuth2ClientsQuery
}

// GetAllOAuth2Clients gets a list of OAuth2 clients regardless of ownership.
func (q *Postgres) GetAllOAuth2Clients(ctx context.Context) ([]*types.OAuth2Client, error) {
	rows, err := q.db.QueryContext(ctx, q.buildGetAllOAuth2ClientsQuery())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying database for oauth2 clients: %w", err)
	}

	list, _, err := q.scanOAuth2Clients(rows, false)
	if err != nil {
		return nil, fmt.Errorf("fetching list of OAuth2Clients: %w", err)
	}

	return list, nil
}

// GetAllOAuth2ClientsForUser gets a list of OAuth2 clients belonging to a given user.
func (q *Postgres) GetAllOAuth2ClientsForUser(ctx context.Context, userID uint64) ([]*types.OAuth2Client, error) {
	query, args := q.buildGetOAuth2ClientsForUserQuery(userID, nil)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying database for oauth2 clients: %w", err)
	}

	list, _, err := q.scanOAuth2Clients(rows, false)
	if err != nil {
		return nil, fmt.Errorf("fetching list of OAuth2Clients: %w", err)
	}

	return list, nil
}

// buildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (q *Postgres) buildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
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
	query, args := q.buildGetOAuth2ClientQuery(clientID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	client, _, err := q.scanOAuth2Client(row, false)
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
func (q *Postgres) buildGetAllOAuth2ClientsCountQuery() string {
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

// GetAllOAuth2ClientCount will get the count of OAuth2 clients that match the current filter.
func (q *Postgres) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := q.db.QueryRowContext(ctx, q.buildGetAllOAuth2ClientsCountQuery()).Scan(&count)

	return count, err
}

// buildGetOAuth2ClientsForUserQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (q *Postgres) buildGetOAuth2ClientsForUserQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.PlaceholderFormat(squirrel.Question).
		Select(allCountQuery).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                  nil,
		})

	if filter != nil {
		countQueryBuilder = queriers.ApplyFilterToSubCountQueryBuilder(filter, countQueryBuilder, queriers.ItemsTableName)
	}

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(queriers.OAuth2ClientsTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                  nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.OAuth2ClientsTableName)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetOAuth2ClientsForUser gets a list of OAuth2 clients.
func (q *Postgres) GetOAuth2ClientsForUser(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.OAuth2ClientList, error) {
	query, args := q.buildGetOAuth2ClientsForUserQuery(userID, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 clients: %w", err)
	}

	list, count, err := q.scanOAuth2Clients(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	ocl := &types.OAuth2ClientList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
	}

	// de-pointer-ize clients
	ocl.Clients = make([]types.OAuth2Client, len(list))
	for i, t := range list {
		ocl.Clients[i] = *t
	}

	return ocl, nil
}

// buildCreateOAuth2ClientQuery returns a SQL query (and args) that will create the given OAuth2Client in the database.
func (q *Postgres) buildCreateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
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
	query, args := q.buildCreateOAuth2ClientQuery(x)

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

// buildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *Postgres) buildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
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
	query, args := q.buildArchiveOAuth2ClientQuery(clientID, userID)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// LogOAuth2ClientCreationEvent saves a OAuth2ClientCreationEvent in the audit log table.
func (q *Postgres) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	q.createAuditLogEntry(ctx, audit.BuildOAuth2ClientCreationEventEntry(client))
}

// LogOAuth2ClientArchiveEvent saves a OAuth2ClientArchiveEvent in the audit log table.
func (q *Postgres) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildOAuth2ClientArchiveEventEntry(userID, clientID))
}

// buildGetAuditLogEntriesForOAuth2ClientQuery constructs a SQL query for fetching audit log entries
// associated with a given oauth2 client.
func (q *Postgres) buildGetAuditLogEntriesForOAuth2ClientQuery(clientID uint64) (query string, args []interface{}) {
	var err error

	clientIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.OAuth2ClientAssignmentKey)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{clientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForOAuth2Client fetches a audit log entries for a given oauth2 client from the database.
func (q *Postgres) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]types.AuditLogEntry, error) {
	query, args := q.buildGetAuditLogEntriesForOAuth2ClientQuery(clientID)

	q.logger.WithValue("query", query).WithValue("client_id", clientID).Debug("GetAuditLogEntriesForOAuth2Client invoked")

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

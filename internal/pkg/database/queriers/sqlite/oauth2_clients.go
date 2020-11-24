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

// scanOAuth2Client takes a Scanner (i.e. *sql.Row) and scans its results into an OAuth2Client struct.
func (s *Sqlite) scanOAuth2Client(scan database.Scanner) (*types.OAuth2Client, error) {
	var (
		x      = &types.OAuth2Client{}
		scopes string
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.ClientID,
		&scopes,
		&x.RedirectURI,
		&x.ClientSecret,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	if scopes := strings.Split(scopes, queriers.OAuth2ClientsTableScopeSeparator); len(scopes) >= 1 && scopes[0] != "" {
		x.Scopes = scopes
	}

	return x, nil
}

// scanOAuth2Clients takes sql rows and turns them into a slice of OAuth2Clients.
func (s *Sqlite) scanOAuth2Clients(rows database.ResultIterator) ([]*types.OAuth2Client, error) {
	var (
		list []*types.OAuth2Client
	)

	for rows.Next() {
		client, err := s.scanOAuth2Client(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, client)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		s.logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID.
func (s *Sqlite) buildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	var err error

	// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
	// care about ownership. It does still care about archived status
	query, args, err = s.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableClientIDColumn): clientID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                 nil,
		}).ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientByClientID gets an OAuth2 client.
func (s *Sqlite) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*types.OAuth2Client, error) {
	query, args := s.buildGetOAuth2ClientByClientIDQuery(clientID)
	row := s.db.QueryRowContext(ctx, query, args...)

	return s.scanOAuth2Client(row)
}

// buildGetAllOAuth2ClientsQuery builds a SQL query.
func (s *Sqlite) buildGetAllOAuth2ClientsQuery() (query string) {
	var err error

	getAllOAuth2ClientsQuery, _, err := s.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return getAllOAuth2ClientsQuery
}

// GetAllOAuth2Clients gets a list of OAuth2 clients regardless of ownership.
func (s *Sqlite) GetAllOAuth2Clients(ctx context.Context) ([]*types.OAuth2Client, error) {
	rows, err := s.db.QueryContext(ctx, s.buildGetAllOAuth2ClientsQuery())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying database for oauth2 clients: %w", err)
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, fmt.Errorf("fetching list of OAuth2Clients: %w", err)
	}

	return list, nil
}

// GetAllOAuth2ClientsForUser gets a list of OAuth2 clients belonging to a given user.
func (s *Sqlite) GetAllOAuth2ClientsForUser(ctx context.Context, userID uint64) ([]*types.OAuth2Client, error) {
	query, args := s.buildGetOAuth2ClientsForUserQuery(userID, nil)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying database for oauth2 clients: %w", err)
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, fmt.Errorf("fetching list of OAuth2Clients: %w", err)
	}

	return list, nil
}

// buildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (s *Sqlite) buildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn):                          clientID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                  nil,
		}).ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2Client retrieves an OAuth2 client from the database.
func (s *Sqlite) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*types.OAuth2Client, error) {
	query, args := s.buildGetOAuth2ClientQuery(clientID, userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	client, err := s.scanOAuth2Client(row)
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
func (s *Sqlite) buildGetAllOAuth2ClientsCountQuery() string {
	var err error

	getAllOAuth2ClientCountQuery, _, err := s.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.OAuth2ClientsTableName)).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return getAllOAuth2ClientCountQuery
}

// GetAllOAuth2ClientCount will get the count of OAuth2 clients that match the current filter.
func (s *Sqlite) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := s.db.QueryRowContext(ctx, s.buildGetAllOAuth2ClientsCountQuery()).Scan(&count)

	return count, err
}

// buildGetOAuth2ClientsForUserQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (s *Sqlite) buildGetOAuth2ClientsForUserQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
		Select(queriers.OAuth2ClientsTableColumns...).
		From(queriers.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.ArchivedOnColumn):                  nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.OAuth2ClientsTableName, queriers.IDColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, queriers.OAuth2ClientsTableName)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientsForUser gets a list of OAuth2 clients.
func (s *Sqlite) GetOAuth2ClientsForUser(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.OAuth2ClientList, error) {
	query, args := s.buildGetOAuth2ClientsForUserQuery(userID, filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying for oauth2 clients: %w", err)
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	ocl := &types.OAuth2ClientList{
		Pagination: types.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
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
func (s *Sqlite) buildCreateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
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

	s.logQueryBuildingError(err)

	return query, args
}

// CreateOAuth2Client creates an OAuth2 client.
func (s *Sqlite) CreateOAuth2Client(ctx context.Context, input *types.OAuth2ClientCreationInput) (*types.OAuth2Client, error) {
	x := &types.OAuth2Client{
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		RedirectURI:   input.RedirectURI,
		Scopes:        input.Scopes,
		BelongsToUser: input.BelongsToUser,
	}
	query, args := s.buildCreateOAuth2ClientQuery(x)

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing client creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	s.logIDRetrievalError(err)

	x.ID = uint64(id)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = s.timeTeller.Now()

	return x, nil
}

// buildUpdateOAuth2ClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (s *Sqlite) buildUpdateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
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

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateOAuth2Client updates a OAuth2 client.
// NOTE: this function expects the input's ID field to be valid and non-zero.
func (s *Sqlite) UpdateOAuth2Client(ctx context.Context, input *types.OAuth2Client) error {
	query, args := s.buildUpdateOAuth2ClientQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (s *Sqlite) buildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(queriers.OAuth2ClientsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                          clientID,
			queriers.OAuth2ClientsTableOwnershipColumn: userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (s *Sqlite) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	query, args := s.buildArchiveOAuth2ClientQuery(clientID, userID)
	_, err := s.db.ExecContext(ctx, query, args...)

	return err
}

// LogOAuth2ClientCreationEvent saves a OAuth2ClientCreationEvent in the audit log table.
func (s *Sqlite) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	s.createAuditLogEntry(ctx, audit.BuildOAuth2ClientCreationEventEntry(client))
}

// LogOAuth2ClientArchiveEvent saves a OAuth2ClientArchiveEvent in the audit log table.
func (s *Sqlite) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildOAuth2ClientArchiveEventEntry(userID, clientID))
}

// buildGetAuditLogEntriesForOAuth2ClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (s *Sqlite) buildGetAuditLogEntriesForOAuth2ClientQuery(clientID uint64) (query string, args []interface{}) {
	var err error

	oauth2ClientIDKey := fmt.Sprintf(
		jsonPluckQuery,
		queriers.AuditLogEntriesTableName,
		queriers.AuditLogEntriesTableContextColumn,
		audit.OAuth2ClientAssignmentKey,
	)
	builder := s.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{oauth2ClientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForOAuth2Client fetches an audit log entry from the database.
func (s *Sqlite) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]types.AuditLogEntry, error) {
	query, args := s.buildGetAuditLogEntriesForOAuth2ClientQuery(clientID)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := s.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

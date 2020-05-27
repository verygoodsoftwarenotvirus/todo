package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
)

const (
	scopesSeparator                   = ","
	oauth2ClientsTableName            = "oauth2_clients"
	oauth2ClientsTableOwnershipColumn = "belongs_to_user"
)

var (
	oauth2ClientsTableColumns = []string{
		fmt.Sprintf("%s.id", oauth2ClientsTableName),
		fmt.Sprintf("%s.name", oauth2ClientsTableName),
		fmt.Sprintf("%s.client_id", oauth2ClientsTableName),
		fmt.Sprintf("%s.scopes", oauth2ClientsTableName),
		fmt.Sprintf("%s.redirect_uri", oauth2ClientsTableName),
		fmt.Sprintf("%s.client_secret", oauth2ClientsTableName),
		fmt.Sprintf("%s.created_on", oauth2ClientsTableName),
		fmt.Sprintf("%s.updated_on", oauth2ClientsTableName),
		fmt.Sprintf("%s.archived_on", oauth2ClientsTableName),
		fmt.Sprintf("%s.%s", oauth2ClientsTableName, oauth2ClientsTableOwnershipColumn),
	}
)

// scanOAuth2Client takes a Scanner (i.e. *sql.Row) and scans its results into an OAuth2Client struct.
func (m *MariaDB) scanOAuth2Client(scan database.Scanner, includeCount bool) (*models.OAuth2Client, uint64, error) {
	var (
		x      = &models.OAuth2Client{}
		scopes string
		count  uint64
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.ClientID,
		&scopes,
		&x.RedirectURI,
		&x.ClientSecret,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	if scopes := strings.Split(scopes, scopesSeparator); len(scopes) >= 1 && scopes[0] != "" {
		x.Scopes = scopes
	}

	return x, count, nil
}

// scanOAuth2Clients takes sql rows and turns them into a slice of OAuth2Clients.
func (m *MariaDB) scanOAuth2Clients(rows database.ResultIterator) ([]*models.OAuth2Client, uint64, error) {
	var (
		list  []*models.OAuth2Client
		count uint64
	)

	for rows.Next() {
		client, c, err := m.scanOAuth2Client(rows, true)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 {
			count = c
		}

		list = append(list, client)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := rows.Close(); err != nil {
		m.logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID.
func (m *MariaDB) buildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	var err error

	// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
	// care about ownership. It does still care about archived status
	query, args, err = m.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.client_id", oauth2ClientsTableName):   clientID,
			fmt.Sprintf("%s.archived_on", oauth2ClientsTableName): nil,
		}).ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientByClientID gets an OAuth2 client.
func (m *MariaDB) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	query, args := m.buildGetOAuth2ClientByClientIDQuery(clientID)
	row := m.db.QueryRowContext(ctx, query, args...)

	client, _, err := m.scanOAuth2Client(row, false)
	return client, err
}

var (
	getAllOAuth2ClientsQueryBuilder sync.Once
	getAllOAuth2ClientsQuery        string
)

// buildGetAllOAuth2ClientsQuery builds a SQL query.
func (m *MariaDB) buildGetAllOAuth2ClientsQuery() (query string) {
	getAllOAuth2ClientsQueryBuilder.Do(func() {
		var err error

		getAllOAuth2ClientsQuery, _, err = m.sqlBuilder.
			Select(oauth2ClientsTableColumns...).
			From(oauth2ClientsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", oauth2ClientsTableName): nil,
			}).
			ToSql()

		m.logQueryBuildingError(err)
	})

	return getAllOAuth2ClientsQuery
}

// GetAllOAuth2Clients gets a list of OAuth2 clients regardless of ownership.
func (m *MariaDB) GetAllOAuth2Clients(ctx context.Context) ([]*models.OAuth2Client, error) {
	rows, err := m.db.QueryContext(ctx, m.buildGetAllOAuth2ClientsQuery())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database for oauth2 clients: %w", err)
	}

	list, _, err := m.scanOAuth2Clients(rows)
	if err != nil {
		return nil, fmt.Errorf("fetching list of OAuth2Clients: %w", err)
	}

	return list, nil
}

// GetAllOAuth2ClientsForUser gets a list of OAuth2 clients belonging to a given user.
func (m *MariaDB) GetAllOAuth2ClientsForUser(ctx context.Context, userID uint64) ([]*models.OAuth2Client, error) {
	query, args := m.buildGetOAuth2ClientsQuery(userID, nil)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database for oauth2 clients: %w", err)
	}

	list, _, err := m.scanOAuth2Clients(rows)
	if err != nil {
		return nil, fmt.Errorf("fetching list of OAuth2Clients: %w", err)
	}

	return list, nil
}

// buildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (m *MariaDB) buildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", oauth2ClientsTableName):                                    clientID,
			fmt.Sprintf("%s.%s", oauth2ClientsTableName, oauth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.archived_on", oauth2ClientsTableName):                           nil,
		}).ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2Client retrieves an OAuth2 client from the database.
func (m *MariaDB) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	query, args := m.buildGetOAuth2ClientQuery(clientID, userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	client, _, err := m.scanOAuth2Client(row, false)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying for oauth2 client: %w", err)
	}

	return client, nil
}

var (
	getAllOAuth2ClientCountQueryBuilder sync.Once
	getAllOAuth2ClientCountQuery        string
)

// buildGetAllOAuth2ClientCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (m *MariaDB) buildGetAllOAuth2ClientCountQuery() string {
	getAllOAuth2ClientCountQueryBuilder.Do(func() {
		var err error

		getAllOAuth2ClientCountQuery, _, err = m.sqlBuilder.
			Select(fmt.Sprintf(countQuery, oauth2ClientsTableName)).
			From(oauth2ClientsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", oauth2ClientsTableName): nil,
			}).
			ToSql()

		m.logQueryBuildingError(err)
	})

	return getAllOAuth2ClientCountQuery
}

// GetAllOAuth2ClientCount will get the count of OAuth2 clients that match the current filter.
func (m *MariaDB) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := m.db.QueryRowContext(ctx, m.buildGetAllOAuth2ClientCountQuery()).Scan(&count)
	return count, err
}

// buildGetOAuth2ClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (m *MariaDB) buildGetOAuth2ClientsQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(append(oauth2ClientsTableColumns, fmt.Sprintf(countQuery, oauth2ClientsTableName))...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", oauth2ClientsTableName, oauth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.archived_on", oauth2ClientsTableName):                           nil,
		}).
		GroupBy(fmt.Sprintf("%s.id", oauth2ClientsTableName))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, oauth2ClientsTableName)
	}

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientsForUser gets a list of OAuth2 clients.
func (m *MariaDB) GetOAuth2ClientsForUser(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.OAuth2ClientList, error) {
	query, args := m.buildGetOAuth2ClientsQuery(userID, filter)
	rows, err := m.db.QueryContext(ctx, query, args...)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying for oauth2 clients: %w", err)
	}

	list, count, err := m.scanOAuth2Clients(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	ocl := &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
	}

	// de-pointer-ize clients
	ocl.Clients = make([]models.OAuth2Client, len(list))
	for i, t := range list {
		ocl.Clients[i] = *t
	}

	return ocl, nil
}

// buildCreateOAuth2ClientQuery returns a SQL query (and args) that will create the given OAuth2Client in the database
func (m *MariaDB) buildCreateOAuth2ClientQuery(input *models.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Insert(oauth2ClientsTableName).
		Columns(
			"name",
			"client_id",
			"client_secret",
			"scopes",
			"redirect_uri",
			oauth2ClientsTableOwnershipColumn,
		).
		Values(
			input.Name,
			input.ClientID,
			input.ClientSecret,
			strings.Join(input.Scopes, scopesSeparator),
			input.RedirectURI,
			input.BelongsToUser,
		).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// CreateOAuth2Client creates an OAuth2 client.
func (m *MariaDB) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	x := &models.OAuth2Client{
		Name:          input.Name,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		RedirectURI:   input.RedirectURI,
		Scopes:        input.Scopes,
		BelongsToUser: input.BelongsToUser,
	}
	query, args := m.buildCreateOAuth2ClientQuery(x)

	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing client creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	m.logIDRetrievalError(err)
	x.ID = uint64(id)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = m.timeTeller.Now()

	return x, nil
}

// buildUpdateOAuth2ClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database
func (m *MariaDB) buildUpdateOAuth2ClientQuery(input *models.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(oauth2ClientsTableName).
		Set("client_id", input.ClientID).
		Set("client_secret", input.ClientSecret).
		Set("scopes", strings.Join(input.Scopes, scopesSeparator)).
		Set("redirect_uri", input.RedirectURI).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                              input.ID,
			oauth2ClientsTableOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateOAuth2Client updates a OAuth2 client.
// NOTE: this function expects the input's ID field to be valid and non-zero.
func (m *MariaDB) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) error {
	query, args := m.buildUpdateOAuth2ClientQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (m *MariaDB) buildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(oauth2ClientsTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                              clientID,
			oauth2ClientsTableOwnershipColumn: userID,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (m *MariaDB) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	query, args := m.buildArchiveOAuth2ClientQuery(clientID, userID)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

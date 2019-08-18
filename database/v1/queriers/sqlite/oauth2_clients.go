package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

const (
	scopesSeparator = `,`

	oauth2ClientsTableName = "oauth2_clients"
)

var (
	oauth2ClientsTableColumns = []string{
		"id",
		"name",
		"client_id",
		"scopes",
		"redirect_uri",
		"client_secret",
		"created_on",
		"updated_on",
		"archived_on",
		"belongs_to",
	}
)

// scanOAuth2Client takes a Scanner (i.e. *sql.Row) and scans its ressults into an OAuth2Client struct
func scanOAuth2Client(scan database.Scanner) (*models.OAuth2Client, error) {
	var (
		x      = &models.OAuth2Client{}
		scopes string
	)

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.ClientID,
		&scopes,
		&x.RedirectURI,
		&x.ClientSecret,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsTo,
	); err != nil {
		return nil, err
	}

	if scopes := strings.Split(scopes, scopesSeparator); len(scopes) >= 1 && scopes[0] != "" {
		x.Scopes = scopes
	}

	return x, nil
}

// scanOAuth2Clients takes sql rows and turns them into a slice of OAuth2Clients
func (s *Sqlite) scanOAuth2Clients(rows *sql.Rows) ([]*models.OAuth2Client, error) {
	var list []*models.OAuth2Client

	for rows.Next() {
		client, err := scanOAuth2Client(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, client)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	s.logQueryBuildingError(rows.Close())

	return list, nil
}

// buildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID
func (s *Sqlite) buildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	var err error
	// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
	// care about ownership. It does still care about archived status
	query, args, err = s.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			"client_id":   clientID,
			"archived_on": nil,
		}).ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientByClientID gets an OAuth2 client
func (s *Sqlite) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	query, args := s.buildGetOAuth2ClientByClientIDQuery(clientID)
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanOAuth2Client(row)
}

var (
	getAllOAuth2ClientsQueryBuilder sync.Once
	getAllOAuth2ClientsQuery        string
)

// buildGetAllOAuth2ClientsQuery builds a SQL query
func (s *Sqlite) buildGetAllOAuth2ClientsQuery() (query string) {
	getAllOAuth2ClientsQueryBuilder.Do(func() {
		var err error
		getAllOAuth2ClientsQuery, _, err = s.sqlBuilder.
			Select(oauth2ClientsTableColumns...).
			From(oauth2ClientsTableName).
			Where(squirrel.Eq{"archived_on": nil}).
			ToSql()

		s.logQueryBuildingError(err)
	})

	return getAllOAuth2ClientsQuery
}

// GetAllOAuth2Clients gets a list of OAuth2 clients regardless of ownership
func (s *Sqlite) GetAllOAuth2Clients(ctx context.Context) ([]*models.OAuth2Client, error) {
	rows, err := s.db.QueryContext(ctx, s.buildGetAllOAuth2ClientsQuery())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrap(err, "querying database for oauth2 clients")
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, errors.Wrap(err, "fetching list of OAuth2Clients")
	}

	return list, nil
}

// GetAllOAuth2ClientsForUser gets a list of OAuth2 clients belonging to a given user
func (s *Sqlite) GetAllOAuth2ClientsForUser(ctx context.Context, userID uint64) ([]*models.OAuth2Client, error) {
	query, args := s.buildGetOAuth2ClientsQuery(nil, userID)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrap(err, "querying database for oauth2 clients")
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, errors.Wrap(err, "fetching list of OAuth2Clients")
	}

	return list, nil
}

// buildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID
func (s *Sqlite) buildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			"id":          clientID,
			"belongs_to":  userID,
			"archived_on": nil,
		}).ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2Client retrieves an OAuth2 client from the database
func (s *Sqlite) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	query, args := s.buildGetOAuth2ClientQuery(clientID, userID)
	row := s.db.QueryRowContext(ctx, query, args...)
	client, err := scanOAuth2Client(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrap(err, "querying for oauth2 client")
	}

	return client, nil
}

// buildGetOAuth2ClientCountQuery returns a SQL query (and arguments) that fetches a list of OAuth2 clients that meet certain filter
// restrictions (if relevant) and belong to a given user
func (s *Sqlite) buildGetOAuth2ClientCountQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := s.sqlBuilder.
		Select(CountQuery).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			"belongs_to":  userID,
			"archived_on": nil,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2ClientCount will get the count of OAuth2 clients that match the given filter and belong to the user
func (s *Sqlite) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	query, args := s.buildGetOAuth2ClientCountQuery(filter, userID)
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

var (
	getAllOAuth2ClientCountQueryBuilder sync.Once
	getAllOAuth2ClientCountQuery        string
)

// buildGetAllOAuth2ClientCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (s *Sqlite) buildGetAllOAuth2ClientCountQuery() string {
	getAllOAuth2ClientCountQueryBuilder.Do(func() {
		var err error
		getAllOAuth2ClientCountQuery, _, err = s.sqlBuilder.Select(CountQuery).
			From(oauth2ClientsTableName).
			Where(squirrel.Eq{"archived_on": nil}).
			ToSql()

		s.logQueryBuildingError(err)
	})

	return getAllOAuth2ClientCountQuery
}

// GetAllOAuth2ClientCount will get the count of OAuth2 clients that match the current filter
func (s *Sqlite) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := s.db.QueryRowContext(ctx, s.buildGetAllOAuth2ClientCountQuery()).Scan(&count)
	return count, err
}

// buildGetOAuth2ClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (s *Sqlite) buildGetOAuth2ClientsQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := s.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			"belongs_to":  userID,
			"archived_on": nil,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (s *Sqlite) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	query, args := s.buildGetOAuth2ClientsQuery(filter, userID)
	rows, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrap(err, "querying for oauth2 clients")
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, errors.Wrap(err, "scanning response from database")
	}

	// de-pointer-ize clients
	ll := len(list)
	var clients = make([]models.OAuth2Client, ll)
	for i, t := range list {
		clients[i] = *t
	}

	totalCount, err := s.GetOAuth2ClientCount(ctx, filter, userID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching oauth2 client count")
	}

	ocl := &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: totalCount,
		},
		Clients: clients,
	}

	return ocl, nil
}

// buildCreateOAuth2ClientQuery returns a SQL query (and args) that will create the given OAuth2Client in the database
func (s *Sqlite) buildCreateOAuth2ClientQuery(input *models.OAuth2Client) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Insert(oauth2ClientsTableName).
		Columns(
			"name",
			"client_id",
			"client_secret",
			"scopes",
			"redirect_uri",
			"belongs_to",
		).
		Values(
			input.Name,
			input.ClientID,
			input.ClientSecret,
			strings.Join(input.Scopes, scopesSeparator),
			input.RedirectURI,
			input.BelongsTo,
		).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (s *Sqlite) buildOAuth2ClientCreationTimeQuery(clientID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select("created_on").
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{"id": clientID}).
		ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// CreateOAuth2Client creates an OAuth2 client
func (s *Sqlite) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	x := &models.OAuth2Client{
		Name:         input.Name,
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		RedirectURI:  input.RedirectURI,
		Scopes:       input.Scopes,
		BelongsTo:    input.BelongsTo,
	}
	query, args := s.buildCreateOAuth2ClientQuery(x)

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error executing client creation query")
	}

	// fetch the last inserted ID
	if id, idErr := res.LastInsertId(); idErr == nil {
		x.ID = uint64(id)

		query, args = s.buildOAuth2ClientCreationTimeQuery(x.ID)
		s.logCreationTimeRetrievalError(s.db.QueryRowContext(ctx, query, args...).Scan(&x.CreatedOn))
	}

	return x, nil
}

// buildUpdateOAuth2ClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database
func (s *Sqlite) buildUpdateOAuth2ClientQuery(input *models.OAuth2Client) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Update(oauth2ClientsTableName).
		Set("client_id", input.ClientID).
		Set("client_secret", input.ClientSecret).
		Set("scopes", strings.Join(input.Scopes, scopesSeparator)).
		Set("redirect_uri", input.RedirectURI).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":         input.ID,
			"belongs_to": input.BelongsTo,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateOAuth2Client updates a OAuth2 client.
// NOTE: this function expects the input's ID field to be valid and non-zero.
func (s *Sqlite) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) error {
	query, args := s.buildUpdateOAuth2ClientQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (s *Sqlite) buildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Update(oauth2ClientsTableName).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":         clientID,
			"belongs_to": userID,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveOAuth2Client archives an OAuth2 client
func (s *Sqlite) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	query, args := s.buildArchiveOAuth2ClientQuery(clientID, userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}
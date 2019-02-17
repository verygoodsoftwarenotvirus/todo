package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	scopesSeparator           = `,`
	getOAuth2ClientCountQuery = `
		SELECT
			COUNT(*)
		FROM
			oauth_clients
		WHERE 
			archived_on IS NULL
			AND belongs_to = ?
	`
	getOAuth2ClientQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			id = ?
			AND belongs_to = ?
	`
	getOAuth2ClientByClientIDQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			client_id = ?
			AND belongs_to = ?
	`
	getAnyOAuth2ClientQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			client_id = ?
			AND belongs_to = ?
	`
	getAllOAuth2ClientsQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			archived_on IS NULL
	`
	getOAuth2ClientsQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			archived_on IS NULL
			AND belongs_to = ?
		LIMIT ?
		OFFSET ?
	`
	createOAuth2ClientQuery = `
		INSERT INTO oauth_clients
		(
			client_id, client_secret, scopes, redirect_uri, belongs_to
		)
		VALUES
		(
			?, ?, ?, ?, ?
		)
	`
	updateOAuth2ClientQuery = `
		UPDATE oauth_clients SET
			client_id = ?,
			client_secret = ?,
			scopes = ?,
			redirect_uri = ?,
			updated_on = (strftime('%s','now'))
		WHERE 
			id = ?
	`
	archiveOAuth2ClientQuery = `
		UPDATE oauth_clients SET
			updated_on = (strftime('%s','now')),
			archived_on = (strftime('%s','now'))
		WHERE 
			id = ?
			AND belongs_to = ?
	`
)

func (s *Sqlite) scanOAuth2Client(scan database.Scannable) (*models.OAuth2Client, error) {
	var (
		x      = &models.OAuth2Client{}
		scopes string
	)
	err := scan.Scan(
		&x.ID,
		&x.ClientID,
		&scopes,
		&x.RedirectURI,
		&x.ClientSecret,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsTo,
	)
	if err != nil {
		return nil, err
	}

	x.Scopes = strings.Split(scopes, scopesSeparator)

	return x, nil
}

func (s *Sqlite) scanOAuth2Clients(rows *sql.Rows) ([]models.OAuth2Client, error) {
	var list []models.OAuth2Client

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		x, err := s.scanOAuth2Client(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

var _ models.OAuth2ClientHandler = (*Sqlite)(nil)

// GetOAuth2Client gets an OAuth2 client
func (s *Sqlite) GetOAuth2Client(ctx context.Context, clientID string, userID uint64) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetOAuth2Client")
	defer span.Finish()

	s.logger.WithValue("client_id", clientID).Debug("GetOAuth2Client called")
	row := s.database.QueryRow(getOAuth2ClientByClientIDQuery, clientID, userID)
	return s.scanOAuth2Client(row)
}

// GetOAuth2ClientByClientID fetches any OAuth2 client by client ID, regardless of ownership. This is used by
// authenticating middleware to fetch client information it needs to validate
func (s *Sqlite) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetOAuth2Client")
	defer span.Finish()

	s.logger.WithValue("client_id", clientID).Debug("GetOAuth2Client called")
	row := s.database.QueryRow(getAnyOAuth2ClientQuery, clientID)
	return s.scanOAuth2Client(row)
}

// GetOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (s *Sqlite) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetOAuth2ClientCount")
	defer span.Finish()

	var count uint64
	err := s.database.QueryRow(getOAuth2ClientCountQuery, userID).Scan(&count)
	return count, err
}

// GetAllOAuth2Clients returns all OAuth2 clients, irrespective of ownership. It is called on startup to populate
// the OAuth2 Client handler
func (s *Sqlite) GetAllOAuth2Clients(ctx context.Context) ([]models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetAllOAuth2Clients")
	defer span.Finish()
	s.logger.Debug("Postgres.GetAllOAuth2Clients called")

	prep, err := s.database.Prepare(getAllOAuth2ClientsQuery)
	if err != nil {
		s.logger.Error(err, "error preparing query")
		return nil, err
	}

	rows, err := prep.Query()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			s.logger.Error(err, "closing rows")
		}
	}()

	var list []models.OAuth2Client
	for rows.Next() {
		var x *models.OAuth2Client
		x, err = s.scanOAuth2Client(rows)
		if err != nil {
			s.logger.Error(err, "error encountered scanning OAuth2Client")
			return nil, err
		}
		list = append(list, *x)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error(err, "error encountered fetching list of OAuth2Clients")
		return nil, err
	}

	return list, err
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (s *Sqlite) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetOAuth2Clients")
	defer span.Finish()

	filter = s.prepareFilter(filter, span)
	rows, err := s.database.Query(getOAuth2ClientsQuery, userID, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, err
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, err
	}

	ocl := &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Clients: list,
	}
	if ocl.TotalCount, err = s.GetOAuth2ClientCount(ctx, filter, userID); err != nil {
		return nil, err
	}

	return ocl, err
}

func prepareOAuth2Client(input *models.OAuth2ClientCreationInput) *models.OAuth2Client {
	x := &models.OAuth2Client{
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		RedirectURI:  input.RedirectURI,
		Scopes:       input.Scopes,
		BelongsTo:    input.BelongsTo,
	}
	return x
}

// CreateOAuth2Client creates an OAuth2 client
func (s *Sqlite) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "CreateOAuth2Client")
	defer span.Finish()

	s.logger.Debug("CreateOAuth2Client called.")

	// create the client
	x := prepareOAuth2Client(input)
	res, err := s.database.Exec(
		createOAuth2ClientQuery,
		x.ClientID,
		x.ClientSecret,
		strings.Join(x.Scopes, scopesSeparator),
		x.RedirectURI,
		x.BelongsTo,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error executing client creation query")
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching last inserted item ID")
	}

	// fetch full updated client
	row := s.database.QueryRow(getOAuth2ClientQuery, id)
	if x, err = s.scanOAuth2Client(row); err != nil {
		return nil, errors.Wrap(err, "error fetching newly created")
	}

	s.logger.Debug("returning from CreateOAuth2Client")
	return x, nil
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (s *Sqlite) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) error {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "UpdateOAuth2Client")
	defer span.Finish()

	// update the client
	_, err := s.database.Exec(
		updateOAuth2ClientQuery,
		input.ClientID,
		input.ClientSecret,
		strings.Join(input.Scopes, scopesSeparator),
		input.RedirectURI,
		input.ID,
	)

	return err
}

// DeleteOAuth2Client deletes an OAuth2 client
func (s *Sqlite) DeleteOAuth2Client(ctx context.Context, identifier string, userID uint64) error {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "DeleteOAuth2Client")
	defer span.Finish()

	_, err := s.database.Exec(archiveOAuth2ClientQuery, identifier, userID)
	return err
}

package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const scopesSeparator = `,`

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

func scanOAuth2Client(scan Scannable) (*models.OAuth2Client, error) {
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
		x, err := scanOAuth2Client(rows)
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

const getOAuth2ClientQuery = `
	SELECT
		id,
		client_id,
		scopes,
		redirect_uri,
		client_secret,
		created_on,
		updated_on,
		archived_on,
		belongs_to
	FROM
		oauth_clients
	WHERE
		client_id = ?
		AND belongs_to = ?
`

// GetOAuth2Client gets an OAuth2 client
func (s *Sqlite) GetOAuth2Client(ctx context.Context, clientID string, userID uint64) (*models.OAuth2Client, error) {
	s.logger.WithValue("client_id", clientID).Debug("GetOAuth2Client called")
	row := s.database.QueryRowContext(ctx, getOAuth2ClientQuery, clientID, userID)
	return scanOAuth2Client(row)
}

const getOAuth2ClientByClientIDQuery = `
	SELECT
		id,
		client_id,
		scopes,
		redirect_uri,
		client_secret,
		created_on,
		updated_on,
		archived_on,
		belongs_to
	FROM
		oauth_clients
	WHERE
		client_id = ?
`

// GetOAuth2ClientByClientID gets an OAuth2 client regardless of ownership
func (s *Sqlite) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	s.logger.WithValue("client_id", clientID).Debug("GetOAuth2Client called")
	row := s.database.QueryRowContext(ctx, getOAuth2ClientByClientIDQuery, clientID)
	return scanOAuth2Client(row)
}

const getOAuth2ClientCountQuery = `
	SELECT
		COUNT(*)
	FROM
		oauth_clients
	WHERE
		archived_on is null
`

// GetOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (s *Sqlite) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	var count uint64
	err := s.database.QueryRowContext(ctx, getOAuth2ClientCountQuery).Scan(&count)
	return count, err
}

const getOAuth2ClientsQuery = `
	SELECT
		id,
		client_id,
		scopes,
		redirect_uri,
		client_secret,
		created_on,
		updated_on,
		archived_on,
		belongs_to
	FROM
		oauth_clients
	WHERE
		archived_on is null
	LIMIT ?
	OFFSET ?
`

// GetOAuth2Clients gets a list of OAuth2 clients
func (s *Sqlite) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	rows, err := s.database.QueryContext(ctx, getOAuth2ClientsQuery, filter.Limit, filter.QueryPage())
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

const getAllOAuth2ClientsQuery = `
	SELECT
		id,
		client_id,
		scopes,
		redirect_uri,
		client_secret,
		created_on,
		updated_on,
		archived_on,
		belongs_to
	FROM
		oauth_clients
	WHERE
		archived_on IS NULL
`

// GetAllOAuth2Clients gets a list of all OAuth2 clients, regardless of ownership
func (s *Sqlite) GetAllOAuth2Clients(ctx context.Context) ([]models.OAuth2Client, error) {
	rows, err := s.database.QueryContext(ctx, getAllOAuth2ClientsQuery)
	if err != nil {
		return nil, err
	}

	list, err := s.scanOAuth2Clients(rows)
	if err != nil {
		return nil, err
	}

	return list, err
}

const getOAuth2ClientByIDQuery = `
	SELECT
		id,
		client_id,
		scopes,
		redirect_uri,
		client_secret,
		created_on,
		updated_on,
		archived_on,
		belongs_to
	FROM
		oauth_clients
	WHERE
		id = ?
`

const createOAuth2ClientQuery = `
	INSERT INTO oauth_clients
	(
		client_id,
		client_secret,
		scopes,
		redirect_uri,
		belongs_to
	)
	VALUES
	(
		?, ?, ?, ?, ?
	)
`

// CreateOAuth2Client creates an OAuth2 client
func (s *Sqlite) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	x := prepareOAuth2Client(input)
	res, err := s.database.ExecContext(ctx,
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
	row := s.database.QueryRowContext(ctx, getOAuth2ClientByIDQuery, id)
	if x, err = scanOAuth2Client(row); err != nil {
		return nil, errors.Wrap(err, "error fetching newly created")
	}

	s.logger.Debug("returning from CreateOAuth2Client")
	return x, nil
}

const updateOAuth2ClientQuery = `
	UPDATE oauth_clients
	SET
		client_id = ?,
		client_secret = ?,
		scopes = ?,
		redirect_uri = ?,
		updated_on = (strftime('%s','now'))
	WHERE
		id = ?
		AND belongs_to = ?
`

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (s *Sqlite) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) error {
	_, err := s.database.ExecContext(ctx,
		updateOAuth2ClientQuery,
		input.ClientID,
		input.ClientSecret,
		strings.Join(input.Scopes, scopesSeparator),
		input.RedirectURI,
		input.ID,
		input.BelongsTo,
	)

	return err
}

const archiveOAuth2ClientQuery = `
	UPDATE oauth_clients
	SET
		updated_on = (strftime('%s','now')),
		archived_on = (strftime('%s','now'))
	WHERE
		id = ?
		AND belongs_to = ?
`

// DeleteOAuth2Client deletes an OAuth2 client
func (s *Sqlite) DeleteOAuth2Client(ctx context.Context, id string, userID uint64) error {
	_, err := s.database.ExecContext(ctx, archiveOAuth2ClientQuery, id, userID)
	return err
}

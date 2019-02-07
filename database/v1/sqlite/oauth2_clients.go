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
		WHERE archived_on is null
	`
	getOAuth2ClientQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			id = ?
	`
	getOAuth2ClientByClientIDQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			client_id = ?
	`
	getOAuth2ClientsQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			archived_on is null
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
		WHERE id = ?
	`
	archiveOAuth2ClientQuery = `
		UPDATE oauth_clients SET
			updated_on = (strftime('%s','now')),
			archived_on = (strftime('%s','now'))
		WHERE id = ?
	`
)

func scanOAuth2Client(scan database.Scannable) (*models.OAuth2Client, error) {
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

func scanOAuth2Clients(rows *sql.Rows) ([]models.OAuth2Client, error) {
	list := []models.OAuth2Client{}
	defer rows.Close()
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

var _ models.OAuth2ClientHandler = (*Sqlite)(nil)

// GetOAuth2Client gets an OAuth2 client
func (s *Sqlite) GetOAuth2Client(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetOAuth2Client")
	defer span.Finish()

	s.logger.WithValue("client_id", clientID).Debug("GetOAuth2Client called")
	row := s.database.QueryRow(getOAuth2ClientByClientIDQuery, clientID)
	return scanOAuth2Client(row)
}

// GetOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (s *Sqlite) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetOAuth2ClientCount")
	defer span.Finish()

	var count uint64
	err := s.database.QueryRow(getOAuth2ClientCountQuery).Scan(&count)
	return count, err
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (s *Sqlite) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter) (*models.OAuth2ClientList, error) {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "GetOAuth2Clients")
	defer span.Finish()

	filter = s.prepareFilter(filter, span)
	rows, err := s.database.Query(getOAuth2ClientsQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, err
	}

	list, err := scanOAuth2Clients(rows)
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
	if ocl.TotalCount, err = s.GetOAuth2ClientCount(ctx, filter); err != nil {
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
	if x, err = scanOAuth2Client(row); err != nil {
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
func (s *Sqlite) DeleteOAuth2Client(ctx context.Context, id string) error {
	span := tracing.FetchSpanFromContext(ctx, s.tracer, "DeleteOAuth2Client")
	defer span.Finish()

	_, err := s.database.Exec(archiveOAuth2ClientQuery, id)
	return err
}

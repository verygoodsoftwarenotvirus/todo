package postgres

import (
	"context"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	scopesSeparator = `,`
)

func (p Postgres) scanOAuth2Client(scan database.Scanner) (*models.OAuth2Client, error) {
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
		$1, $2, $3, $4, $5
	)
	RETURNING
		id,
		created_on
`

// CreateOAuth2Client creates an OAuth2 client
func (p *Postgres) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	x := &models.OAuth2Client{
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		RedirectURI:  input.RedirectURI,
		Scopes:       input.Scopes,
		BelongsTo:    input.BelongsTo,
	}

	err := p.database.QueryRowContext(
		ctx,
		createOAuth2ClientQuery,
		x.ClientID,
		x.ClientSecret,
		strings.Join(x.Scopes, scopesSeparator),
		x.RedirectURI,
		x.BelongsTo,
	).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, errors.Wrap(err, "error executing client creation query")
	}

	return x, nil
}

// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
// care about ownership. It does still care about archived or not
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
		client_id = $1
		AND archived_on IS NULL
`

// GetOAuth2ClientByClientID gets an OAuth2 client
func (p *Postgres) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	row := p.database.QueryRowContext(ctx, getOAuth2ClientByClientIDQuery, clientID)
	client, err := p.scanOAuth2Client(row)

	if err != nil {
		return nil, err
	}
	return client, nil
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

// GetAllOAuth2Clients gets a list of OAuth2 clients regardless of ownership
func (p *Postgres) GetAllOAuth2Clients(ctx context.Context) ([]models.OAuth2Client, error) {
	rows, err := p.database.QueryContext(ctx, getAllOAuth2ClientsQuery)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	var list []models.OAuth2Client
	for rows.Next() {
		var x *models.OAuth2Client
		x, err = p.scanOAuth2Client(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scanning OAuth2Client")
		}
		list = append(list, *x)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "fetching list of OAuth2Clients")
	}

	return list, nil
}

// begin "standard" CRUD operations that respect

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
		id = $1
		AND belongs_to = $2
		AND archived_on IS NULL
`

// GetOAuth2Client gets an OAuth2 client
func (p *Postgres) GetOAuth2Client(ctx context.Context, id, userID uint64) (*models.OAuth2Client, error) {
	row := p.database.QueryRowContext(ctx, getOAuth2ClientQuery, id, userID)
	client, err := p.scanOAuth2Client(row)

	if err != nil {
		return nil, err
	}
	return client, nil
}

const getOAuth2ClientCountQuery = `
	SELECT
		COUNT(*)
	FROM
		oauth_clients
	WHERE
		archived_on IS NULL
		AND belongs_to = $1
`

// GetOAuth2ClientCount will get the count of OAuth2 clients that match the current filter
func (p *Postgres) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	var count uint64
	err := p.database.QueryRowContext(ctx, getOAuth2ClientCountQuery, userID).Scan(&count)
	return count, err
}

const getAllOAuth2ClientCountQuery = `
	SELECT
		COUNT(*)
	FROM
		oauth_clients
	WHERE
		archived_on IS NULL
`

// GetAllOAuth2ClientCount will get the count of OAuth2 clients that match the current filter
func (p *Postgres) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := p.database.QueryRowContext(ctx, getAllOAuth2ClientCountQuery).Scan(&count)
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
		archived_on IS NULL
		AND belongs_to = $1
	LIMIT $2
	OFFSET $3
`

// GetOAuth2Clients gets a list of OAuth2 clients
func (p *Postgres) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	rows, err := p.database.QueryContext(ctx, getOAuth2ClientsQuery, userID, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, errors.Wrap(err, "executing query")
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	var list []models.OAuth2Client
	for rows.Next() {
		var x *models.OAuth2Client
		x, err = p.scanOAuth2Client(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scanning OAuth2Client")
		}
		list = append(list, *x)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "fetching list of OAuth2Clients")
	}

	ocl := &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: 666,
		},
		Clients: list,
	}
	if ocl.TotalCount, err = p.GetOAuth2ClientCount(ctx, filter, userID); err != nil {
		return nil, err
	}

	return ocl, err
}

const updateOAuth2ClientQuery = `
	UPDATE oauth_clients SET
		client_id = $1,
		client_secret = $2,
		scopes = $3,
		redirect_uri = $4,
		updated_on = extract(epoch FROM NOW())
	WHERE
		id = $5
		AND belongs_to = $6
	RETURNING
		updated_on
`

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (p *Postgres) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) error {
	err := p.database.QueryRowContext(
		ctx,
		updateOAuth2ClientQuery,
		input.ClientID,
		input.ClientSecret,
		strings.Join(input.Scopes, scopesSeparator),
		input.RedirectURI,
		input.ID,
		input.BelongsTo,
	).Scan(&input.UpdatedOn)

	if err != nil {
		return errors.Wrap(err, "error preparing query")
	}

	return nil
}

const archiveOAuth2ClientQuery = `
	UPDATE oauth_clients SET
		updated_on = extract(epoch FROM NOW()),
		archived_on = extract(epoch FROM NOW())
	WHERE
		id = $1
		AND belongs_to = $2
	RETURNING
		archived_on
`

// DeleteOAuth2Client deletes an OAuth2 client
func (p *Postgres) DeleteOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	_, err := p.database.ExecContext(ctx, archiveOAuth2ClientQuery, clientID, userID)

	return err
}

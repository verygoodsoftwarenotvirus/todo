package postgres

import (
	"context"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
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
			id = $1
	`
	getOAuth2ClientByClientIDQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			client_id = $1
	`
	getOAuth2ClientsQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			archived_on is null
		LIMIT $1
		OFFSET $2
	`
	createOAuth2ClientQuery = `
		INSERT INTO oauth_clients
		(
			client_id, client_secret, scopes, redirect_uri, belongs_to
		)
		VALUES
		(
			$1, $2, $3, $4, $5
		)
		RETURNING
			id, created_on
	`
	updateOAuth2ClientQuery = `
		UPDATE oauth_clients SET
			client_id = $1,
			client_secret = $2,
			scopes = $3,
			redirect_uri = $4,
			updated_on = extract(epoch FROM NOW())
		WHERE id = $5
		RETURNING
			updated_on
	`
	archiveOAuth2ClientQuery = `
		UPDATE oauth_clients SET
			updated_on = extract(epoch FROM NOW()),
			archived_on = extract(epoch FROM NOW())
		WHERE id = $1
		RETURNING
			archived_on
	`
)

func (p Postgres) scanOAuth2Client(scan database.Scannable) (*models.OAuth2Client, error) {
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

var _ models.OAuth2ClientHandler = (*Postgres)(nil)

// GetOAuth2Client gets an OAuth2 client
func (p *Postgres) GetOAuth2Client(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetOAuth2Client")
	defer span.Finish()

	logger := p.logger.WithValue("oauth2_client_id", clientID)
	logger.Debug("Postgres.GetOAuth2Client called")

	prep, err := p.database.Prepare(getOAuth2ClientByClientIDQuery)
	if err != nil {
		logger.Error(err, "error preparing OAuth2 retrieval query")
		return nil, err
	}

	row := prep.QueryRow(clientID)
	client, err := p.scanOAuth2Client(row)
	if err != nil {
		logger.Error(err, "error scanning returned row")
		return nil, err
	}

	logger.WithError(nil).Debug("returning from Postgres.GetOAuth2Client")

	return client, nil
}

// GetOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (p *Postgres) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetOAuth2ClientCount")
	defer span.Finish()

	logger := p.logger.WithValue("filter", filter)
	logger.Debug("Postgres.GetOAuth2ClientCount called")

	prep, err := p.database.Prepare(getOAuth2ClientCountQuery)
	if err != nil {
		logger.Error(err, "error preparing OAuth2 count retrieval query")
		return 0, err
	}

	var count uint64
	err = prep.QueryRow().Scan(&count)
	return count, err
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (p *Postgres) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter) (*models.OAuth2ClientList, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "GetOAuth2Clients")
	defer span.Finish()

	logger := p.logger.WithValue("filter", filter)
	logger.Debug("Postgres.GetOAuth2Clients called")

	if filter == nil {
		logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}

	filter.SetPage(filter.Page)
	queryPage := filter.QueryPage()
	logger = logger.WithValue("query_page", queryPage)

	prep, err := p.database.Prepare(getOAuth2ClientsQuery)
	if err != nil {
		logger.Error(err, "error preparing  query")
		return nil, err
	}

	rows, err := prep.Query(filter.Limit, queryPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.OAuth2Client{}
	for rows.Next() {
		x, err := p.scanOAuth2Client(rows)
		if err != nil {
			logger.Error(err, "error encountered scanning OAuth2Client")
			return nil, err
		}
		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		logger.Error(err, "error encountered fetching list of OAuth2Clients")
		return nil, err
	}

	ocl := &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: 666,
		},
		Clients: list,
	}
	if ocl.TotalCount, err = p.GetOAuth2ClientCount(ctx, filter); err != nil {
		return nil, err
	}

	return ocl, err
}

// CreateOAuth2Client creates an OAuth2 client
func (p *Postgres) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "CreateOAuth2Client")
	defer span.Finish()

	logger := p.logger.WithValues(map[string]interface{}{
		"redirect_uri": input.RedirectURI,
		"scopes":       input.Scopes,
		"belongs_to":   input.BelongsTo,
	})
	logger.Debug("CreateOAuth2Client called.")

	var err error
	x := &models.OAuth2Client{
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		RedirectURI:  input.RedirectURI,
		Scopes:       input.Scopes,
		BelongsTo:    input.BelongsTo,
	}

	prep, err := p.database.Prepare(createOAuth2ClientQuery)
	if err != nil {
		logger.Error(err, "error preparing  query")
		return nil, err
	}

	// create the client
	if err = prep.QueryRow(
		x.ClientID,
		x.ClientSecret,
		strings.Join(x.Scopes, scopesSeparator),
		x.RedirectURI,
		x.BelongsTo,
	).Scan(&x.ID, &x.CreatedOn); err != nil {
		logger.Error(err, "error executing client creation query")
		return nil, err
	}

	return x, err
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (p *Postgres) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) error {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "UpdateOAuth2Client")
	defer span.Finish()

	logger := p.logger.WithValues(map[string]interface{}{
		"redirect_uri": input.RedirectURI,
		"scopes":       input.Scopes,
		"belongs_to":   input.BelongsTo,
	})
	logger.Debug("UpdateOAuth2Client called.")

	prep, err := p.database.Prepare(updateOAuth2ClientQuery)
	if err != nil {
		logger.Error(err, "error preparing  query")
		return err
	}

	err = prep.QueryRow(
		input.ClientID,
		input.ClientSecret,
		strings.Join(input.Scopes, scopesSeparator),
		input.RedirectURI,
		input.ID,
	).Scan(&input.UpdatedOn)

	return err
}

// DeleteOAuth2Client deletes an OAuth2 client
func (p *Postgres) DeleteOAuth2Client(ctx context.Context, id string) error {
	span := tracing.FetchSpanFromContext(ctx, p.tracer, "DeleteOAuth2Client")
	defer span.Finish()

	logger := p.logger.WithValue("id", id)
	logger.Debug("Postgres.DeleteOAuth2Client called")

	prep, err := p.database.Prepare(archiveOAuth2ClientQuery)
	if err != nil {
		logger.Error(err, "error preparing  query")
		return err
	}

	_, err = prep.Exec(id)
	return err
}

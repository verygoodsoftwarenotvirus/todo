package sqlite

import (
	"context"
	"math"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
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

var _ models.OAuth2ClientHandler = (*Sqlite)(nil)

// GetOAuth2Client gets an OAuth2 client
func (s *Sqlite) GetOAuth2Client(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	s.logger.Debugf("GetOAuth2Client called for %s", clientID)
	row := s.database.QueryRow(getOAuth2ClientByClientIDQuery, clientID)
	return scanOAuth2Client(row)
}

// GetOAuth2ClientCount gets the count of OAuth2 clients that match the current filter
func (s *Sqlite) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	var count uint64
	err := s.database.QueryRow(getOAuth2ClientCountQuery).Scan(&count)
	return count, err
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (s *Sqlite) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter) (*models.OAuth2ClientList, error) {
	if filter == nil {
		s.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.Page = uint64(math.Max(1, float64(filter.Page)))
	queryPage := uint(filter.Limit * (filter.Page - 1))

	list := []models.OAuth2Client{}

	s.logger.Debugf("query limit: %d, query page: %d, calculated page: %d", filter.Limit, filter.Page, queryPage)

	rows, err := s.database.Query(getOAuth2ClientsQuery, filter.Limit, queryPage)
	if err != nil {
		return nil, err
	}
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

	ocl := &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: 666,
		},
		Clients: list,
	}
	if ocl.TotalCount, err = s.GetOAuth2ClientCount(ctx, filter); err != nil {
		return nil, err
	}

	return ocl, err
}

// CreateOAuth2Client creates an OAuth2 client
func (s *Sqlite) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (x *models.OAuth2Client, err error) {
	s.logger.Debugln("CreateOAuth2Client called.")

	x = &models.OAuth2Client{
		RedirectURI: input.RedirectURI,
		Scopes:      input.Scopes,
	}

	if x.ClientID, err = auth.RandString(64); err != nil {
		return nil, err
	}

	if x.ClientSecret, err = auth.RandString(64); err != nil {
		return nil, err
	}

	tx, err := s.database.Begin()
	if err != nil {
		s.logger.Errorf("error beginning database connection: %v", err)
		return nil, err
	}

	// create the client
	res, err := tx.Exec(
		createOAuth2ClientQuery,
		x.ClientID,
		x.ClientSecret,
		strings.Join(x.Scopes, scopesSeparator),
		x.RedirectURI,
		x.BelongsTo,
	)
	if err != nil {
		s.logger.Errorf("error executing client creation query: %v", err)
		tx.Rollback()
		return nil, err
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		s.logger.Errorf("error fetching last inserted item ID: %v", err)
		return nil, err
	}

	// fetch full updated client
	row := tx.QueryRow(getOAuth2ClientQuery, id)
	if x, err = scanOAuth2Client(row); err != nil {
		s.logger.Errorf("error fetching newly created client %s: %v", x.ClientID, err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorf("error committing transaction: %v", err)
		return nil, err
	}

	s.logger.Debugln("returning from CreateOAuth2Client")
	return
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (s *Sqlite) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) (err error) {
	tx, err := s.database.Begin()
	if err != nil {
		return
	}

	// update the client
	if _, err = tx.Exec(
		updateOAuth2ClientQuery,
		input.ClientID,
		input.ClientSecret,
		strings.Join(input.Scopes, scopesSeparator),
		input.RedirectURI,
		input.ID,
	); err != nil {
		tx.Rollback()
		return
	}

	// fetch full updated client
	row := tx.QueryRow(getOAuth2ClientQuery, input.ID)
	if input, err = scanOAuth2Client(row); err != nil {
		tx.Rollback()
		return
	}

	// commit the changes
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return
	}

	return
}

// DeleteOAuth2Client deletes an OAuth2 client
func (s *Sqlite) DeleteOAuth2Client(ctx context.Context, id string) error {
	_, err := s.database.Exec(archiveOAuth2ClientQuery, id)
	return err
}

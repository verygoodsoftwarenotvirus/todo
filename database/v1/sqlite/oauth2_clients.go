package sqlite

import (
	"math"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	scopesSeparator           = `,`
	getOauth2ClientCountQuery = `
		SELECT
			COUNT(*)
		FROM
			oauth_clients
		WHERE archived_on is null
	`
	getOauth2ClientQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			id = ?
	`
	getOauth2ClientByClientIDQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			client_id = ?
	`
	getOauth2ClientsQuery = `
		SELECT
			id, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to
		FROM
			oauth_clients
		WHERE
			archived_on is null
		LIMIT ?
		OFFSET ?
	`
	createOauth2ClientQuery = `
		INSERT INTO oauth_clients
		(
			client_id, client_secret, scopes, redirect_uri, belongs_to
		)
		VALUES
		(
			?, ?, ?, ?, ?
		)
	`
	updateOauth2ClientQuery = `
		UPDATE oauth_clients SET
			client_id = ?,
			client_secret = ?,
			scopes = ?,
			redirect_uri = ?,
			updated_on = (strftime('%s','now'))
		WHERE id = ?
	`
	archiveOauth2ClientQuery = `
		UPDATE oauth_clients SET
			updated_on = (strftime('%s','now')),
			archived_on = (strftime('%s','now'))
		WHERE id = ?
	`
)

func scanOauth2Client(scan database.Scannable) (*models.Oauth2Client, error) {
	var (
		x      = &models.Oauth2Client{}
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

var _ models.Oauth2ClientHandler = (*sqlite)(nil)

func (s *sqlite) GetOauth2Client(clientID string) (*models.Oauth2Client, error) {
	s.logger.Debugf("GetOauth2Client called for %s", clientID)
	row := s.database.QueryRow(getOauth2ClientByClientIDQuery, clientID)
	return scanOauth2Client(row)
}

func (s *sqlite) GetOauth2ClientCount(filter *models.QueryFilter) (uint64, error) {
	var count uint64
	err := s.database.QueryRow(getOauth2ClientCountQuery).Scan(&count)
	return count, err
}

func (s *sqlite) GetOauth2Clients(filter *models.QueryFilter) (*models.Oauth2ClientList, error) {
	if filter == nil {
		s.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.Page = uint64(math.Max(1, float64(filter.Page)))
	queryPage := uint(filter.Limit * (filter.Page - 1))

	list := []models.Oauth2Client{}

	s.logger.Debugf("query limit: %d, query page: %d, calculated page: %d", filter.Limit, filter.Page, queryPage)

	rows, err := s.database.Query(getOauth2ClientsQuery, filter.Limit, queryPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		x, err := scanOauth2Client(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *x)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ocl := &models.Oauth2ClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: 666,
		},
		Clients: list,
	}
	if ocl.TotalCount, err = s.GetOauth2ClientCount(filter); err != nil {
		return nil, err
	}

	return ocl, err
}

func (s *sqlite) CreateOauth2Client(input *models.Oauth2ClientCreationInput) (x *models.Oauth2Client, err error) {
	s.logger.Debugln("CreateOauth2Client called.")

	x = &models.Oauth2Client{
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
		createOauth2ClientQuery,
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
	row := tx.QueryRow(getOauth2ClientQuery, id)
	if x, err = scanOauth2Client(row); err != nil {
		s.logger.Errorf("error fetching newly created client %s: %v", x.ClientID, err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorf("error committing transaction: %v", err)
		return nil, err
	}

	s.logger.Debugln("returning from CreateOauth2Client")
	return
}

func (s *sqlite) UpdateOauth2Client(input *models.Oauth2Client) (err error) {
	tx, err := s.database.Begin()
	if err != nil {
		return
	}

	// update the client
	if _, err = tx.Exec(
		updateOauth2ClientQuery,
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
	row := tx.QueryRow(getOauth2ClientQuery, input.ID)
	if input, err = scanOauth2Client(row); err != nil {
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

func (s *sqlite) DeleteOauth2Client(id string) error {
	_, err := s.database.Exec(archiveOauth2ClientQuery, id)
	return err
}

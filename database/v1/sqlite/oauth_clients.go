package sqlite

import (
	"crypto/rand"
	"encoding/base64"
	"math"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	scopesSeparator          = `,`
	getOauthClientCountQuery = `
		SELECT
			COUNT(*)
		FROM
			oauth_clients
		WHERE archived_on is null
	`
	getOauthClientQuery = `
		SELECT
			id, client_id, scopes, client_secret, created_on, updated_on, archived_on
		FROM
			oauth_clients
		WHERE
			id = ? AND archived_on is null
	`
	getOauthClientByClientIDQuery = `
		SELECT
			id, client_id, scopes, client_secret, created_on, updated_on, archived_on
		FROM
			oauth_clients
		WHERE
			client_id = ? AND archived_on is null
	`
	getOauthClientsQuery = `
		SELECT
			id, client_id, scopes, client_secret, created_on, updated_on, archived_on
		FROM
			oauth_clients
		WHERE
			archived_on is null
		LIMIT ?
		OFFSET ?
	`
	createOauthClientQuery = `
		INSERT INTO oauth_clients
		(
			client_id, client_secret, scopes
		)
		VALUES
		(
			?, ?, ?
		)
	`
	updateOauthClientQuery = `
		UPDATE oauth_clients SET
			client_id = ?,
			client_secret = ?,
			scopes = ?,
			updated_on = (strftime('%s','now'))
		WHERE id = ?
	`
	archiveOauthClientQuery = `
		UPDATE oauth_clients SET
			client_id = "__ARCHIVED__",
			client_secret = "__ARCHIVED__",
			updated_on = (strftime('%s','now')),
			archived_on = (strftime('%s','now'))
		WHERE id = ?
	`
)

// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func buildRandoString() (string, error) {
	b := make([]byte, 64)
	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func scanOauthClient(scan database.Scannable) (*models.OauthClient, error) {
	var (
		x      = &models.OauthClient{}
		scopes string
	)
	err := scan.Scan(
		&x.ID,
		&x.ClientID,
		&scopes,
		&x.ClientSecret,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
	)
	if err != nil {
		return nil, err
	}

	x.Scopes = strings.Split(scopes, scopesSeparator)

	return x, nil
}

var _ models.OauthClientHandler = (*sqlite)(nil)

func (s *sqlite) GetOauthClient(id string) (*models.OauthClient, error) {
	row := s.database.QueryRow(getOauthClientByClientIDQuery, id)
	return scanOauthClient(row)
}

func (s *sqlite) GetOauthClientCount() (uint64, error) {
	var count uint64
	err := s.database.QueryRow(getOauthClientCountQuery).Scan(&count)
	return count, err
}

func (s *sqlite) GetOauthClients(filter *models.QueryFilter) (*models.OauthClientList, error) {
	if filter == nil {
		s.logger.Debugln("using default query filter")
		filter = models.DefaultQueryFilter
	}
	filter.Page = uint64(math.Max(1, float64(filter.Page)))
	queryPage := uint(filter.Limit * (filter.Page - 1))

	list := []models.OauthClient{}

	s.logger.Debugf("query limit: %d, query page: %d, calculated page: %d", filter.Limit, filter.Page, queryPage)

	rows, err := s.database.Query(getOauthClientsQuery, filter.Limit, queryPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		x, err := scanOauthClient(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *x)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ocl := &models.OauthClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: 666,
		},
		Clients: list,
	}
	if ocl.TotalCount, err = s.GetOauthClientCount(); err != nil {
		return nil, err
	}

	return ocl, err
}

func (s *sqlite) CreateOauthClient(input *models.OauthClientInput) (x *models.OauthClient, err error) {
	x = &models.OauthClient{
		Scopes: input.Scopes,
	}

	if x.ClientID, err = buildRandoString(); err != nil {
		return nil, err
	}

	if x.ClientSecret, err = buildRandoString(); err != nil {
		return nil, err
	}

	tx, err := s.database.Begin()
	if err != nil {
		s.logger.Errorf("error beginning database connection: %v", err)
		return nil, err
	}

	// create the client
	res, err := tx.Exec(
		createOauthClientQuery,
		x.ClientID,
		x.ClientSecret,
		strings.Join(x.Scopes, scopesSeparator),
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
	row := tx.QueryRow(getOauthClientQuery, id)
	if x, err = scanOauthClient(row); err != nil {
		s.logger.Errorf("error fetching newly created client %s: %v", x.ClientID, err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorf("error committing transaction: %v", err)
		return nil, err
	}

	s.logger.Debugln("returning from CreateOauthClient")
	return
}

func (s *sqlite) UpdateOauthClient(input *models.OauthClient) (err error) {
	tx, err := s.database.Begin()
	if err != nil {
		return
	}

	// update the client
	if _, err = tx.Exec(
		updateOauthClientQuery,
		input.ClientID,
		input.ClientSecret,
		strings.Join(input.Scopes, scopesSeparator),
		input.ClientID,
		input.ID,
	); err != nil {
		tx.Rollback()
		return
	}

	// fetch full updated client
	row := tx.QueryRow(getOauthClientQuery, input.ID)
	if input, err = scanOauthClient(row); err != nil {
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

func (s *sqlite) DeleteOauthClient(id uint) error {
	_, err := s.database.Exec(archiveOauthClientQuery, id)
	return err
}

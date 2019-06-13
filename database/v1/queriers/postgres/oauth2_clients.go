package postgres

import (
	"context"
	"database/sql"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
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

func scanOAuth2Client(scan database.Scanner) (*models.OAuth2Client, error) {
	var (
		x      = &models.OAuth2Client{}
		scopes string
	)

	err := scan.Scan(
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
	)
	if err != nil {
		return nil, err
	}
	x.Scopes = strings.Split(scopes, scopesSeparator)

	return x, nil
}

func scanOAuth2Clients(logger logging.Logger, rows *sql.Rows) ([]*models.OAuth2Client, error) {
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

	logQueryBuildingError(logger, rows.Close())

	return list, nil
}

func (p *Postgres) buildCreateOAuth2ClientQuery(input *models.OAuth2Client) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
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
		Suffix("RETURNING id, created_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// CreateOAuth2Client creates an OAuth2 client
func (p *Postgres) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	x := &models.OAuth2Client{
		Name:         input.ClientName,
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		RedirectURI:  input.RedirectURI,
		Scopes:       input.Scopes,
		BelongsTo:    input.BelongsTo,
	}
	query, args := p.buildCreateOAuth2ClientQuery(x)

	err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, errors.Wrap(err, "error executing client creation query")
	}

	return x, nil
}

func (p *Postgres) buildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	var err error
	// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
	// care about ownership. It does still care about archived status
	query, args, err = p.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			"client_id":   clientID,
			"archived_on": nil,
		}).ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetOAuth2ClientByClientID gets an OAuth2 client
func (p *Postgres) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	query, args := p.buildGetOAuth2ClientByClientIDQuery(clientID)
	row := p.db.QueryRowContext(ctx, query, args...)
	return scanOAuth2Client(row)
}

func (p *Postgres) buildGetAllOAuth2ClientsQuery() (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{"archived_on": nil}).
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetAllOAuth2Clients gets a list of OAuth2 clients regardless of ownership
func (p *Postgres) GetAllOAuth2Clients(ctx context.Context) ([]*models.OAuth2Client, error) {
	query, args := p.buildGetAllOAuth2ClientsQuery()
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	list, err := scanOAuth2Clients(p.logger, rows)
	if err != nil {
		return nil, errors.Wrap(err, "fetching list of OAuth2Clients")
	}

	return list, nil
}

func (p *Postgres) buildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Select(oauth2ClientsTableColumns...).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{
			"id":          clientID,
			"belongs_to":  userID,
			"archived_on": nil,
		}).ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetOAuth2Client gets an OAuth2 client
func (p *Postgres) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	query, args := p.buildGetOAuth2ClientQuery(clientID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)
	client, err := scanOAuth2Client(row)

	if err != nil {
		return nil, err
	}
	return client, nil
}

func (p *Postgres) buildGetOAuth2ClientCountQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
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
	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetOAuth2ClientCount will get the count of OAuth2 clients that match the current filter
func (p *Postgres) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	query, args := p.buildGetOAuth2ClientCountQuery(filter, userID)
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

func (p *Postgres) buildGetAllOAuth2ClientCountQuery() string {
	query, _, err := p.sqlBuilder.Select(CountQuery).
		From(oauth2ClientsTableName).
		Where(squirrel.Eq{"archived_on": nil}).
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query
}

// GetAllOAuth2ClientCount will get the count of OAuth2 clients that match the current filter
func (p *Postgres) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := p.db.QueryRowContext(ctx, p.buildGetAllOAuth2ClientCountQuery()).Scan(&count)
	return count, err
}

func (p *Postgres) buildGetOAuth2ClientsQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
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

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetOAuth2Clients gets a list of OAuth2 clients
func (p *Postgres) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	query, args := p.buildGetOAuth2ClientsQuery(filter, userID)
	rows, err := p.db.QueryContext(ctx, query, args...)

	logger := p.logger.WithValue("user_id", userID)

	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, errors.Wrap(err, "executing query")
	}

	list, err := scanOAuth2Clients(p.logger, rows)
	if err != nil {
		return nil, errors.Wrap(err, "scanning results")
	}

	// de-pointer-ize clients
	var tmpL = []models.OAuth2Client{}
	for _, t := range list {
		tmpL = append(tmpL, *t)
	}

	totalCount, err := p.GetOAuth2ClientCount(ctx, filter, userID)
	if err != nil {
		logger.Error(err, "error fetching client count for user")
		return nil, err
	}

	ocl := &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: totalCount,
		},
		Clients: tmpL,
	}

	return ocl, nil
}

func (p *Postgres) buildUpdateOAuth2ClientQuery(input *models.OAuth2Client) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Update(oauth2ClientsTableName).
		Set("client_id", input.ClientID).
		Set("client_secret", input.ClientSecret).
		Set("scopes", strings.Join(input.Scopes, scopesSeparator)).
		Set("redirect_uri", input.RedirectURI).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{
			"id":         input.ID,
			"belongs_to": input.BelongsTo,
		}).
		Suffix("RETURNING updated_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// UpdateOAuth2Client updates a OAuth2 client. Note that this function expects the input's
// ID field to be valid.
func (p *Postgres) UpdateOAuth2Client(ctx context.Context, input *models.OAuth2Client) error {
	query, args := p.buildUpdateOAuth2ClientQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.UpdatedOn)
}

func (p *Postgres) buildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Update(oauth2ClientsTableName).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Set("archived_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{
			"id":         clientID,
			"belongs_to": userID,
		}).
		Suffix("RETURNING archived_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// DeleteOAuth2Client deletes an OAuth2 client
func (p *Postgres) DeleteOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	query, args := p.buildArchiveOAuth2ClientQuery(clientID, userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

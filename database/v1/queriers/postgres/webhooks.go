package postgres

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

func (p Postgres) scanWebhook(scan database.Scanner) (*models.Webhook, error) {
	var (
		x = &models.Webhook{}
	)

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.ContentType,
		&x.URL,
		&x.Method,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.CompletedOn,
		&x.BelongsTo,
	); err != nil {
		return nil, err
	}

	return x, nil
}

const getWebhookQuery = `
	SELECT
		id,
		name,
		content_type,
		url,
		method,
		created_on,
		updated_on,
		completed_on,
		belongs_to
	FROM
		webhooks
	WHERE
		id = $1
		AND belongs_to = $2
`

// GetWebhook fetches an webhook from the postgres database
func (p *Postgres) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	row := p.database.QueryRowContext(ctx, getWebhookQuery, webhookID, userID)
	i, err := p.scanWebhook(row)
	return i, err
}

const getWebhookCountQuery = `
	SELECT
		COUNT(*)
	FROM
		webhooks
	WHERE
		completed_on IS NULL
		AND belongs_to = $1
` // FINISHME: finish adding filters to this query

// GetWebhookCount will fetch the count of webhooks from the postgres database that meet a particular filter and belong to a particular user.
func (p *Postgres) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	err = p.database.QueryRowContext(ctx, getWebhookCountQuery, userID).Scan(&count)
	return
}

const getAllWebhooksCountQuery = `
	SELECT
		COUNT(*)
	FROM
		webhooks
	WHERE
		completed_on IS NULL
` // FINISHME: finish adding filters to this query

// GetAllWebhooksCount will fetch the count of webhooks from the postgres database that meet a particular filter
func (p *Postgres) GetAllWebhooksCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	err = p.database.QueryRowContext(ctx, getAllWebhooksCountQuery).Scan(&count)
	return
}

const getWebhooksQuery = `
	SELECT
		id,
		name,
		content_type,
		url,
		method,
		created_on,
		updated_on,
		completed_on,
		belongs_to
	FROM
		webhooks
	WHERE
		completed_on IS NULL
		AND belongs_to = $1
	LIMIT $2
	OFFSET $3
` // FINISHME: finish adding filters to this query

// GetWebhooks fetches a list of webhooks from the postgres database that meet a particular filter
func (p *Postgres) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	var list []models.Webhook
	rows, err := p.database.QueryContext(ctx, getWebhooksQuery, userID, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		webhook, ierr := p.scanWebhook(rows)
		if ierr != nil {
			return nil, ierr
		}
		list = append(list, *webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	count, err := p.GetWebhookCount(ctx, filter, userID)
	if err != nil {
		return nil, err
	}

	x := &models.WebhookList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			TotalCount: count,
			Limit:      filter.Limit,
		},
		Webhooks: list,
	}

	return x, err
}

const createWebhookQuery = `
	INSERT INTO webhooks
	(
		name,
		content_type,
		url,
		method,
		belongs_to
	)
	VALUES
	(
		$1, $2, $3
	)
	RETURNING
		id,
		created_on
`

// CreateWebhook creates an webhook in a postgres database
func (p *Postgres) CreateWebhook(ctx context.Context, input *models.WebhookInput) (*models.Webhook, error) {
	i := &models.Webhook{
		BelongsTo: input.BelongsTo,
	}

	// create the webhook
	if err := p.database.
		QueryRow(
			createWebhookQuery,
			input.Name,
			input.ContentType,
			input.URL,
			input.Method,
			input.BelongsTo,
		).Scan(&i.ID, &i.CreatedOn); err != nil {
		return nil, errors.Wrap(err, "error executing webhook creation query")
	}

	return i, nil
}

const updateWebhookQuery = `
	UPDATE webhooks SET
		name = $1,
		content_type = $2,
		url = $3,
		method = $4,
		updated_on = extract(epoch FROM NOW())
	WHERE
		id = $5
		AND belongs_to = $6
	RETURNING
		updated_on
`

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (p *Postgres) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	// update the webhook
	err := p.database.
		QueryRowContext(
			ctx,
			updateWebhookQuery,
			input.Name,
			input.ContentType,
			input.URL,
			input.Method,
			input.ID,
			input.BelongsTo,
		).Scan(&input.UpdatedOn)
	return err
}

const archiveWebhookQuery = `
	UPDATE webhooks SET
		updated_on = extract(epoch FROM NOW()),
		completed_on = extract(epoch FROM NOW())
	WHERE
		id = $1
		AND completed_on IS NULL
		AND belongs_to = $2
	RETURNING
		completed_on
`

// DeleteWebhook deletes an webhook from the database by its ID
func (p *Postgres) DeleteWebhook(ctx context.Context, webhookID uint64, userID uint64) error {
	_, err := p.database.ExecContext(ctx, archiveWebhookQuery, webhookID, userID)
	return err
}

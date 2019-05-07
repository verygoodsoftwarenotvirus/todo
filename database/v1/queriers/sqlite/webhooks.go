package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

func scanWebhook(scan database.Scanner) (*models.Webhook, error) {
	x := &models.Webhook{}

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

func (s *Sqlite) scanWebhooks(rows *sql.Rows) ([]models.Webhook, error) {
	var list []models.Webhook

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		webhook, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *webhook)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
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
		id = ?
		AND belongs_to = ?
`

// GetWebhook fetches an webhook from the sqlite database
func (s *Sqlite) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	row := s.database.QueryRowContext(ctx, getWebhookQuery, webhookID, userID)
	i, err := scanWebhook(row)
	return i, err
}

const getWebhookCountQuery = `
	SELECT
		COUNT(*)
	FROM
		webhooks
	WHERE
		completed_on IS NULL
		AND belongs_to = ?
`

// GetWebhookCount fetches the count of webhooks from the sqlite database that meet a particular filter and belong to a particular user
func (s *Sqlite) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	var count uint64
	err := s.database.QueryRowContext(ctx, getWebhookCountQuery, userID).Scan(&count)
	return count, err
}

const getAllWebhooksCountQuery = `
	SELECT
		COUNT(*)
	FROM
		webhooks
	WHERE
		completed_on IS NULL
`

// GetAllWebhooksCount fetches the count of webhooks from the sqlite database that meet a particular filter
func (s *Sqlite) GetAllWebhooksCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	var count uint64
	err := s.database.QueryRowContext(ctx, getWebhookCountQuery).Scan(&count)
	return count, err
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
	LIMIT ?
	OFFSET ?
`

// GetWebhooks fetches a list of webhooks from the sqlite database that meet a particular filter
func (s *Sqlite) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	rows, err := s.database.QueryContext(ctx, getWebhooksQuery, filter.Limit, filter.QueryPage())
	if err != nil {
		return nil, errors.Wrap(err, "querying database for webhooks")
	}

	list, err := s.scanWebhooks(rows)
	if err != nil {
		return nil, errors.Wrap(err, "scanning webhooks")
	}

	count, err := s.GetWebhookCount(ctx, filter, userID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching webhook count")
	}

	x := &models.WebhookList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			TotalCount: count,
			Limit:      filter.Limit,
		},
		Webhooks: list,
	}

	return x, nil
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
		?, ?, ?, ?, ?
	)
`

// CreateWebhook creates an webhook in a sqlite database
func (s *Sqlite) CreateWebhook(ctx context.Context, input *models.WebhookInput) (*models.Webhook, error) {
	// create the webhook
	res, err := s.database.ExecContext(
		ctx,
		createWebhookQuery,
		input.Name,
		input.ContentType,
		input.URL,
		input.Method,
		input.BelongsTo,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error executing webhook creation query")
	}

	// determine its id
	id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching last inserted webhook ID")
	}

	// fetch full updated webhook
	i, err := s.GetWebhook(ctx, uint64(id), input.BelongsTo)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching newly created webhook")
	}

	return i, nil
}

const updateWebhookQuery = `
	UPDATE webhooks SET
		name = ?,
		content_type = ?,
		url = ?,
		method = ?,
		updated_on = (strftime('%s','now'))
	WHERE
		id = ?
		AND belongs_to = ?
`

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (s *Sqlite) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	_, err := s.database.ExecContext(
		ctx,
		updateWebhookQuery,
		input.Name,
		input.ContentType,
		input.URL,
		input.Method,
		input.ID,
		input.BelongsTo,
	)
	if err != nil {
		return errors.Wrap(err, "updating webhook")
	}

	return nil
}

const archiveWebhookQuery = `
	UPDATE webhooks SET
		updated_on = (strftime('%s','now')),
		completed_on = (strftime('%s','now'))
	WHERE
		id = ?
		AND belongs_to = ?
		AND completed_on IS NULL
`

// DeleteWebhook deletes an webhook from the database by its ID
func (s *Sqlite) DeleteWebhook(ctx context.Context, id uint64, userID uint64) error {
	_, err := s.database.ExecContext(ctx, archiveWebhookQuery, id, userID)
	return err
}

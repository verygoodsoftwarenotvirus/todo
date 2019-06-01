package postgres

import (
	"context"
	"database/sql"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

const (
	eventsSeparator = `,`
	typesSeparator  = `,`
	topicsSeparator = `,`

	webhooksTableName = "webhooks"
)

var (
	webhooksTableColumns = []string{
		"id",
		"name",
		"content_type",
		"url",
		"method",
		"events",
		"data_types",
		"topics",
		"created_on",
		"updated_on",
		"archived_on",
		"belongs_to",
	}
)

func (p Postgres) scanWebhook(scan database.Scanner) (*models.Webhook, error) {
	var (
		x = &models.Webhook{}

		eventsStr,
		dataTypesStr,
		topicsStr string
	)

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.ContentType,
		&x.URL,
		&x.Method,
		&eventsStr,
		&dataTypesStr,
		&topicsStr,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsTo,
	); err != nil {
		return nil, err
	}

	x.Events = strings.Split(eventsStr, eventsSeparator)
	x.DataTypes = strings.Split(dataTypesStr, typesSeparator)
	x.Topics = strings.Split(topicsStr, topicsSeparator)

	return x, nil
}

func (p *Postgres) scanWebhooks(rows *sql.Rows) ([]models.Webhook, error) {
	var list []models.Webhook

	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		webhook, err := p.scanWebhook(rows)
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
		events,
		data_types,
		topics,
		created_on,
		updated_on,
		archived_on,
		belongs_to
	FROM
		webhooks
	WHERE
		id = $1
		AND belongs_to = $2
`

// GetWebhook fetches an webhook from the postgres db
func (p *Postgres) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	row := p.db.QueryRowContext(ctx, getWebhookQuery, webhookID, userID)
	i, err := p.scanWebhook(row)
	return i, err
}

// GetWebhookCount will fetch the count of webhooks from the postgres db that meet a particular filter and belong to a particular user.
func (p *Postgres) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	builder := p.sqlBuilder.
		Select("COUNT(*)").
		From(webhooksTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":  userID,
			"archived_on": nil,
		}))

	builder = filter.ApplyToQueryBuilder(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "generating query")
	}

	err = p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return
}

const getAllWebhooksCountQuery = `
	SELECT
		COUNT(*)
	FROM
		webhooks
	WHERE
		archived_on IS NULL
`

// GetAllWebhooksCount will fetch the count of webhooks from the postgres db that meet a particular filter
func (p *Postgres) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = p.db.QueryRowContext(ctx, getAllWebhooksCountQuery).Scan(&count)
	return
}

const getAllWebhooksQuery = `
	SELECT
		id,
		name,
		content_type,
		url,
		method,
		events,
		data_types,
		topics,
		created_on,
		updated_on,
		archived_on,
		belongs_to
	FROM
		webhooks
	WHERE
		archived_on IS NULL
`

// GetAllWebhooks fetches a list of all webhooks from the postgres db
func (p *Postgres) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	var list []models.Webhook
	rows, err := p.db.QueryContext(ctx, getAllWebhooksQuery)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			p.logger.Error(err, "closing rows")
		}
	}()

	for rows.Next() {
		var webhook *models.Webhook
		webhook, err = p.scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	count, err := p.GetAllWebhooksCount(ctx)
	if err != nil {
		return nil, err
	}

	x := &models.WebhookList{
		Pagination: models.Pagination{
			TotalCount: count,
		},
		Webhooks: list,
	}

	return x, err
}

// GetWebhooks fetches a list of webhooks from the postgres db that meet a particular filter
func (p *Postgres) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	builder := p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":  userID,
			"archived_on": nil,
		}))

	builder = filter.ApplyToQueryBuilder(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "generating query")
	}

	rows, err := p.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if e := rows.Close(); e != nil {
			p.logger.Error(e, "closing rows")
		}
	}()

	list, err := p.scanWebhooks(rows)
	if err != nil {
		return nil, errors.Wrap(err, "scanning webhooks")
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
		events,
		data_types,
		topics,
		belongs_to
	)
	VALUES
	(
		$1, $2, $3, $4, $5, $6, $7, $8
	)
	RETURNING
		id,
		created_on
`

// CreateWebhook creates an webhook in a postgres db
func (p *Postgres) CreateWebhook(ctx context.Context, input *models.WebhookInput) (*models.Webhook, error) {
	x := &models.Webhook{
		Name:        input.Name,
		ContentType: input.ContentType,
		URL:         input.URL,
		Method:      input.Method,
		Events:      input.Events,
		DataTypes:   input.DataTypes,
		Topics:      input.Topics,
		BelongsTo:   input.BelongsTo,
	}

	// create the webhook
	if err := p.db.
		QueryRow(
			createWebhookQuery,
			input.Name,
			input.ContentType,
			input.URL,
			input.Method,
			strings.Join(input.Events, eventsSeparator),
			strings.Join(input.DataTypes, typesSeparator),
			strings.Join(input.Topics, topicsSeparator),
			input.BelongsTo,
		).Scan(&x.ID, &x.CreatedOn); err != nil {
		return nil, errors.Wrap(err, "error executing webhook creation query")
	}

	return x, nil
}

const updateWebhookQuery = `
	UPDATE webhooks SET
		name = $1,
		content_type = $2,
		url = $3,
		method = $4,
		events = $5,
		data_types = $6,
		topics = $7,
		updated_on = extract(epoch FROM NOW())
	WHERE
		id = $8
		AND belongs_to = $9
	RETURNING
		updated_on
`

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (p *Postgres) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	// update the webhook
	err := p.db.
		QueryRowContext(
			ctx,
			updateWebhookQuery,
			input.Name,
			input.ContentType,
			input.URL,
			input.Method,
			strings.Join(input.Events, eventsSeparator),
			strings.Join(input.DataTypes, typesSeparator),
			strings.Join(input.Topics, topicsSeparator),
			input.ID,
			input.BelongsTo,
		).Scan(&input.UpdatedOn)
	return err
}

const archiveWebhookQuery = `
	UPDATE webhooks SET
		updated_on = extract(epoch FROM NOW()),
		archived_on = extract(epoch FROM NOW())
	WHERE
		id = $1
		AND archived_on IS NULL
		AND belongs_to = $2
	RETURNING
		archived_on
`

// DeleteWebhook deletes an webhook from the db by its ID
func (p *Postgres) DeleteWebhook(ctx context.Context, webhookID uint64, userID uint64) error {
	_, err := p.db.ExecContext(ctx, archiveWebhookQuery, webhookID, userID)
	return err
}

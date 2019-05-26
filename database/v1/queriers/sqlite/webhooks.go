package sqlite

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

func scanWebhook(scan database.Scanner) (*models.Webhook, error) {
	var (
		x = &models.Webhook{}

		eventsStr,
		typesStr,
		topicsStr string
	)

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.ContentType,
		&x.URL,
		&x.Method,
		&eventsStr,
		&typesStr,
		&topicsStr,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsTo,
	); err != nil {
		return nil, err
	}

	x.Events = strings.Split(eventsStr, eventsSeparator)
	x.DataTypes = strings.Split(typesStr, typesSeparator)
	x.Topics = strings.Split(topicsStr, topicsSeparator)

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
		id = ?
		AND belongs_to = ?
`

// GetWebhook fetches an webhook from the sqlite database
func (s *Sqlite) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	row := s.database.QueryRowContext(ctx, getWebhookQuery, webhookID, userID)
	i, err := scanWebhook(row)
	return i, err
}

// GetWebhookCount fetches the count of webhooks from the sqlite database that meet a particular filter and belong to a particular user
func (s *Sqlite) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	builder := s.sqlBuilder.
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

	err = s.database.QueryRowContext(ctx, query, args...).Scan(&count)
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

// GetAllWebhooksCount fetches the count of webhooks from the sqlite database that meet a particular filter
func (s *Sqlite) GetAllWebhooksCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := s.database.QueryRowContext(ctx, getAllWebhooksCountQuery).Scan(&count)
	return count, err
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

// GetAllWebhooks fetches a list of webhooks from the sqlite database that meet a particular filter
func (s *Sqlite) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	rows, err := s.database.QueryContext(ctx, getAllWebhooksQuery)
	if err != nil {
		return nil, errors.Wrap(err, "querying database for webhooks")
	}

	list, err := s.scanWebhooks(rows)
	if err != nil {
		return nil, errors.Wrap(err, "scanning webhooks")
	}

	count, err := s.GetAllWebhooksCount(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetching webhook count")
	}

	x := &models.WebhookList{
		Pagination: models.Pagination{
			TotalCount: count,
		},
		Webhooks: list,
	}

	return x, nil
}

// GetWebhooks fetches a list of webhooks from the sqlite database that meet a particular filter
func (s *Sqlite) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	builder := s.sqlBuilder.
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

	rows, err := s.database.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if e := rows.Close(); e != nil {
			s.logger.Error(e, "closing rows")
		}
	}()

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
		events,
		data_types,
		topics,
		belongs_to
	)
	VALUES
	(
		?, ?, ?, ?, ?, ?, ?, ?
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
		strings.Join(input.Events, eventsSeparator),
		strings.Join(input.DataTypes, typesSeparator),
		strings.Join(input.Topics, topicsSeparator),
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
	x, err := s.GetWebhook(ctx, uint64(id), input.BelongsTo)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching newly created webhook")
	}

	return x, nil
}

const updateWebhookQuery = `
	UPDATE webhooks SET
		name = ?,
		content_type = ?,
		url = ?,
		method = ?,
		events = ?,
		data_types = ?,
		topics = ?,
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
		strings.Join(input.Events, eventsSeparator),
		strings.Join(input.DataTypes, typesSeparator),
		strings.Join(input.Topics, topicsSeparator),
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
		archived_on = (strftime('%s','now'))
	WHERE
		id = ?
		AND belongs_to = ?
		AND archived_on IS NULL
`

// DeleteWebhook deletes an webhook from the database by its ID
func (s *Sqlite) DeleteWebhook(ctx context.Context, id uint64, userID uint64) error {
	_, err := s.database.ExecContext(ctx, archiveWebhookQuery, id, userID)
	return err
}

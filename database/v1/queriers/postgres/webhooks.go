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

func scanWebhooks(logger logging.Logger, rows *sql.Rows) ([]models.Webhook, error) {
	var list []models.Webhook

	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(err, "closing rows")
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

func (p *Postgres) buildGetWebhookQuery(webhookID, userID uint64) (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			"id":         webhookID,
			"belongs_to": userID,
		}).ToSql()

	logQueryBuildingError(p.logger, err)
	return query, args
}

// GetWebhook fetches an webhook from the postgres db
func (p *Postgres) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	query, args := p.buildGetWebhookQuery(webhookID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)
	return scanWebhook(row)
}

func (p *Postgres) buildGetWebhookCountQuery(filter *models.QueryFilter, userID uint64) (string, []interface{}) {
	builder := p.sqlBuilder.
		Select("COUNT(*)").
		From(webhooksTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":  userID,
			"archived_on": nil,
		}))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err := builder.ToSql()
	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetWebhookCount will fetch the count of webhooks from the postgres db that meet a particular filter and belong to a particular user.
func (p *Postgres) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	query, args := p.buildGetWebhookCountQuery(filter, userID)
	return count, p.db.QueryRowContext(ctx, query, args...).Scan(&count)
}

func (p *Postgres) buildGetAllWebhooksCountQuery() (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Select("COUNT(*)").
		From(webhooksTableName).
		Where(squirrel.Eq{"archived_on": nil}).
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetAllWebhooksCount will fetch the count of webhooks from the postgres db that meet a particular filter
func (p *Postgres) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	query, args := p.buildGetAllWebhooksCountQuery()
	return count, p.db.QueryRowContext(ctx, query, args...).Scan(&count)
}

func (p *Postgres) buildGetAllWebhooksQuery() (string, []interface{}) {
	query, args, err := p.sqlBuilder.Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			"archived_on": nil,
		}).ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetAllWebhooks fetches a list of all webhooks from the postgres db
func (p *Postgres) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	query, args := p.buildGetAllWebhooksQuery()
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	list, err := scanWebhooks(p.logger, rows)
	if err != nil {
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

	list, err := scanWebhooks(p.logger, rows)
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

func (p *Postgres) buildWebhookCreationQuery(x *models.Webhook) (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Insert(webhooksTableName).
		Columns(
			"name",
			"content_type",
			"url",
			"method",
			"events",
			"data_types",
			"topics",
			"belongs_to",
		).
		Values(
			x.Name,
			x.ContentType,
			x.URL,
			x.Method,
			strings.Join(x.Events, eventsSeparator),
			strings.Join(x.DataTypes, typesSeparator),
			strings.Join(x.Topics, topicsSeparator),
			x.BelongsTo,
		).
		Suffix("RETURNING id, created_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

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

	query, args := p.buildWebhookCreationQuery(x)
	if err := p.db.QueryRow(query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		return nil, errors.Wrap(err, "error executing webhook creation query")
	}

	return x, nil
}

func (p *Postgres) buildUpdateWebhookQuery(input *models.Webhook) (string, []interface{}) {
	query, args, err := p.sqlBuilder.Update(webhooksTableName).
		Set("name", input.Name).
		Set("content_type", input.ContentType).
		Set("url", input.URL).
		Set("method", input.Method).
		Set("events", strings.Join(input.Events, topicsSeparator)).
		Set("data_types", strings.Join(input.DataTypes, typesSeparator)).
		Set("topics", strings.Join(input.Topics, topicsSeparator)).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{
			"id":         input.ID,
			"belongs_to": input.BelongsTo,
		}).Suffix("RETURNING updated_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (p *Postgres) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	query, args := p.buildUpdateWebhookQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.UpdatedOn)
}

func (p *Postgres) buildArchiveWebhookQuery(webhookID uint64, userID uint64) (string, []interface{}) {
	query, args, err := p.sqlBuilder.Update(webhooksTableName).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Set("archived_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{
			"id":          webhookID,
			"belongs_to":  userID,
			"archived_on": nil,
		}).
		Suffix("RETURNING archived_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// DeleteWebhook deletes an webhook from the db by its ID
func (p *Postgres) DeleteWebhook(ctx context.Context, webhookID uint64, userID uint64) error {
	query, args := p.buildArchiveWebhookQuery(webhookID, userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

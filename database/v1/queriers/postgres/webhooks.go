package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
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

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct
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

	if events := strings.Split(eventsStr, eventsSeparator); len(events) >= 1 && events[0] != "" {
		x.Events = events
	}
	if dataTypes := strings.Split(dataTypesStr, typesSeparator); len(dataTypes) >= 1 && dataTypes[0] != "" {
		x.DataTypes = dataTypes
	}
	if topics := strings.Split(topicsStr, topicsSeparator); len(topics) >= 1 && topics[0] != "" {
		x.Topics = topics
	}

	return x, nil
}

// scanWebhooks provides a consistent way to turn sql rows into a slice of webhooks
func scanWebhooks(logger logging.Logger, rows *sql.Rows) ([]models.Webhook, error) {
	var list []models.Webhook

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

	if err := rows.Close(); err != nil {
		logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook
func (p *Postgres) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			"id":         webhookID,
			"belongs_to": userID,
		}).ToSql()

	p.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database
func (p *Postgres) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	query, args := p.buildGetWebhookQuery(webhookID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	webhook, err := scanWebhook(row)
	if err != nil {
		return nil, buildError(err, "querying for webhook")
	}

	return webhook, nil
}

// buildGetWebhookCountQuery returns a SQL query (and arguments) that returns a list of webhooks
// meeting a given filter's criteria and belonging to a given user.
func (p *Postgres) buildGetWebhookCountQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
		Select(CountQuery).
		From(webhooksTableName).
		Where(squirrel.Eq{
			"belongs_to":  userID,
			"archived_on": nil,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetWebhookCount will fetch the count of webhooks from the database that meet a particular filter,
// and belong to a particular user.
func (p *Postgres) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	query, args := p.buildGetWebhookCountQuery(filter, userID)
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

var (
	getAllWebhooksCountQueryBuilder sync.Once
	getAllWebhooksCountQuery        string
)

// buildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (p *Postgres) buildGetAllWebhooksCountQuery() string {
	getAllWebhooksCountQueryBuilder.Do(func() {
		var err error
		getAllWebhooksCountQuery, _, err = p.sqlBuilder.
			Select(CountQuery).
			From(webhooksTableName).
			Where(squirrel.Eq{"archived_on": nil}).
			ToSql()

		p.logQueryBuildingError(err)
	})

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database
func (p *Postgres) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = p.db.QueryRowContext(ctx, p.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

var (
	getAllWebhooksQueryBuilder sync.Once
	getAllWebhooksQuery        string
)

// buildGetAllWebhooksQuery returns a SQL query which will return all webhooks, regardless of ownership
func (p *Postgres) buildGetAllWebhooksQuery() string {
	getAllWebhooksQueryBuilder.Do(func() {
		var err error
		getAllWebhooksQuery, _, err = p.sqlBuilder.
			Select(webhooksTableColumns...).
			From(webhooksTableName).
			Where(squirrel.Eq{"archived_on": nil}).
			ToSql()

		p.logQueryBuildingError(err)
	})

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database
func (p *Postgres) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	rows, err := p.db.QueryContext(ctx, p.buildGetAllWebhooksQuery())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying for webhooks: %w", err)
	}

	list, err := scanWebhooks(p.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	count, err := p.GetAllWebhooksCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching webhook count: %w", err)
	}

	x := &models.WebhookList{
		Pagination: models.Pagination{
			Page:       1,
			TotalCount: count,
		},
		Webhooks: list,
	}

	return x, err
}

// GetAllWebhooksForUser fetches a list of all webhooks from the database
func (p *Postgres) GetAllWebhooksForUser(ctx context.Context, userID uint64) ([]models.Webhook, error) {
	query, args := p.buildGetWebhooksQuery(nil, userID)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database for webhooks: %w", err)
	}

	list, err := scanWebhooks(p.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return list, nil
}

// buildGetWebhooksQuery returns a SQL query (and arguments) that would return a
func (p *Postgres) buildGetWebhooksQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			"belongs_to":  userID,
			"archived_on": nil,
		})

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter
func (p *Postgres) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	query, args := p.buildGetWebhooksQuery(filter, userID)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, err := scanWebhooks(p.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	count, err := p.GetWebhookCount(ctx, filter, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching count: %w", err)
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

// buildWebhookCreationQuery returns a SQL query (and arguments) that would create a given webhook
func (p *Postgres) buildWebhookCreationQuery(x *models.Webhook) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
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

	p.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database
func (p *Postgres) CreateWebhook(ctx context.Context, input *models.WebhookCreationInput) (*models.Webhook, error) {
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
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update
func (p *Postgres) buildUpdateWebhookQuery(input *models.Webhook) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Update(webhooksTableName).
		Set("name", input.Name).
		Set("content_type", input.ContentType).
		Set("url", input.URL).
		Set("method", input.Method).
		Set("events", strings.Join(input.Events, topicsSeparator)).
		Set("data_types", strings.Join(input.DataTypes, typesSeparator)).
		Set("topics", strings.Join(input.Topics, topicsSeparator)).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":         input.ID,
			"belongs_to": input.BelongsTo,
		}).Suffix("RETURNING updated_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (p *Postgres) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	query, args := p.buildUpdateWebhookQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.UpdatedOn)
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (p *Postgres) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = p.sqlBuilder.
		Update(webhooksTableName).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":          webhookID,
			"belongs_to":  userID,
			"archived_on": nil,
		}).Suffix("RETURNING archived_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// ArchiveWebhook archives a webhook from the database by its ID
func (p *Postgres) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	query, args := p.buildArchiveWebhookQuery(webhookID, userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

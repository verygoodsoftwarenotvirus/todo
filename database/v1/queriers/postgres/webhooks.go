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
)

const (
	eventsSeparator = `,`
	typesSeparator  = `,`
	topicsSeparator = `,`

	webhooksTableName            = "webhooks"
	webhooksTableOwnershipColumn = "belongs_to_user"
)

var (
	webhooksTableColumns = []string{
		fmt.Sprintf("%s.id", webhooksTableName),
		fmt.Sprintf("%s.name", webhooksTableName),
		fmt.Sprintf("%s.content_type", webhooksTableName),
		fmt.Sprintf("%s.url", webhooksTableName),
		fmt.Sprintf("%s.method", webhooksTableName),
		fmt.Sprintf("%s.events", webhooksTableName),
		fmt.Sprintf("%s.data_types", webhooksTableName),
		fmt.Sprintf("%s.topics", webhooksTableName),
		fmt.Sprintf("%s.created_on", webhooksTableName),
		fmt.Sprintf("%s.updated_on", webhooksTableName),
		fmt.Sprintf("%s.archived_on", webhooksTableName),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableOwnershipColumn),
	}
)

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct.
func (p *Postgres) scanWebhook(scan database.Scanner, includeCount bool) (*models.Webhook, uint64, error) {
	var (
		x     = &models.Webhook{}
		count uint64
		eventsStr,
		dataTypesStr,
		topicsStr string
	)

	targetVars := []interface{}{
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
		&x.BelongsToUser,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
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

	return x, count, nil
}

// scanWebhooks provides a consistent way to turn sql rows into a slice of webhooks.
func (p *Postgres) scanWebhooks(rows database.ResultIterator) ([]models.Webhook, uint64, error) {
	var (
		list  []models.Webhook
		count uint64
	)

	for rows.Next() {
		webhook, c, err := p.scanWebhook(rows, true)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 {
			count = c
		}

		list = append(list, *webhook)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := rows.Close(); err != nil {
		p.logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook
func (p *Postgres) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", webhooksTableName):                               webhookID,
			fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableOwnershipColumn): userID,
		}).ToSql()

	p.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database.
func (p *Postgres) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	query, args := p.buildGetWebhookQuery(webhookID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	webhook, _, err := p.scanWebhook(row, false)
	if err != nil {
		return nil, buildError(err, "querying for webhook")
	}

	return webhook, nil
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
			Select(fmt.Sprintf(countQuery, webhooksTableName)).
			From(webhooksTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", webhooksTableName): nil,
			}).
			ToSql()

		p.logQueryBuildingError(err)
	})

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database.
func (p *Postgres) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = p.db.QueryRowContext(ctx, p.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

var (
	getAllWebhooksQueryBuilder sync.Once
	getAllWebhooksQuery        string
)

// buildGetAllWebhooksQuery returns a SQL query which will return all webhooks, regardless of ownership.
func (p *Postgres) buildGetAllWebhooksQuery() string {
	getAllWebhooksQueryBuilder.Do(func() {
		var err error

		getAllWebhooksQuery, _, err = p.sqlBuilder.
			Select(webhooksTableColumns...).
			From(webhooksTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", webhooksTableName): nil,
			}).
			ToSql()

		p.logQueryBuildingError(err)
	})

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database.
func (p *Postgres) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	rows, err := p.db.QueryContext(ctx, p.buildGetAllWebhooksQuery())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying for webhooks: %w", err)
	}

	list, count, err := p.scanWebhooks(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
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

// buildGetWebhooksQuery returns a SQL query (and arguments) that would return a
func (p *Postgres) buildGetWebhooksQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := p.sqlBuilder.
		Select(append(webhooksTableColumns, fmt.Sprintf(countQuery, webhooksTableName))...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableOwnershipColumn): userID,
			fmt.Sprintf("%s.archived_on", webhooksTableName):                      nil,
		}).
		GroupBy(fmt.Sprintf("%s.id", webhooksTableName))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, webhooksTableName)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (p *Postgres) GetWebhooks(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.WebhookList, error) {
	query, args := p.buildGetWebhooksQuery(userID, filter)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, count, err := p.scanWebhooks(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
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
			webhooksTableOwnershipColumn,
		).
		Values(
			x.Name,
			x.ContentType,
			x.URL,
			x.Method,
			strings.Join(x.Events, eventsSeparator),
			strings.Join(x.DataTypes, typesSeparator),
			strings.Join(x.Topics, topicsSeparator),
			x.BelongsToUser,
		).
		Suffix("RETURNING id, created_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database.
func (p *Postgres) CreateWebhook(ctx context.Context, input *models.WebhookCreationInput) (*models.Webhook, error) {
	x := &models.Webhook{
		Name:          input.Name,
		ContentType:   input.ContentType,
		URL:           input.URL,
		Method:        input.Method,
		Events:        input.Events,
		DataTypes:     input.DataTypes,
		Topics:        input.Topics,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := p.buildWebhookCreationQuery(x)
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
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
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                         input.ID,
			webhooksTableOwnershipColumn: input.BelongsToUser,
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
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                         webhookID,
			webhooksTableOwnershipColumn: userID,
			"archived_on":                nil,
		}).Suffix("RETURNING archived_on").
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// ArchiveWebhook archives a webhook from the database by its ID.
func (p *Postgres) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	query, args := p.buildArchiveWebhookQuery(webhookID, userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

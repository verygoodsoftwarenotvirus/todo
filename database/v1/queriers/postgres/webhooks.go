package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
)

const (
	commaSeparator = ","

	eventsSeparator = commaSeparator
	typesSeparator  = commaSeparator
	topicsSeparator = commaSeparator

	webhooksTableName              = "webhooks"
	webhooksTableNameColumn        = "name"
	webhooksTableContentTypeColumn = "content_type"
	webhooksTableURLColumn         = "url"
	webhooksTableMethodColumn      = "method"
	webhooksTableEventsColumn      = "events"
	webhooksTableDataTypesColumn   = "data_types"
	webhooksTableTopicsColumn      = "topics"
	webhooksTableOwnershipColumn   = "belongs_to_user"
)

var (
	webhooksTableColumns = []string{
		fmt.Sprintf("%s.%s", webhooksTableName, idColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableNameColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableContentTypeColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableURLColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableMethodColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableEventsColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableDataTypesColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableTopicsColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, createdOnColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, lastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, archivedOnColumn),
		fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableOwnershipColumn),
	}
)

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct.
func (p *Postgres) scanWebhook(scan database.Scanner) (*models.Webhook, error) {
	var (
		x = &models.Webhook{}
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
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if err := scan.Scan(targetVars...); err != nil {
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

// scanWebhooks provides a consistent way to turn sql rows into a slice of webhooks.
func (p *Postgres) scanWebhooks(rows database.ResultIterator) ([]models.Webhook, error) {
	var list []models.Webhook

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

	if err := rows.Close(); err != nil {
		p.logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook.
func (p *Postgres) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", webhooksTableName, idColumn):                     webhookID,
			fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableOwnershipColumn): userID,
		}).ToSql()

	p.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database.
func (p *Postgres) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	query, args := p.buildGetWebhookQuery(webhookID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	webhook, err := p.scanWebhook(row)
	if err != nil {
		return nil, buildError(err, "querying for webhook")
	}

	return webhook, nil
}

// buildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (p *Postgres) buildGetAllWebhooksCountQuery() string {
	var err error

	getAllWebhooksCountQuery, _, err := p.sqlBuilder.
		Select(fmt.Sprintf(countQuery, webhooksTableName)).
		From(webhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", webhooksTableName, archivedOnColumn): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database.
func (p *Postgres) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = p.db.QueryRowContext(ctx, p.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

// buildGetAllWebhooksQuery returns a SQL query which will return all webhooks, regardless of ownership.
func (p *Postgres) buildGetAllWebhooksQuery() string {
	var err error

	getAllWebhooksQuery, _, err := p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", webhooksTableName, archivedOnColumn): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database.
func (p *Postgres) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	rows, err := p.db.QueryContext(ctx, p.buildGetAllWebhooksQuery())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("querying for webhooks: %w", err)
	}

	list, err := p.scanWebhooks(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	x := &models.WebhookList{
		Pagination: models.Pagination{
			Page: 1,
		},
		Webhooks: list,
	}

	return x, err
}

// buildGetWebhooksQuery returns a SQL query (and arguments) that would return a list of webhooks.
func (p *Postgres) buildGetWebhooksQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := p.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", webhooksTableName, archivedOnColumn):             nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", webhooksTableName, idColumn))

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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, err := p.scanWebhooks(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	x := &models.WebhookList{
		Pagination: models.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Webhooks: list,
	}

	return x, err
}

// buildCreateWebhookQuery returns a SQL query (and arguments) that would create a given webhook
func (p *Postgres) buildCreateWebhookQuery(x *models.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Insert(webhooksTableName).
		Columns(
			webhooksTableNameColumn,
			webhooksTableContentTypeColumn,
			webhooksTableURLColumn,
			webhooksTableMethodColumn,
			webhooksTableEventsColumn,
			webhooksTableDataTypesColumn,
			webhooksTableTopicsColumn,
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
		Suffix(fmt.Sprintf("RETURNING %s, %s", idColumn, createdOnColumn)).
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

	query, args := p.buildCreateWebhookQuery(x)
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
		Set(webhooksTableNameColumn, input.Name).
		Set(webhooksTableContentTypeColumn, input.ContentType).
		Set(webhooksTableURLColumn, input.URL).
		Set(webhooksTableMethodColumn, input.Method).
		Set(webhooksTableEventsColumn, strings.Join(input.Events, topicsSeparator)).
		Set(webhooksTableDataTypesColumn, strings.Join(input.DataTypes, typesSeparator)).
		Set(webhooksTableTopicsColumn, strings.Join(input.Topics, topicsSeparator)).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn:                     input.ID,
			webhooksTableOwnershipColumn: input.BelongsToUser,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", lastUpdatedOnColumn)).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (p *Postgres) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	query, args := p.buildUpdateWebhookQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (p *Postgres) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(webhooksTableName).
		Set(lastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(archivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			idColumn:                     webhookID,
			webhooksTableOwnershipColumn: userID,
			archivedOnColumn:             nil,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", archivedOnColumn)).
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

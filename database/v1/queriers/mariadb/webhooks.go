package mariadb

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

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct
func scanWebhook(scan database.Scanner, includeCount bool) (*models.Webhook, uint64, error) {
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

// scanWebhooks provides a consistent way to turn sql rows into a slice of webhooks
func scanWebhooks(logger logging.Logger, rows *sql.Rows) ([]models.Webhook, uint64, error) {
	var (
		list  []models.Webhook
		count uint64
	)

	for rows.Next() {
		webhook, c, err := scanWebhook(rows, true)
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
		logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook
func (m *MariaDB) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.id", webhooksTableName):                               webhookID,
			fmt.Sprintf("%s.%s", webhooksTableName, webhooksTableOwnershipColumn): userID,
		}).ToSql()

	m.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database
func (m *MariaDB) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	query, args := m.buildGetWebhookQuery(webhookID, userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	webhook, _, err := scanWebhook(row, false)
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
func (m *MariaDB) buildGetAllWebhooksCountQuery() string {
	getAllWebhooksCountQueryBuilder.Do(func() {
		var err error

		getAllWebhooksCountQuery, _, err = m.sqlBuilder.
			Select(fmt.Sprintf(countQuery, webhooksTableName)).
			From(webhooksTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", webhooksTableName): nil,
			}).
			ToSql()

		m.logQueryBuildingError(err)
	})

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database
func (m *MariaDB) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = m.db.QueryRowContext(ctx, m.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

var (
	getAllWebhooksQueryBuilder sync.Once
	getAllWebhooksQuery        string
)

// buildGetAllWebhooksQuery returns a SQL query which will return all webhooks, regardless of ownership
func (m *MariaDB) buildGetAllWebhooksQuery() string {
	getAllWebhooksQueryBuilder.Do(func() {
		var err error

		getAllWebhooksQuery, _, err = m.sqlBuilder.
			Select(webhooksTableColumns...).
			From(webhooksTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.archived_on", webhooksTableName): nil,
			}).
			ToSql()

		m.logQueryBuildingError(err)
	})

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database
func (m *MariaDB) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	rows, err := m.db.QueryContext(ctx, m.buildGetAllWebhooksQuery())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying for webhooks: %w", err)
	}

	list, count, err := scanWebhooks(m.logger, rows)
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
func (m *MariaDB) buildGetWebhooksQuery(userID uint64, filter *models.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
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
	m.logQueryBuildingError(err)

	return query, args
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter
func (m *MariaDB) GetWebhooks(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.WebhookList, error) {
	query, args := m.buildGetWebhooksQuery(userID, filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, count, err := scanWebhooks(m.logger, rows)
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
func (m *MariaDB) buildWebhookCreationQuery(x *models.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
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
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database
func (m *MariaDB) CreateWebhook(ctx context.Context, input *models.WebhookCreationInput) (*models.Webhook, error) {
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

	query, args := m.buildWebhookCreationQuery(x)
	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	// fetch the last inserted ID
	id, err := res.LastInsertId()
	m.logIDRetrievalError(err)
	x.ID = uint64(id)

	// this won't be completely accurate, but it will suffice
	x.CreatedOn = m.timeTeller.Now()

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update
func (m *MariaDB) buildUpdateWebhookQuery(input *models.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
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
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (m *MariaDB) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	query, args := m.buildUpdateWebhookQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (m *MariaDB) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(webhooksTableName).
		Set("updated_on", squirrel.Expr(currentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":                         webhookID,
			webhooksTableOwnershipColumn: userID,
			"archived_on":                nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ArchiveWebhook archives a webhook from the database by its ID
func (m *MariaDB) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	query, args := m.buildArchiveWebhookQuery(webhookID, userID)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

package sqlite

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
func (s *Sqlite) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Select(webhooksTableColumns...).
		From(webhooksTableName).
		Where(squirrel.Eq{
			"id":         webhookID,
			"belongs_to": userID,
		}).ToSql()

	s.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database
func (s *Sqlite) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	query, args := s.buildGetWebhookQuery(webhookID, userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	webhook, err := scanWebhook(row)
	if err != nil {
		return nil, buildError(err, "querying for webhook")
	}

	return webhook, nil
}

// buildGetWebhookCountQuery returns a SQL query (and arguments) that returns a list of webhooks
// meeting a given filter's criteria and belonging to a given user.
func (s *Sqlite) buildGetWebhookCountQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := s.sqlBuilder.
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
	s.logQueryBuildingError(err)

	return query, args
}

// GetWebhookCount will fetch the count of webhooks from the database that meet a particular filter,
// and belong to a particular user.
func (s *Sqlite) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	query, args := s.buildGetWebhookCountQuery(filter, userID)
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

var (
	getAllWebhooksCountQueryBuilder sync.Once
	getAllWebhooksCountQuery        string
)

// buildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (s *Sqlite) buildGetAllWebhooksCountQuery() string {
	getAllWebhooksCountQueryBuilder.Do(func() {
		var err error
		getAllWebhooksCountQuery, _, err = s.sqlBuilder.
			Select(CountQuery).
			From(webhooksTableName).
			Where(squirrel.Eq{"archived_on": nil}).
			ToSql()

		s.logQueryBuildingError(err)
	})

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database
func (s *Sqlite) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = s.db.QueryRowContext(ctx, s.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

var (
	getAllWebhooksQueryBuilder sync.Once
	getAllWebhooksQuery        string
)

// buildGetAllWebhooksQuery returns a SQL query which will return all webhooks, regardless of ownership
func (s *Sqlite) buildGetAllWebhooksQuery() string {
	getAllWebhooksQueryBuilder.Do(func() {
		var err error
		getAllWebhooksQuery, _, err = s.sqlBuilder.
			Select(webhooksTableColumns...).
			From(webhooksTableName).
			Where(squirrel.Eq{"archived_on": nil}).
			ToSql()

		s.logQueryBuildingError(err)
	})

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database
func (s *Sqlite) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	rows, err := s.db.QueryContext(ctx, s.buildGetAllWebhooksQuery())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying for webhooks: %w", err)
	}

	list, err := scanWebhooks(s.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	count, err := s.GetAllWebhooksCount(ctx)
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
func (s *Sqlite) GetAllWebhooksForUser(ctx context.Context, userID uint64) ([]models.Webhook, error) {
	query, args := s.buildGetWebhooksQuery(nil, userID)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database for webhooks: %w", err)
	}

	list, err := scanWebhooks(s.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return list, nil
}

// buildGetWebhooksQuery returns a SQL query (and arguments) that would return a
func (s *Sqlite) buildGetWebhooksQuery(filter *models.QueryFilter, userID uint64) (query string, args []interface{}) {
	var err error
	builder := s.sqlBuilder.
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
	s.logQueryBuildingError(err)

	return query, args
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter
func (s *Sqlite) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	query, args := s.buildGetWebhooksQuery(filter, userID)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, err := scanWebhooks(s.logger, rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	count, err := s.GetWebhookCount(ctx, filter, userID)
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
func (s *Sqlite) buildWebhookCreationQuery(x *models.Webhook) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
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
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// buildWebhookCreationTimeQuery returns a SQL query (and arguments) that fetches the DB creation time for a given row
func (s *Sqlite) buildWebhookCreationTimeQuery(webhookID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Select("created_on").
		From(webhooksTableName).
		Where(squirrel.Eq{"id": webhookID}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database
func (s *Sqlite) CreateWebhook(ctx context.Context, input *models.WebhookCreationInput) (*models.Webhook, error) {
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

	query, args := s.buildWebhookCreationQuery(x)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	if id, idErr := res.LastInsertId(); idErr == nil {
		x.ID = uint64(id)

		query, args = s.buildWebhookCreationTimeQuery(x.ID)
		s.logCreationTimeRetrievalError(s.db.QueryRowContext(ctx, query, args...).Scan(&x.CreatedOn))
	}

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update
func (s *Sqlite) buildUpdateWebhookQuery(input *models.Webhook) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
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
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (s *Sqlite) UpdateWebhook(ctx context.Context, input *models.Webhook) error {
	query, args := s.buildUpdateWebhookQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (s *Sqlite) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error
	query, args, err = s.sqlBuilder.
		Update(webhooksTableName).
		Set("updated_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Set("archived_on", squirrel.Expr(CurrentUnixTimeQuery)).
		Where(squirrel.Eq{
			"id":          webhookID,
			"belongs_to":  userID,
			"archived_on": nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveWebhook archives a webhook from the database by its ID
func (s *Sqlite) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	query, args := s.buildArchiveWebhookQuery(webhookID, userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

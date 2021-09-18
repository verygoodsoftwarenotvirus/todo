package postgres

import (
	"context"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	// webhooksTableEventsSeparator is what the webhooks table uses to separate event subscriptions.
	webhooksTableEventsSeparator = commaSeparator
	// webhooksTableDataTypesSeparator is what the webhooks table uses to separate data type subscriptions.
	webhooksTableDataTypesSeparator = commaSeparator
	// webhooksTableTopicsSeparator is what the webhooks table uses to separate topic subscriptions.
	webhooksTableTopicsSeparator = commaSeparator
)

var (
	_ types.WebhookDataManager = (*SQLQuerier)(nil)

	// webhooksTableColumns are the columns for the webhooks table.
	webhooksTableColumns = []string{
		"webhooks.id",
		"webhooks.name",
		"webhooks.content_type",
		"webhooks.url",
		"webhooks.method",
		"webhooks.events",
		"webhooks.data_types",
		"webhooks.topics",
		"webhooks.created_on",
		"webhooks.last_updated_on",
		"webhooks.archived_on",
		"webhooks.belongs_to_account",
	}
)

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct.
func (q *SQLQuerier) scanWebhook(ctx context.Context, scan database.Scanner, includeCounts bool) (webhook *types.Webhook, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)
	webhook = &types.Webhook{}

	var (
		eventsStr,
		dataTypesStr,
		topicsStr string
	)

	targetVars := []interface{}{
		&webhook.ID,
		&webhook.Name,
		&webhook.ContentType,
		&webhook.URL,
		&webhook.Method,
		&eventsStr,
		&dataTypesStr,
		&topicsStr,
		&webhook.CreatedOn,
		&webhook.LastUpdatedOn,
		&webhook.ArchivedOn,
		&webhook.BelongsToAccount,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "scanning webhook")
	}

	if events := strings.Split(eventsStr, webhooksTableEventsSeparator); len(events) >= 1 && events[0] != "" {
		webhook.Events = events
	}

	if dataTypes := strings.Split(dataTypesStr, webhooksTableDataTypesSeparator); len(dataTypes) >= 1 && dataTypes[0] != "" {
		webhook.DataTypes = dataTypes
	}

	if topics := strings.Split(topicsStr, webhooksTableTopicsSeparator); len(topics) >= 1 && topics[0] != "" {
		webhook.Topics = topics
	}

	return webhook, filteredCount, totalCount, nil
}

// scanWebhooks provides a consistent way to turn sql rows into a slice of webhooks.
func (q *SQLQuerier) scanWebhooks(ctx context.Context, rows database.ResultIterator, includeCounts bool) (webhooks []*types.Webhook, filteredCount, totalCount uint64, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		webhook, fc, tc, scanErr := q.scanWebhook(ctx, rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, scanErr
		}

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}

		webhooks = append(webhooks, webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "fetching webhook from database")
	}

	if err = rows.Close(); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "fetching webhook from database")
	}

	return webhooks, filteredCount, totalCount, nil
}

const getWebhookQuery = `
	SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_account FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = $1 AND webhooks.id = $2
`

// GetWebhook fetches a webhook from the database.
func (q *SQLQuerier) GetWebhook(ctx context.Context, webhookID, accountID string) (*types.Webhook, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" || accountID == "" {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachWebhookIDToSpan(span, webhookID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.AccountIDKey: accountID,
	})

	args := []interface{}{
		accountID,
		webhookID,
	}

	row := q.getOneRow(ctx, q.db, "webhook", getWebhookQuery, args)

	webhook, _, _, err := q.scanWebhook(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning webhook")
	}

	return webhook, nil
}

const getAllWebhooksCountQuery = `
	SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL
`

// GetAllWebhooksCount fetches the count of webhooks from the database that meet a particular filter.
func (q *SQLQuerier) GetAllWebhooksCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, getAllWebhooksCountQuery, "fetching count of webhooks")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of webhooks")
	}

	return count, nil
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (q *SQLQuerier) GetWebhooks(ctx context.Context, accountID string, filter *types.QueryFilter) (*types.WebhookList, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachQueryFilterToSpan(span, filter)

	x := &types.WebhookList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.buildListQuery(
		ctx,
		"webhooks",
		nil,
		nil,
		"belongs_to_account",
		webhooksTableColumns,
		accountID,
		false,
		filter,
	)

	rows, err := q.performReadQuery(ctx, q.db, "webhooks", query, args)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching webhook from database")
	}

	if x.Webhooks, x.FilteredCount, x.TotalCount, err = q.scanWebhooks(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning database response")
	}

	return x, nil
}

const createWebhookQuery = `
	INSERT INTO webhooks (id,name,content_type,url,method,events,data_types,topics,belongs_to_account) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
`

// CreateWebhook creates a webhook in a database.
func (q *SQLQuerier) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachAccountIDToSpan(span, input.BelongsToAccount)
	logger := q.logger.WithValue(keys.AccountIDKey, input.BelongsToAccount)

	args := []interface{}{
		input.ID,
		input.Name,
		input.ContentType,
		input.URL,
		input.Method,
		strings.Join(input.Events, webhooksTableEventsSeparator),
		strings.Join(input.DataTypes, webhooksTableDataTypesSeparator),
		strings.Join(input.Topics, webhooksTableTopicsSeparator),
		input.BelongsToAccount,
	}

	if err := q.performWriteQuery(ctx, q.db, "webhook creation", createWebhookQuery, args); err != nil {
		return nil, observability.PrepareError(err, logger, span, "creating webhook")
	}

	x := &types.Webhook{
		ID:               input.ID,
		Name:             input.Name,
		ContentType:      input.ContentType,
		URL:              input.URL,
		Method:           input.Method,
		Events:           input.Events,
		DataTypes:        input.DataTypes,
		Topics:           input.Topics,
		BelongsToAccount: input.BelongsToAccount,
		CreatedOn:        q.currentTime(),
	}

	tracing.AttachWebhookIDToSpan(span, x.ID)
	logger = logger.WithValue(keys.WebhookIDKey, x.ID)

	logger.Info("webhook created")

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
	last_updated_on = extract(epoch FROM NOW())
WHERE archived_on IS NULL 
AND belongs_to_account = $8 
AND id = $9
`

// UpdateWebhook updates a particular webhook.
func (q *SQLQuerier) UpdateWebhook(ctx context.Context, updated *types.Webhook) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	tracing.AttachWebhookIDToSpan(span, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.BelongsToAccount)

	logger := q.logger.
		WithValue(keys.WebhookIDKey, updated.ID).
		WithValue(keys.AccountIDKey, updated.BelongsToAccount)

	args := []interface{}{
		updated.Name,
		updated.ContentType,
		updated.URL,
		updated.Method,
		strings.Join(updated.Events, webhooksTableEventsSeparator),
		strings.Join(updated.DataTypes, webhooksTableDataTypesSeparator),
		strings.Join(updated.Topics, webhooksTableTopicsSeparator),
		updated.BelongsToAccount,
		updated.ID,
	}

	if err := q.performWriteQuery(ctx, q.db, "webhook update", updateWebhookQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "updating webhook")
	}

	return nil
}

const archiveWebhookQuery = `
UPDATE webhooks SET
	last_updated_on = extract(epoch FROM NOW()), 
	archived_on = extract(epoch FROM NOW())
WHERE archived_on IS NULL 
AND belongs_to_account = $1
AND id = $2
`

// ArchiveWebhook archives a webhook from the database.
func (q *SQLQuerier) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	tracing.AttachWebhookIDToSpan(span, webhookID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.AccountIDKey: accountID,
	})

	args := []interface{}{accountID, webhookID}

	if err := q.performWriteQuery(ctx, q.db, "webhook archive", archiveWebhookQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "archiving webhook")
	}

	logger.Info("webhook archived")

	return nil
}

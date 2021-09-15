package queriers

import (
	"context"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	_ types.WebhookDataManager = (*SQLQuerier)(nil)
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

	if events := strings.Split(eventsStr, querybuilding.WebhooksTableEventsSeparator); len(events) >= 1 && events[0] != "" {
		webhook.Events = events
	}

	if dataTypes := strings.Split(dataTypesStr, querybuilding.WebhooksTableDataTypesSeparator); len(dataTypes) >= 1 && dataTypes[0] != "" {
		webhook.DataTypes = dataTypes
	}

	if topics := strings.Split(topicsStr, querybuilding.WebhooksTableTopicsSeparator); len(topics) >= 1 && topics[0] != "" {
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

	query, args := q.sqlQueryBuilder.BuildGetWebhookQuery(ctx, webhookID, accountID)
	row := q.getOneRow(ctx, q.db, "webhook", query, args...)

	webhook, _, _, err := q.scanWebhook(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning webhook")
	}

	return webhook, nil
}

// GetAllWebhooksCount fetches the count of webhooks from the database that meet a particular filter.
func (q *SQLQuerier) GetAllWebhooksCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetAllWebhooksCountQuery(ctx), "fetching count of webhooks")
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

	query, args := q.sqlQueryBuilder.BuildGetWebhooksQuery(ctx, accountID, filter)

	rows, err := q.performReadQuery(ctx, q.db, "webhooks", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching webhook from database")
	}

	if x.Webhooks, x.FilteredCount, x.TotalCount, err = q.scanWebhooks(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning database response")
	}

	return x, nil
}

// CreateWebhook creates a webhook in a database.
func (q *SQLQuerier) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachAccountIDToSpan(span, input.BelongsToAccount)
	logger := q.logger.WithValue(keys.AccountIDKey, input.BelongsToAccount)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildCreateWebhookQuery(ctx, input)
	if writeErr := q.performWriteQueryIgnoringReturn(ctx, tx, "webhook creation", query, args); writeErr != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(writeErr, logger, span, "creating webhook")
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

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachWebhookIDToSpan(span, x.ID)
	logger = logger.WithValue(keys.WebhookIDKey, x.ID)

	logger.Info("webhook created")

	return x, nil
}

// UpdateWebhook updates a particular webhook.
// NOTE: this function expects the provided input to have a non-zero ID.
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

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildUpdateWebhookQuery(ctx, updated)
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "webhook update", query, args); err != nil {
		logger.Error(err, "updating webhook")
		q.rollbackTransaction(ctx, tx)

		return observability.PrepareError(err, logger, span, "updating webhook")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Debug("webhook updated")

	return nil
}

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

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildArchiveWebhookQuery(ctx, webhookID, accountID)

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "webhook archive", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving webhook")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("webhook archived")

	return nil
}

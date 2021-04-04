package querier

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
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
		&webhook.ExternalID,
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
func (q *SQLQuerier) GetWebhook(ctx context.Context, webhookID, accountID uint64) (*types.Webhook, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachWebhookIDToSpan(span, webhookID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.AccountIDKey: accountID,
	})

	query, args := q.sqlQueryBuilder.BuildGetWebhookQuery(ctx, webhookID, accountID)
	row := q.db.QueryRowContext(ctx, query, args...)

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
func (q *SQLQuerier) GetWebhooks(ctx context.Context, accountID uint64, filter *types.QueryFilter) (*types.WebhookList, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
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

	rows, err := q.performReadQuery(ctx, "webhooks", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching webhook from database")
	}

	if x.Webhooks, x.FilteredCount, x.TotalCount, err = q.scanWebhooks(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning database response")
	}

	return x, nil
}

// GetAllWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (q *SQLQuerier) GetAllWebhooks(ctx context.Context, resultChannel chan []*types.Webhook, batchSize uint16) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if resultChannel == nil {
		return ErrNilInputProvided
	}

	if batchSize == 0 {
		batchSize = defaultBatchSize
	}

	logger := q.logger.WithValue("batch_size", batchSize)

	count, err := q.GetAllWebhooksCount(ctx)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching count of webhooks")
	}

	increment := uint64(batchSize)

	for beginID := uint64(1); beginID <= count; beginID += increment {
		endID := beginID + increment
		go func(begin, end uint64) {
			query, args := q.sqlQueryBuilder.BuildGetBatchOfWebhooksQuery(ctx, begin, end)
			logger = logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := q.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			webhooks, _, _, scanErr := q.scanWebhooks(ctx, rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			resultChannel <- webhooks
		}(beginID, endID)
	}

	return nil
}

// CreateWebhook creates a webhook in a database.
func (q *SQLQuerier) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput, createdByUser uint64) (*types.Webhook, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachRequestingUserIDToSpan(span, createdByUser)
	tracing.AttachAccountIDToSpan(span, input.BelongsToAccount)
	logger := q.logger.WithValue(keys.AccountIDKey, input.BelongsToAccount)

	query, args := q.sqlQueryBuilder.BuildCreateWebhookQuery(ctx, input)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	id, err := q.performWriteQuery(ctx, tx, false, "webhook creation", query, args)
	if err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating webhook")
	}

	x := &types.Webhook{
		ID:               id,
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

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookCreationEventEntry(x, createdByUser)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing webhook creation audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachWebhookIDToSpan(span, x.ID)

	return x, nil
}

// UpdateWebhook updates a particular webhook.
// NOTE: this function expects the provided input to have a non-zero ID.
func (q *SQLQuerier) UpdateWebhook(ctx context.Context, updated *types.Webhook, changedByUser uint64, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	tracing.AttachRequestingUserIDToSpan(span, changedByUser)
	tracing.AttachWebhookIDToSpan(span, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.BelongsToAccount)

	logger := q.logger.
		WithValue(keys.WebhookIDKey, updated.ID).
		WithValue(keys.RequesterKey, changedByUser).
		WithValue(keys.AccountIDKey, updated.BelongsToAccount)

	query, args := q.sqlQueryBuilder.BuildUpdateWebhookQuery(ctx, updated)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "webhook update", query, args); err != nil {
		logger.Error(err, "updating webhook")
		q.rollbackTransaction(ctx, tx)

		return observability.PrepareError(err, logger, span, "updating webhook")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookUpdateEventEntry(changedByUser, updated.BelongsToAccount, updated.ID, changes)); err != nil {
		logger.Error(err, "writing webhook update audit log entry")
		q.rollbackTransaction(ctx, tx)

		return observability.PrepareError(err, logger, span, "writing webhook update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Debug("successfully updated webhook")

	return nil
}

// ArchiveWebhook archives a webhook from the database.
func (q *SQLQuerier) ArchiveWebhook(ctx context.Context, webhookID, accountID, archivedByUserID uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return ErrInvalidIDProvided
	}

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	tracing.AttachRequestingUserIDToSpan(span, archivedByUserID)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.AccountIDKey: accountID,
		keys.RequesterKey: archivedByUserID,
	})

	query, args := q.sqlQueryBuilder.BuildArchiveWebhookQuery(ctx, webhookID, accountID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "webhook archive", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving webhook")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookArchiveEventEntry(archivedByUserID, accountID, webhookID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing webhook archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForWebhook fetches a list of audit log entries from the database that relate to a given webhook.
func (q *SQLQuerier) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.WebhookIDKey, webhookID)
	query, args := q.sqlQueryBuilder.BuildGetAuditLogEntriesForWebhookQuery(ctx, webhookID)

	rows, err := q.performReadQuery(ctx, "audit log entries for webhook", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return auditLogEntries, nil
}

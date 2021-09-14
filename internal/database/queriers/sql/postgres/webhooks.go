package postgres

import (
	"context"
	"github.com/segmentio/ksuid"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
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
		querybuilding.WebhooksTableName,
		nil,
		nil,
		querybuilding.WebhooksTableOwnershipColumn,
		querybuilding.WebhooksTableColumns,
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
func (q *SQLQuerier) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput, createdByUser string) (*types.Webhook, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachRequestingUserIDToSpan(span, createdByUser)
	tracing.AttachAccountIDToSpan(span, input.BelongsToAccount)
	logger := q.logger.WithValue(keys.AccountIDKey, input.BelongsToAccount)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	args := []interface{}{
		input.ID,
		input.Name,
		input.ContentType,
		input.URL,
		input.Method,
		strings.Join(input.Events, querybuilding.WebhooksTableEventsSeparator),
		strings.Join(input.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
		strings.Join(input.Topics, querybuilding.WebhooksTableTopicsSeparator),
		input.BelongsToAccount,
	}

	if writeErr := q.performWriteQuery(ctx, tx, "webhook creation", createWebhookQuery, args); writeErr != nil {
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

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookCreationEventEntry(x, createdByUser)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing webhook creation audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachWebhookIDToSpan(span, x.ID)
	logger = logger.WithValue(keys.WebhookIDKey, x.ID)

	logger.Info("webhook created")

	return x, nil
}

const updateWebhookQuery = `
	UPDATE webhooks SET name = $1, content_type = $2, url = $3, method = $4, events = $5, data_types = $6, topics = $7, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_account = $8 AND id = $9
`

// UpdateWebhook updates a particular webhook.
func (q *SQLQuerier) UpdateWebhook(ctx context.Context, updated *types.Webhook, changedByUser string, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if changedByUser == "" {
		return ErrInvalidIDProvided
	}

	if updated == nil {
		return ErrNilInputProvided
	}

	tracing.AttachRequestingUserIDToSpan(span, changedByUser)
	tracing.AttachWebhookIDToSpan(span, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.BelongsToAccount)

	logger := q.logger.
		WithValue(keys.WebhookIDKey, updated.ID).
		WithValue(keys.RequesterIDKey, changedByUser).
		WithValue(keys.AccountIDKey, updated.BelongsToAccount)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	args := []interface{}{
		updated.Name,
		updated.ContentType,
		updated.URL,
		updated.Method,
		strings.Join(updated.Events, querybuilding.WebhooksTableEventsSeparator),
		strings.Join(updated.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
		strings.Join(updated.Topics, querybuilding.WebhooksTableTopicsSeparator),
		updated.BelongsToAccount,
		updated.ID,
	}

	if err = q.performWriteQuery(ctx, tx, "webhook update", updateWebhookQuery, args); err != nil {
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

	logger.Debug("webhook updated")

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
func (q *SQLQuerier) ArchiveWebhook(ctx context.Context, webhookID, accountID, archivedByUserID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" || accountID == "" || archivedByUserID == "" {
		return ErrInvalidIDProvided
	}

	tracing.AttachRequestingUserIDToSpan(span, archivedByUserID)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey:   webhookID,
		keys.AccountIDKey:   accountID,
		keys.RequesterIDKey: archivedByUserID,
	})

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	ksuid.New()

	args := []interface{}{accountID, webhookID}

	if err = q.performWriteQuery(ctx, tx, "webhook archive", archiveWebhookQuery, args); err != nil {
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

	logger.Info("webhook archived")

	return nil
}

const getAuditLogEntriesForWebhookQuery = `
SELECT 
	audit_log.id,
	audit_log.event_type,
	audit_log.context,
	audit_log.created_on
FROM audit_log 
WHERE audit_log.context->>'webhook_id' = $1 
ORDER BY audit_log.created_on
`

// GetAuditLogEntriesForWebhook fetches a list of audit log entries from the database that relate to a given webhook.
func (q *SQLQuerier) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID string) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.WebhookIDKey, webhookID)

	args := []interface{}{webhookID}

	rows, err := q.performReadQuery(ctx, q.db, "audit log entries for webhook", getAuditLogEntriesForWebhookQuery, args)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return auditLogEntries, nil
}

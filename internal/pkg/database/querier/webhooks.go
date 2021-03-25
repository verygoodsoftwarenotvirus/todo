package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	_ types.WebhookDataManager = (*Client)(nil)
)

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct.
func (c *Client) scanWebhook(ctx context.Context, scan database.Scanner, includeCounts bool) (webhook *types.Webhook, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)
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
func (c *Client) scanWebhooks(ctx context.Context, rows database.ResultIterator, includeCounts bool) (webhooks []*types.Webhook, filteredCount, totalCount uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		webhook, fc, tc, scanErr := c.scanWebhook(ctx, rows, includeCounts)
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
		c.logger.Error(err, "closing rows")
		return nil, 0, 0, observability.PrepareError(err, logger, span, "fetching webhook from database")
	}

	return webhooks, filteredCount, totalCount, nil
}

// GetWebhook fetches a webhook from the database.
func (c *Client) GetWebhook(ctx context.Context, webhookID, accountID uint64) (*types.Webhook, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, accountID)
	tracing.AttachWebhookIDToSpan(span, webhookID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.AccountIDKey: accountID,
	})

	query, args := c.sqlQueryBuilder.BuildGetWebhookQuery(webhookID, accountID)
	row := c.db.QueryRowContext(ctx, query, args...)

	webhook, _, _, err := c.scanWebhook(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning webhook")
	}

	return webhook, nil
}

// GetAllWebhooksCount fetches the count of webhooks from the database that meet a particular filter.
func (c *Client) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllWebhooksCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllWebhooksCountQuery())
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (c *Client) GetWebhooks(ctx context.Context, accountID uint64, filter *types.QueryFilter) (*types.WebhookList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachUserIDToSpan(span, accountID)
	tracing.AttachQueryFilterToSpan(span, filter)

	logger.Debug("GetWebhookCount called")

	x := &types.WebhookList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetWebhooksQuery(accountID, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching webhook from database")
	}

	if x.Webhooks, x.FilteredCount, x.TotalCount, err = c.scanWebhooks(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning database response")
	}

	return x, nil
}

// GetAllWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (c *Client) GetAllWebhooks(ctx context.Context, resultChannel chan []*types.Webhook, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("batch_size", batchSize)

	count, err := c.GetAllWebhooksCount(ctx)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching count of webhooks")
	}

	increment := uint64(batchSize)

	for beginID := uint64(1); beginID <= count; beginID += increment {
		endID := beginID + increment
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfWebhooksQuery(begin, end)
			logger = logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			webhooks, _, _, scanErr := c.scanWebhooks(ctx, rows, false)
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
func (c *Client) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput, createdByUser uint64) (*types.Webhook, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachRequestingUserIDToSpan(span, createdByUser)
	tracing.AttachAccountIDToSpan(span, input.BelongsToAccount)
	logger := c.logger.WithValue(keys.AccountIDKey, input.BelongsToAccount)

	logger.Debug("CreateWebhook called")

	query, args := c.sqlQueryBuilder.BuildCreateWebhookQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	id, err := c.performWriteQuery(ctx, tx, false, "webhook creation", query, args)
	if err != nil {
		c.rollbackTransaction(ctx, tx)
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
		CreatedOn:        c.currentTime(),
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookCreationEventEntry(x, createdByUser)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return nil, fmt.Errorf("writing <> audit log entry: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	return x, nil
}

// UpdateWebhook updates a particular webhook.
// NOTE: this function expects the provided input to have a non-zero ID.
func (c *Client) UpdateWebhook(ctx context.Context, updated *types.Webhook, changedByUser uint64, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachRequestingUserIDToSpan(span, changedByUser)
	tracing.AttachWebhookIDToSpan(span, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.BelongsToAccount)

	logger := c.logger.
		WithValue(keys.WebhookIDKey, updated.ID).
		WithValue(keys.RequesterKey, changedByUser).
		WithValue(keys.AccountIDKey, updated.BelongsToAccount)

	logger.Debug("UpdateWebhook called")

	query, args := c.sqlQueryBuilder.BuildUpdateWebhookQuery(updated)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "webhook update", query, args); err != nil {
		logger.Error(err, "updating webhook")
		c.rollbackTransaction(ctx, tx)

		return observability.PrepareError(err, logger, span, "updating webhook")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookUpdateEventEntry(changedByUser, updated.BelongsToAccount, updated.ID, changes)); err != nil {
		logger.Error(err, "writing webhook update audit log entry")
		c.rollbackTransaction(ctx, tx)

		return observability.PrepareError(err, logger, span, "writing webhook update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Debug("successfully updated webhook")

	return nil
}

// ArchiveWebhook archives a webhook from the database.
func (c *Client) ArchiveWebhook(ctx context.Context, webhookID, accountID, archivedByUserID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachRequestingUserIDToSpan(span, archivedByUserID)
	tracing.AttachWebhookIDToSpan(span, webhookID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.AccountIDKey: accountID,
		keys.RequesterKey: archivedByUserID,
	})

	logger.Debug("ArchiveWebhook called")

	query, args := c.sqlQueryBuilder.BuildArchiveWebhookQuery(webhookID, accountID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "webhook archive", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving webhook")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookArchiveEventEntry(archivedByUserID, accountID, webhookID)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing webhook archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForWebhook fetches a list of audit log entries from the database that relate to a given webhook.
func (c *Client) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.WebhookIDKey, webhookID)
	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForWebhookQuery(webhookID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return auditLogEntries, nil
}

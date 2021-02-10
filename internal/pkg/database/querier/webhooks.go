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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.WebhookDataManager = (*Client)(nil)
)

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct.
func (c *Client) scanWebhook(scan database.Scanner, includeCounts bool) (webhook *types.Webhook, filteredCount, totalCount uint64, err error) {
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
		&webhook.BelongsToUser,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
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
func (c *Client) scanWebhooks(rows database.ResultIterator, includeCounts bool) (webhooks []*types.Webhook, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		webhook, fc, tc, scanErr := c.scanWebhook(rows, includeCounts)
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

	if rowErr := rows.Err(); rowErr != nil {
		return nil, 0, 0, rowErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing rows")
		return nil, 0, 0, closeErr
	}

	return webhooks, filteredCount, totalCount, nil
}

// GetWebhook fetches a webhook from the database.
func (c *Client) GetWebhook(ctx context.Context, webhookID, userID uint64) (*types.Webhook, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachWebhookIDToSpan(span, webhookID)

	c.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.UserIDKey:    userID,
	}).Debug("GetWebhook called")

	query, args := c.sqlQueryBuilder.BuildGetWebhookQuery(webhookID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	webhook, _, _, err := c.scanWebhook(row, false)

	return webhook, err
}

// GetAllWebhooksCount fetches the count of webhooks from the database that meet a particular filter.
func (c *Client) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllWebhooksCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllWebhooksCountQuery())
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (c *Client) GetWebhooks(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.WebhookList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.WebhookList{}
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	tracing.AttachUserIDToSpan(span, userID)
	logger.Debug("GetWebhookCount called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetWebhooksQuery(userID, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if x.Webhooks, x.FilteredCount, x.TotalCount, err = c.scanWebhooks(rows, true); err != nil {
		return nil, fmt.Errorf("scanning database response: %w", err)
	}

	return x, nil
}

// GetAllWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (c *Client) GetAllWebhooks(ctx context.Context, resultChannel chan []*types.Webhook, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllWebhooks called")

	count, countErr := c.GetAllWebhooksCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of webhooks: %w", countErr)
	}

	increment := uint64(batchSize)

	for beginID := uint64(1); beginID <= count; beginID += increment {
		endID := beginID + increment
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfWebhooksQuery(begin, end)
			logger := c.logger.WithValues(map[string]interface{}{
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

			webhooks, _, _, scanErr := c.scanWebhooks(rows, false)
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
func (c *Client) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.BelongsToUser)
	c.logger.WithValue(keys.UserIDKey, input.BelongsToUser).Debug("CreateWebhook called")

	query, args := c.sqlQueryBuilder.BuildCreateWebhookQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}

	id, err := c.performWriteQuery(ctx, tx, false, "webhook creation", query, args)
	if err != nil {
		c.rollbackTransaction(tx)
		return nil, err
	}

	x := &types.Webhook{
		ID:            id,
		Name:          input.Name,
		ContentType:   input.ContentType,
		URL:           input.URL,
		Method:        input.Method,
		Events:        input.Events,
		DataTypes:     input.DataTypes,
		Topics:        input.Topics,
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     c.currentTime(),
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildWebhookCreationEventEntry(x))

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return x, nil
}

// UpdateWebhook updates a particular webhook.
// NOTE: this function expects the provided input to have a non-zero ID.
func (c *Client) UpdateWebhook(ctx context.Context, updated *types.Webhook, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachWebhookIDToSpan(span, updated.ID)
	tracing.AttachUserIDToSpan(span, updated.BelongsToUser)

	c.logger.WithValue(keys.WebhookIDKey, updated.ID).Debug("UpdateWebhook called")

	query, args := c.sqlQueryBuilder.BuildUpdateWebhookQuery(updated)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "webhook update", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error updating webhook: %w", err)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// ArchiveWebhook archives a webhook from the database.
func (c *Client) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachWebhookIDToSpan(span, webhookID)

	c.logger.WithValues(map[string]interface{}{
		keys.WebhookIDKey: webhookID,
		keys.UserIDKey:    userID,
	}).Debug("ArchiveWebhook called")

	query, args := c.sqlQueryBuilder.BuildArchiveWebhookQuery(webhookID, userID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "webhook archive", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error archiving webhook: %w", err)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetAuditLogEntriesForWebhook fetches a list of audit log entries from the database that relate to a given webhook.
func (c *Client) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForWebhook called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForWebhookQuery(webhookID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

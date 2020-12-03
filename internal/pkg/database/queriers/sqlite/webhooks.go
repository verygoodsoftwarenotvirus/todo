package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.WebhookDataManager = (*Sqlite)(nil)

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct.
func (q *Sqlite) scanWebhook(scan database.Scanner, includeCount bool) (*types.Webhook, uint64, error) {
	var (
		x     = &types.Webhook{}
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
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	if events := strings.Split(eventsStr, queriers.WebhooksTableEventsSeparator); len(events) >= 1 && events[0] != "" {
		x.Events = events
	}

	if dataTypes := strings.Split(dataTypesStr, queriers.WebhooksTableDataTypesSeparator); len(dataTypes) >= 1 && dataTypes[0] != "" {
		x.DataTypes = dataTypes
	}

	if topics := strings.Split(topicsStr, queriers.WebhooksTableTopicsSeparator); len(topics) >= 1 && topics[0] != "" {
		x.Topics = topics
	}

	return x, count, nil
}

// scanWebhooks provides a consistent way to turn sql rows into a slice of webhooks.
func (q *Sqlite) scanWebhooks(rows database.ResultIterator, includeCount bool) ([]types.Webhook, uint64, error) {
	var (
		list  []types.Webhook
		count uint64
	)

	for rows.Next() {
		webhook, c, err := q.scanWebhook(rows, includeCount)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 && includeCount {
			count = c
		}

		list = append(list, *webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := rows.Close(); err != nil {
		q.logger.Error(err, "closing rows")
	}

	return list, count, nil
}

// buildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook.
func (q *Sqlite) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn):                     webhookID,
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.WebhooksTableOwnershipColumn): userID,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetWebhook fetches a webhook from the database.
func (q *Sqlite) GetWebhook(ctx context.Context, webhookID, userID uint64) (*types.Webhook, error) {
	query, args := q.buildGetWebhookQuery(webhookID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	webhook, _, err := q.scanWebhook(row, false)
	if err != nil {
		return nil, fmt.Errorf("querying database for webhook: %w", err)
	}

	return webhook, nil
}

// buildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (q *Sqlite) buildGetAllWebhooksCountQuery() string {
	var err error

	getAllWebhooksCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.WebhooksTableName)).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database.
func (q *Sqlite) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfWebhooksQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Sqlite) buildGetBatchOfWebhooksQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAllWebhooks fetches every item from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (q *Sqlite) GetAllWebhooks(ctx context.Context, resultChannel chan []types.Webhook) error {
	count, err := q.GetAllWebhooksCount(ctx)
	if err != nil {
		return fmt.Errorf("error fetching count of webhooks: %w", err)
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := q.buildGetBatchOfWebhooksQuery(begin, end)
			logger := q.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := q.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			webhooks, _, err := q.scanWebhooks(rows, false)
			if err != nil {
				logger.Error(err, "scanning database rows")
				return
			}

			resultChannel <- webhooks
		}(beginID, endID)
	}

	return nil
}

// buildGetWebhooksQuery returns a SQL query (and arguments) that would return a list of webhooks.
func (q *Sqlite) buildGetWebhooksQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.
		Select(allCountQuery).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.WebhooksTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn):             nil,
		})

	if filter != nil {
		countQueryBuilder = queriers.ApplyFilterToSubCountQueryBuilder(filter, countQueryBuilder, queriers.ItemsTableName)
	}

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(queriers.WebhooksTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.WebhooksTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn):             nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.WebhooksTableName)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (q *Sqlite) GetWebhooks(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.WebhookList, error) {
	query, args := q.buildGetWebhooksQuery(userID, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, count, err := q.scanWebhooks(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	x := &types.WebhookList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Webhooks: list,
	}

	return x, err
}

// buildCreateWebhookQuery returns a SQL query (and arguments) that would create a given webhook.
func (q *Sqlite) buildCreateWebhookQuery(x *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(queriers.WebhooksTableName).
		Columns(
			queriers.WebhooksTableNameColumn,
			queriers.WebhooksTableContentTypeColumn,
			queriers.WebhooksTableURLColumn,
			queriers.WebhooksTableMethodColumn,
			queriers.WebhooksTableEventsColumn,
			queriers.WebhooksTableDataTypesColumn,
			queriers.WebhooksTableTopicsColumn,
			queriers.WebhooksTableOwnershipColumn,
		).
		Values(
			x.Name,
			x.ContentType,
			x.URL,
			x.Method,
			strings.Join(x.Events, queriers.WebhooksTableEventsSeparator),
			strings.Join(x.DataTypes, queriers.WebhooksTableDataTypesSeparator),
			strings.Join(x.Topics, queriers.WebhooksTableTopicsSeparator),
			x.BelongsToUser,
		).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database.
func (q *Sqlite) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
	x := &types.Webhook{
		Name:          input.Name,
		ContentType:   input.ContentType,
		URL:           input.URL,
		Method:        input.Method,
		Events:        input.Events,
		DataTypes:     input.DataTypes,
		Topics:        input.Topics,
		BelongsToUser: input.BelongsToUser,
	}
	query, args := q.buildCreateWebhookQuery(x)

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	q.logIDRetrievalError(err)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = q.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
func (q *Sqlite) buildUpdateWebhookQuery(input *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.WebhooksTableName).
		Set(queriers.WebhooksTableNameColumn, input.Name).
		Set(queriers.WebhooksTableContentTypeColumn, input.ContentType).
		Set(queriers.WebhooksTableURLColumn, input.URL).
		Set(queriers.WebhooksTableMethodColumn, input.Method).
		Set(queriers.WebhooksTableEventsColumn, strings.Join(input.Events, queriers.WebhooksTableEventsSeparator)).
		Set(queriers.WebhooksTableDataTypesColumn, strings.Join(input.DataTypes, queriers.WebhooksTableDataTypesSeparator)).
		Set(queriers.WebhooksTableTopicsColumn, strings.Join(input.Topics, queriers.WebhooksTableTopicsSeparator)).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                     input.ID,
			queriers.WebhooksTableOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (q *Sqlite) UpdateWebhook(ctx context.Context, input *types.Webhook) error {
	query, args := q.buildUpdateWebhookQuery(input)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (q *Sqlite) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.WebhooksTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                     webhookID,
			queriers.WebhooksTableOwnershipColumn: userID,
			queriers.ArchivedOnColumn:             nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchiveWebhook archives a webhook from the database by its ID.
func (q *Sqlite) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	query, args := q.buildArchiveWebhookQuery(webhookID, userID)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// LogWebhookCreationEvent saves a WebhookCreationEvent in the audit log table.
func (q *Sqlite) LogWebhookCreationEvent(ctx context.Context, webhook *types.Webhook) {
	q.createAuditLogEntry(ctx, audit.BuildWebhookCreationEventEntry(webhook))
}

// LogWebhookUpdateEvent saves a WebhookUpdateEvent in the audit log table.
func (q *Sqlite) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []types.FieldChangeSummary) {
	q.createAuditLogEntry(ctx, audit.BuildWebhookUpdateEventEntry(userID, webhookID, changes))
}

// LogWebhookArchiveEvent saves a WebhookArchiveEvent in the audit log table.
func (q *Sqlite) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildWebhookArchiveEventEntry(userID, webhookID))
}

// buildGetAuditLogEntriesForWebhookQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Sqlite) buildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{}) {
	var err error

	webhookIDKey := fmt.Sprintf(
		jsonPluckQuery,
		queriers.AuditLogEntriesTableName,
		queriers.AuditLogEntriesTableContextColumn,
		audit.WebhookAssignmentKey,
	)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{webhookIDKey: webhookID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForWebhook fetches an audit log entry from the database.
func (q *Sqlite) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]types.AuditLogEntry, error) {
	query, args := q.buildGetAuditLogEntriesForWebhookQuery(webhookID)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

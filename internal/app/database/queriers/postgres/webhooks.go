package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// scanWebhook is a consistent way to turn a *sql.Row into a webhook struct.
func (p *Postgres) scanWebhook(scan database.Scanner) (*types.Webhook, error) {
	var (
		x = &types.Webhook{}
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

	if events := strings.Split(eventsStr, queriers.WebhooksTableEventsSeparator); len(events) >= 1 && events[0] != "" {
		x.Events = events
	}

	if dataTypes := strings.Split(dataTypesStr, queriers.WebhooksTableDataTypesSeparator); len(dataTypes) >= 1 && dataTypes[0] != "" {
		x.DataTypes = dataTypes
	}

	if topics := strings.Split(topicsStr, queriers.WebhooksTableTopicsSeparator); len(topics) >= 1 && topics[0] != "" {
		x.Topics = topics
	}

	return x, nil
}

// scanWebhooks provides a consistent way to turn sql rows into a slice of webhooks.
func (p *Postgres) scanWebhooks(rows database.ResultIterator) ([]types.Webhook, error) {
	var list []types.Webhook

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
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn):                     webhookID,
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.WebhooksTableOwnershipColumn): userID,
		}).ToSql()

	p.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database.
func (p *Postgres) GetWebhook(ctx context.Context, webhookID, userID uint64) (*types.Webhook, error) {
	query, args := p.buildGetWebhookQuery(webhookID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)

	webhook, err := p.scanWebhook(row)
	if err != nil {
		return nil, fmt.Errorf("querying database webhook: %w", err)
	}

	return webhook, nil
}

// buildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (p *Postgres) buildGetAllWebhooksCountQuery() string {
	var err error

	getAllWebhooksCountQuery, _, err := p.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.WebhooksTableName)).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn): nil,
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
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	p.logQueryBuildingError(err)

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database.
func (p *Postgres) GetAllWebhooks(ctx context.Context) (*types.WebhookList, error) {
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

	x := &types.WebhookList{
		Pagination: types.Pagination{
			Page: 1,
		},
		Webhooks: list,
	}

	return x, err
}

// buildGetWebhooksQuery returns a SQL query (and arguments) that would return a list of webhooks.
func (p *Postgres) buildGetWebhooksQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := p.sqlBuilder.
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.WebhooksTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn):             nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder, queriers.WebhooksTableName)
	}

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (p *Postgres) GetWebhooks(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.WebhookList, error) {
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

	x := &types.WebhookList{
		Pagination: types.Pagination{
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Webhooks: list,
	}

	return x, err
}

// buildCreateWebhookQuery returns a SQL query (and arguments) that would create a given webhook.
func (p *Postgres) buildCreateWebhookQuery(x *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
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
		Suffix(fmt.Sprintf("RETURNING %s, %s", queriers.IDColumn, queriers.CreatedOnColumn)).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database.
func (p *Postgres) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
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

	query, args := p.buildCreateWebhookQuery(x)
	if err := p.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn); err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
func (p *Postgres) buildUpdateWebhookQuery(input *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(queriers.WebhooksTableName).
		Set(queriers.WebhooksTableNameColumn, input.Name).
		Set(queriers.WebhooksTableContentTypeColumn, input.ContentType).
		Set(queriers.WebhooksTableURLColumn, input.URL).
		Set(queriers.WebhooksTableMethodColumn, input.Method).
		Set(queriers.WebhooksTableEventsColumn, strings.Join(input.Events, queriers.WebhooksTableTopicsSeparator)).
		Set(queriers.WebhooksTableDataTypesColumn, strings.Join(input.DataTypes, queriers.WebhooksTableDataTypesSeparator)).
		Set(queriers.WebhooksTableTopicsColumn, strings.Join(input.Topics, queriers.WebhooksTableTopicsSeparator)).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                     input.ID,
			queriers.WebhooksTableOwnershipColumn: input.BelongsToUser,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.LastUpdatedOnColumn)).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (p *Postgres) UpdateWebhook(ctx context.Context, input *types.Webhook) error {
	query, args := p.buildUpdateWebhookQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (p *Postgres) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(queriers.WebhooksTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                     webhookID,
			queriers.WebhooksTableOwnershipColumn: userID,
			queriers.ArchivedOnColumn:             nil,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.ArchivedOnColumn)).
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

// LogWebhookCreationEvent saves a WebhookCreationEvent in the audit log table.
func (p *Postgres) LogWebhookCreationEvent(ctx context.Context, webhook *types.Webhook) {
	p.createAuditLogEntry(ctx, audit.BuildWebhookCreationEventEntry(webhook))
}

// LogWebhookUpdateEvent saves a WebhookUpdateEvent in the audit log table.
func (p *Postgres) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []types.FieldChangeSummary) {
	p.createAuditLogEntry(ctx, audit.BuildWebhookUpdateEventEntry(userID, webhookID, changes))
}

// LogWebhookArchiveEvent saves a WebhookArchiveEvent in the audit log table.
func (p *Postgres) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	p.createAuditLogEntry(ctx, audit.BuildWebhookArchiveEventEntry(userID, webhookID))
}

// buildGetAuditLogEntriesForWebhookQuery constructs a SQL query for fetching audit log entries
// associated with a given webhook.
func (p *Postgres) buildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{}) {
	var err error

	webhookIDKey := fmt.Sprintf("%s.%s->'%s'",
		queriers.AuditLogEntriesTableName,
		queriers.AuditLogEntriesTableContextColumn,
		audit.WebhookAssignmentKey,
	)
	builder := p.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{webhookIDKey: webhookID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	p.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForWebhook fetches a audit log entries for a given webhook from the database.
func (p *Postgres) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]types.AuditLogEntry, error) {
	query, args := p.buildGetAuditLogEntriesForWebhookQuery(webhookID)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := p.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

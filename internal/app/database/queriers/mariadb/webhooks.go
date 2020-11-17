package mariadb

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
func (m *MariaDB) scanWebhook(scan database.Scanner) (*types.Webhook, error) {
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
func (m *MariaDB) scanWebhooks(rows database.ResultIterator) ([]types.Webhook, error) {
	var (
		list []types.Webhook
	)

	for rows.Next() {
		webhook, err := m.scanWebhook(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, *webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		m.logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook.
func (m *MariaDB) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn):                     webhookID,
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.WebhooksTableOwnershipColumn): userID,
		}).ToSql()

	m.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database.
func (m *MariaDB) GetWebhook(ctx context.Context, webhookID, userID uint64) (*types.Webhook, error) {
	query, args := m.buildGetWebhookQuery(webhookID, userID)
	row := m.db.QueryRowContext(ctx, query, args...)

	webhook, err := m.scanWebhook(row)
	if err != nil {
		return nil, fmt.Errorf("fetching webhook from database: %w", err)
	}

	return webhook, nil
}

// buildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (m *MariaDB) buildGetAllWebhooksCountQuery() string {
	var err error

	getAllWebhooksCountQuery, _, err := m.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.WebhooksTableName)).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database.
func (m *MariaDB) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = m.db.QueryRowContext(ctx, m.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

// buildGetAllWebhooksQuery returns a SQL query which will return all webhooks, regardless of ownership.
func (m *MariaDB) buildGetAllWebhooksQuery() string {
	var err error

	getAllWebhooksQuery, _, err := m.sqlBuilder.
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database.
func (m *MariaDB) GetAllWebhooks(ctx context.Context) (*types.WebhookList, error) {
	rows, err := m.db.QueryContext(ctx, m.buildGetAllWebhooksQuery())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("querying for webhooks: %w", err)
	}

	list, err := m.scanWebhooks(rows)
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

// buildGetWebhooksQuery returns a SQL query (and arguments) that would return a query and arguments to retrieve a list of webhooks.
func (m *MariaDB) buildGetWebhooksQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
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
	m.logQueryBuildingError(err)

	return query, args
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (m *MariaDB) GetWebhooks(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.WebhookList, error) {
	query, args := m.buildGetWebhooksQuery(userID, filter)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, err := m.scanWebhooks(rows)
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
func (m *MariaDB) buildCreateWebhookQuery(x *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
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

	m.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database.
func (m *MariaDB) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
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
	query, args := m.buildCreateWebhookQuery(x)

	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	m.logIDRetrievalError(err)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = m.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
func (m *MariaDB) buildUpdateWebhookQuery(input *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
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
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (m *MariaDB) UpdateWebhook(ctx context.Context, input *types.Webhook) error {
	query, args := m.buildUpdateWebhookQuery(input)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (m *MariaDB) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.WebhooksTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                     webhookID,
			queriers.WebhooksTableOwnershipColumn: userID,
			queriers.ArchivedOnColumn:             nil,
		}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// ArchiveWebhook archives a webhook from the database by its ID.
func (m *MariaDB) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	query, args := m.buildArchiveWebhookQuery(webhookID, userID)
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// LogWebhookCreationEvent saves a WebhookCreationEvent in the audit log table.
func (m *MariaDB) LogWebhookCreationEvent(ctx context.Context, webhook *types.Webhook) {
	m.createAuditLogEntry(ctx, audit.BuildWebhookCreationEventEntry(webhook))
}

// LogWebhookUpdateEvent saves a WebhookUpdateEvent in the audit log table.
func (m *MariaDB) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []types.FieldChangeSummary) {
	m.createAuditLogEntry(ctx, audit.BuildWebhookUpdateEventEntry(userID, webhookID, changes))
}

// LogWebhookArchiveEvent saves a WebhookArchiveEvent in the audit log table.
func (m *MariaDB) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	m.createAuditLogEntry(ctx, audit.BuildWebhookArchiveEventEntry(userID, webhookID))
}

// buildGetAuditLogEntriesForWebhookQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (m *MariaDB) buildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{}) {
	var err error

	builder := m.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					queriers.AuditLogEntriesTableName,
					queriers.AuditLogEntriesTableContextColumn,
					webhookID,
					audit.WebhookAssignmentKey,
				),
			),
		).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	m.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForWebhook fetches an audit log entry from the database.
func (m *MariaDB) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]types.AuditLogEntry, error) {
	query, args := m.buildGetAuditLogEntriesForWebhookQuery(webhookID)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := m.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

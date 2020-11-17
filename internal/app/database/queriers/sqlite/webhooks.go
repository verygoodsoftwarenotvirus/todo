package sqlite

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
func (s *Sqlite) scanWebhook(scan database.Scanner) (*types.Webhook, error) {
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
func (s *Sqlite) scanWebhooks(rows database.ResultIterator) ([]types.Webhook, error) {
	var (
		list []types.Webhook
	)

	for rows.Next() {
		webhook, err := s.scanWebhook(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, *webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := rows.Close(); err != nil {
		s.logger.Error(err, "closing rows")
	}

	return list, nil
}

// buildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook.
func (s *Sqlite) buildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.IDColumn):                     webhookID,
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.WebhooksTableOwnershipColumn): userID,
		}).ToSql()

	s.logQueryBuildingError(err)
	return query, args
}

// GetWebhook fetches a webhook from the database.
func (s *Sqlite) GetWebhook(ctx context.Context, webhookID, userID uint64) (*types.Webhook, error) {
	query, args := s.buildGetWebhookQuery(webhookID, userID)
	row := s.db.QueryRowContext(ctx, query, args...)

	webhook, err := s.scanWebhook(row)
	if err != nil {
		return nil, fmt.Errorf("querying database for webhook: %w", err)
	}

	return webhook, nil
}

// buildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (s *Sqlite) buildGetAllWebhooksCountQuery() string {
	var err error

	getAllWebhooksCountQuery, _, err := s.sqlBuilder.
		Select(fmt.Sprintf(countQuery, queriers.WebhooksTableName)).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return getAllWebhooksCountQuery
}

// GetAllWebhooksCount will fetch the count of every active webhook in the database.
func (s *Sqlite) GetAllWebhooksCount(ctx context.Context) (count uint64, err error) {
	err = s.db.QueryRowContext(ctx, s.buildGetAllWebhooksCountQuery()).Scan(&count)
	return count, err
}

// buildGetAllWebhooksQuery returns a SQL query which will return all webhooks, regardless of ownership.
func (s *Sqlite) buildGetAllWebhooksQuery() string {
	var err error

	getAllWebhooksQuery, _, err := s.sqlBuilder.
		Select(queriers.WebhooksTableColumns...).
		From(queriers.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.WebhooksTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return getAllWebhooksQuery
}

// GetAllWebhooks fetches a list of all webhooks from the database.
func (s *Sqlite) GetAllWebhooks(ctx context.Context) (*types.WebhookList, error) {
	rows, err := s.db.QueryContext(ctx, s.buildGetAllWebhooksQuery())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("querying for webhooks: %w", err)
	}

	list, err := s.scanWebhooks(rows)
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
func (s *Sqlite) buildGetWebhooksQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	var err error

	builder := s.sqlBuilder.
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
	s.logQueryBuildingError(err)

	return query, args
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (s *Sqlite) GetWebhooks(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.WebhookList, error) {
	query, args := s.buildGetWebhooksQuery(userID, filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("querying database: %w", err)
	}

	list, err := s.scanWebhooks(rows)
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
func (s *Sqlite) buildCreateWebhookQuery(x *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
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

	s.logQueryBuildingError(err)

	return query, args
}

// CreateWebhook creates a webhook in the database.
func (s *Sqlite) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
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
	query, args := s.buildCreateWebhookQuery(x)

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing webhook creation query: %w", err)
	}

	// fetch the last inserted ID.
	id, err := res.LastInsertId()
	s.logIDRetrievalError(err)

	// this won't be completely accurate, but it will suffice.
	x.CreatedOn = s.timeTeller.Now()
	x.ID = uint64(id)

	return x, nil
}

// buildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
func (s *Sqlite) buildUpdateWebhookQuery(input *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
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

	s.logQueryBuildingError(err)

	return query, args
}

// UpdateWebhook updates a particular webhook. Note that UpdateWebhook expects the provided input to have a valid ID.
func (s *Sqlite) UpdateWebhook(ctx context.Context, input *types.Webhook) error {
	query, args := s.buildUpdateWebhookQuery(input)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// buildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (s *Sqlite) buildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(queriers.WebhooksTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                     webhookID,
			queriers.WebhooksTableOwnershipColumn: userID,
			queriers.ArchivedOnColumn:             nil,
		}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// ArchiveWebhook archives a webhook from the database by its ID.
func (s *Sqlite) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	query, args := s.buildArchiveWebhookQuery(webhookID, userID)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// LogWebhookCreationEvent saves a WebhookCreationEvent in the audit log table.
func (s *Sqlite) LogWebhookCreationEvent(ctx context.Context, webhook *types.Webhook) {
	s.createAuditLogEntry(ctx, audit.BuildWebhookCreationEventEntry(webhook))
}

// LogWebhookUpdateEvent saves a WebhookUpdateEvent in the audit log table.
func (s *Sqlite) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []types.FieldChangeSummary) {
	s.createAuditLogEntry(ctx, audit.BuildWebhookUpdateEventEntry(userID, webhookID, changes))
}

// LogWebhookArchiveEvent saves a WebhookArchiveEvent in the audit log table.
func (s *Sqlite) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	s.createAuditLogEntry(ctx, audit.BuildWebhookArchiveEventEntry(userID, webhookID))
}

// buildGetAuditLogEntriesForWebhookQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (s *Sqlite) buildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{}) {
	var err error

	webhookIDKey := fmt.Sprintf(
		jsonPluckQuery,
		queriers.AuditLogEntriesTableName,
		queriers.AuditLogEntriesTableContextColumn,
		audit.WebhookAssignmentKey,
	)
	builder := s.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{webhookIDKey: webhookID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	s.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForWebhook fetches an audit log entry from the database.
func (s *Sqlite) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]types.AuditLogEntry, error) {
	query, args := s.buildGetAuditLogEntriesForWebhookQuery(webhookID)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, err := s.scanAuditLogEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}

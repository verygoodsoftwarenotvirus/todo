package postgres

import (
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.WebhookSQLQueryBuilder = (*Postgres)(nil)

// BuildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook.
func (q *Postgres) BuildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.WebhooksTableColumns...).
		From(querybuilding.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn):                     webhookID,
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.WebhooksTableOwnershipColumn): userID,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (q *Postgres) BuildGetAllWebhooksCountQuery() string {
	var err error

	getAllWebhooksCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.WebhooksTableName)).
		From(querybuilding.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return getAllWebhooksCountQuery
}

// BuildGetBatchOfWebhooksQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Postgres) BuildGetBatchOfWebhooksQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(querybuilding.WebhooksTableColumns...).
		From(querybuilding.WebhooksTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetWebhooksQuery returns a SQL query (and arguments) that would return a list of webhooks.
func (q *Postgres) BuildGetWebhooksQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.WebhooksTableName,
		querybuilding.WebhooksTableOwnershipColumn,
		querybuilding.WebhooksTableColumns,
		userID,
		false,
		filter,
	)
}

// BuildCreateWebhookQuery returns a SQL query (and arguments) that would create a given webhook.
func (q *Postgres) BuildCreateWebhookQuery(x *types.WebhookCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(querybuilding.WebhooksTableName).
		Columns(
			querybuilding.WebhooksTableNameColumn,
			querybuilding.WebhooksTableContentTypeColumn,
			querybuilding.WebhooksTableURLColumn,
			querybuilding.WebhooksTableMethodColumn,
			querybuilding.WebhooksTableEventsColumn,
			querybuilding.WebhooksTableDataTypesColumn,
			querybuilding.WebhooksTableTopicsColumn,
			querybuilding.WebhooksTableOwnershipColumn,
		).
		Values(
			x.Name,
			x.ContentType,
			x.URL,
			x.Method,
			strings.Join(x.Events, querybuilding.WebhooksTableEventsSeparator),
			strings.Join(x.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
			strings.Join(x.Topics, querybuilding.WebhooksTableTopicsSeparator),
			x.BelongsToUser,
		).
		Suffix(fmt.Sprintf("RETURNING %s", querybuilding.IDColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
func (q *Postgres) BuildUpdateWebhookQuery(input *types.Webhook) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.WebhooksTableName).
		Set(querybuilding.WebhooksTableNameColumn, input.Name).
		Set(querybuilding.WebhooksTableContentTypeColumn, input.ContentType).
		Set(querybuilding.WebhooksTableURLColumn, input.URL).
		Set(querybuilding.WebhooksTableMethodColumn, input.Method).
		Set(querybuilding.WebhooksTableEventsColumn, strings.Join(input.Events, querybuilding.WebhooksTableTopicsSeparator)).
		Set(querybuilding.WebhooksTableDataTypesColumn, strings.Join(input.DataTypes, querybuilding.WebhooksTableDataTypesSeparator)).
		Set(querybuilding.WebhooksTableTopicsColumn, strings.Join(input.Topics, querybuilding.WebhooksTableTopicsSeparator)).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                     input.ID,
			querybuilding.WebhooksTableOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (q *Postgres) BuildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.WebhooksTableName).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                     webhookID,
			querybuilding.WebhooksTableOwnershipColumn: userID,
			querybuilding.ArchivedOnColumn:             nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForWebhookQuery constructs a SQL query for fetching audit log entries
// associated with a given webhook.
func (q *Postgres) BuildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{}) {
	webhookIDKey := fmt.Sprintf(jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.WebhookAssignmentKey,
	)
	builder := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{webhookIDKey: webhookID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	return q.buildQuery(builder)
}

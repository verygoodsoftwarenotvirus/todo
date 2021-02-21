package sqlite

import (
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.WebhookSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook.
func (q *Sqlite) BuildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.WebhooksTableColumns...).
		From(querybuilding.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn):                     webhookID,
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.WebhooksTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.ArchivedOnColumn):             nil,
		}),
	)
}

// BuildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (q *Sqlite) BuildGetAllWebhooksCountQuery() string {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.WebhooksTableName)).
		From(querybuilding.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetBatchOfWebhooksQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Sqlite) BuildGetBatchOfWebhooksQuery(beginID, endID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.WebhooksTableColumns...).
		From(querybuilding.WebhooksTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn): endID,
		}),
	)
}

// BuildGetWebhooksQuery returns a SQL query (and arguments) that would return a list of webhooks.
func (q *Sqlite) BuildGetWebhooksQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
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
func (q *Sqlite) BuildCreateWebhookQuery(x *types.WebhookCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.WebhooksTableName).
		Columns(
			querybuilding.ExternalIDColumn,
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
			q.externalIDGenerator.NewExternalID(),
			x.Name,
			x.ContentType,
			x.URL,
			x.Method,
			strings.Join(x.Events, querybuilding.WebhooksTableEventsSeparator),
			strings.Join(x.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
			strings.Join(x.Topics, querybuilding.WebhooksTableTopicsSeparator),
			x.BelongsToUser,
		),
	)
}

// BuildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
func (q *Sqlite) BuildUpdateWebhookQuery(input *types.Webhook) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.WebhooksTableName).
		Set(querybuilding.WebhooksTableNameColumn, input.Name).
		Set(querybuilding.WebhooksTableContentTypeColumn, input.ContentType).
		Set(querybuilding.WebhooksTableURLColumn, input.URL).
		Set(querybuilding.WebhooksTableMethodColumn, input.Method).
		Set(querybuilding.WebhooksTableEventsColumn, strings.Join(input.Events, querybuilding.WebhooksTableEventsSeparator)).
		Set(querybuilding.WebhooksTableDataTypesColumn, strings.Join(input.DataTypes, querybuilding.WebhooksTableDataTypesSeparator)).
		Set(querybuilding.WebhooksTableTopicsColumn, strings.Join(input.Topics, querybuilding.WebhooksTableTopicsSeparator)).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                     input.ID,
			querybuilding.WebhooksTableOwnershipColumn: input.BelongsToAccount,
			querybuilding.ArchivedOnColumn:             nil,
		}),
	)
}

// BuildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (q *Sqlite) BuildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.WebhooksTableName).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                     webhookID,
			querybuilding.WebhooksTableOwnershipColumn: userID,
			querybuilding.ArchivedOnColumn:             nil,
		}),
	)
}

// BuildGetAuditLogEntriesForWebhookQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Sqlite) BuildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{}) {
	webhookIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.WebhookAssignmentKey,
	)

	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{webhookIDKey: webhookID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

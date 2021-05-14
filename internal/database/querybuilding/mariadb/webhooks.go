package mariadb

import (
	"context"
	"fmt"
	"strings"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ querybuilding.WebhookSQLQueryBuilder = (*MariaDB)(nil)

// BuildGetWebhookQuery returns a SQL query (and arguments) for retrieving a given webhook.
func (b *MariaDB) BuildGetWebhookQuery(ctx context.Context, webhookID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachWebhookIDToSpan(span, webhookID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.WebhooksTableColumns...).
			From(querybuilding.WebhooksTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn):                     webhookID,
				fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.WebhooksTableOwnershipColumn): accountID,
				fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.ArchivedOnColumn):             nil,
			}),
	)
}

// BuildGetAllWebhooksCountQuery returns a query which would return the count of webhooks regardless of ownership.
func (b *MariaDB) BuildGetAllWebhooksCountQuery(ctx context.Context) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQueryOnly(span, b.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.WebhooksTableName)).
		From(querybuilding.WebhooksTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.ArchivedOnColumn): nil,
		}))
}

// BuildGetBatchOfWebhooksQuery returns a query that fetches every item in the database within a bucketed range.
func (b *MariaDB) BuildGetBatchOfWebhooksQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.WebhooksTableColumns...).
			From(querybuilding.WebhooksTableName).
			Where(squirrel.Gt{
				fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn): beginID,
			}).
			Where(squirrel.Lt{
				fmt.Sprintf("%s.%s", querybuilding.WebhooksTableName, querybuilding.IDColumn): endID,
			}),
	)
}

// BuildGetWebhooksQuery returns a SQL query (and arguments) that would return a query and arguments to retrieve a list of webhooks.
func (b *MariaDB) BuildGetWebhooksQuery(ctx context.Context, accountID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}
	return b.buildListQuery(ctx, querybuilding.WebhooksTableName, querybuilding.WebhooksTableOwnershipColumn, querybuilding.WebhooksTableColumns, accountID, false, filter)
}

// BuildCreateWebhookQuery returns a SQL query (and arguments) that would create a given webhook.
func (b *MariaDB) BuildCreateWebhookQuery(ctx context.Context, x *types.WebhookCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.WebhooksTableName).
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
				b.externalIDGenerator.NewExternalID(),
				x.Name,
				x.ContentType,
				x.URL,
				x.Method,
				strings.Join(x.Events, querybuilding.WebhooksTableEventsSeparator),
				strings.Join(x.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
				strings.Join(x.Topics, querybuilding.WebhooksTableTopicsSeparator),
				x.BelongsToAccount,
			),
	)
}

// BuildUpdateWebhookQuery takes a given webhook and returns a SQL query to update.
func (b *MariaDB) BuildUpdateWebhookQuery(ctx context.Context, input *types.Webhook) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachWebhookIDToSpan(span, input.ID)
	tracing.AttachAccountIDToSpan(span, input.BelongsToAccount)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.WebhooksTableName).
			Set(querybuilding.WebhooksTableNameColumn, input.Name).
			Set(querybuilding.WebhooksTableContentTypeColumn, input.ContentType).
			Set(querybuilding.WebhooksTableURLColumn, input.URL).
			Set(querybuilding.WebhooksTableMethodColumn, input.Method).
			Set(querybuilding.WebhooksTableEventsColumn, strings.Join(input.Events, querybuilding.WebhooksTableTopicsSeparator)).
			Set(querybuilding.WebhooksTableDataTypesColumn, strings.Join(input.DataTypes, querybuilding.WebhooksTableDataTypesSeparator)).
			Set(querybuilding.WebhooksTableTopicsColumn, strings.Join(input.Topics, querybuilding.WebhooksTableTopicsSeparator)).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                     input.ID,
				querybuilding.ArchivedOnColumn:             nil,
				querybuilding.WebhooksTableOwnershipColumn: input.BelongsToAccount,
			}),
	)
}

// BuildArchiveWebhookQuery returns a SQL query (and arguments) that will mark a webhook as archived.
func (b *MariaDB) BuildArchiveWebhookQuery(ctx context.Context, webhookID, accountID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachWebhookIDToSpan(span, webhookID)
	tracing.AttachAccountIDToSpan(span, accountID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.WebhooksTableName).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                     webhookID,
				querybuilding.WebhooksTableOwnershipColumn: accountID,
				querybuilding.ArchivedOnColumn:             nil,
			}),
	)
}

// BuildGetAuditLogEntriesForWebhookQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (b *MariaDB) BuildGetAuditLogEntriesForWebhookQuery(ctx context.Context, webhookID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachWebhookIDToSpan(span, webhookID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(
				squirrel.Expr(
					fmt.Sprintf(
						jsonPluckQuery,
						querybuilding.AuditLogEntriesTableName,
						querybuilding.AuditLogEntriesTableContextColumn,
						webhookID,
						audit.WebhookAssignmentKey,
					),
				),
			).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

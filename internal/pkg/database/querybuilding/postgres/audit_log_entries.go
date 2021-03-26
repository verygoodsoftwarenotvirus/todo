package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ querybuilding.AuditLogEntrySQLQueryBuilder = (*Postgres)(nil)

// BuildGetAuditLogEntryQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (b *Postgres) BuildGetAuditLogEntryQuery(ctx context.Context, entryID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAuditLogEntryIDToSpan(span, entryID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): entryID,
			}),
	)
}

// BuildGetAllAuditLogEntriesCountQuery returns a query that fetches the total number of  in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (b *Postgres) BuildGetAllAuditLogEntriesCountQuery(ctx context.Context) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	allAuditLogEntriesCountQuery, _, err := b.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AuditLogEntriesTableName)).
		From(querybuilding.AuditLogEntriesTableName).
		ToSql()
	b.logQueryBuildingError(span, err)

	return allAuditLogEntriesCountQuery
}

// BuildGetBatchOfAuditLogEntriesQuery returns a query that fetches every audit log entry in the database within a bucketed range.
func (b *Postgres) BuildGetBatchOfAuditLogEntriesQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Gt{
				fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): beginID,
			}).
			Where(squirrel.Lt{
				fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.IDColumn): endID,
			}),
	)
}

// BuildCreateAuditLogEntryQuery takes an audit log entry and returns a creation query for that audit log entry and the relevant arguments.
func (b *Postgres) BuildCreateAuditLogEntryQuery(ctx context.Context, input *types.AuditLogEntryCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAuditLogEntryEventTypeToSpan(span, input.EventType)

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.AuditLogEntriesTableName).
			Columns(
				querybuilding.ExternalIDColumn,
				querybuilding.AuditLogEntriesTableEventTypeColumn,
				querybuilding.AuditLogEntriesTableContextColumn,
			).
			Values(
				b.externalIDGenerator.NewExternalID(),
				input.EventType,
				input.Context,
			).
			Suffix(fmt.Sprintf("RETURNING %s", querybuilding.IDColumn)),
	)
}

// BuildGetAuditLogEntriesQuery builds a SQL query selecting  that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (b *Postgres) BuildGetAuditLogEntriesQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}
	countQuery, countQueryArgs := b.buildQuery(
		span,
		b.sqlBuilder.Select(allCountQuery).
			From(querybuilding.AuditLogEntriesTableName))

	query, selectArgs := b.buildQuery(span, querybuilding.ApplyFilterToQueryBuilder(
		filter,
		querybuilding.AuditLogEntriesTableName,
		b.sqlBuilder.Select(append(querybuilding.AuditLogEntriesTableColumns, fmt.Sprintf("(%s)", countQuery))...).
			From(querybuilding.AuditLogEntriesTableName).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	))

	return query, append(countQueryArgs, selectArgs...)
}

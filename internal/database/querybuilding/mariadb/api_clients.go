package mariadb

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ querybuilding.APIClientSQLQueryBuilder = (*MariaDB)(nil)
)

// BuildGetBatchOfAPIClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (b *MariaDB) BuildGetBatchOfAPIClientsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.APIClientsTableColumns...).
			From(querybuilding.APIClientsTableName).
			Where(squirrel.Gt{
				fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.IDColumn): beginID,
			}).
			Where(squirrel.Lt{
				fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.IDColumn): endID,
			}),
	)
}

// BuildGetAPIClientByClientIDQuery returns a SQL query which requests a given API client by its database ID.
func (b *MariaDB) BuildGetAPIClientByClientIDQuery(ctx context.Context, clientID string) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAPIClientClientIDToSpan(span, clientID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.APIClientsTableColumns...).
			From(querybuilding.APIClientsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.APIClientsTableClientIDColumn): clientID,
				fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.ArchivedOnColumn):              nil,
			}),
	)
}

// BuildGetAPIClientByDatabaseIDQuery returns a SQL query which requests a given API client by its database ID.
func (b *MariaDB) BuildGetAPIClientByDatabaseIDQuery(ctx context.Context, clientID, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAPIClientDatabaseIDToSpan(span, clientID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.APIClientsTableColumns...).
			From(querybuilding.APIClientsTableName).
			Where(squirrel.Eq{
				fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.APIClientsTableOwnershipColumn): userID,
				fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.IDColumn):                       clientID,
				fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.ArchivedOnColumn):               nil,
			}),
	)
}

// BuildGetAllAPIClientsCountQuery returns a SQL query for the number of API clients
// in the database, regardless of ownership.
func (b *MariaDB) BuildGetAllAPIClientsCountQuery(ctx context.Context) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQueryOnly(span, b.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.APIClientsTableName)).
		From(querybuilding.APIClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.ArchivedOnColumn): nil,
		}))
}

// BuildGetAPIClientsQuery returns a SQL query (and arguments) that will retrieve a list of API clients that
// meet the given filter's criteria (if relevant) and belong to a given account.
func (b *MariaDB) BuildGetAPIClientsQuery(ctx context.Context, userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	return b.buildListQuery(ctx, querybuilding.APIClientsTableName, querybuilding.APIClientsTableOwnershipColumn, querybuilding.APIClientsTableColumns, userID, false, filter)
}

// BuildCreateAPIClientQuery returns a SQL query (and args) that will create the given APIClient in the database.
func (b *MariaDB) BuildCreateAPIClientQuery(ctx context.Context, input *types.APIClientCreationInput) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return b.buildQuery(
		span,
		b.sqlBuilder.Insert(querybuilding.APIClientsTableName).
			Columns(
				querybuilding.ExternalIDColumn,
				querybuilding.APIClientsTableNameColumn,
				querybuilding.APIClientsTableClientIDColumn,
				querybuilding.APIClientsTableSecretKeyColumn,
				querybuilding.APIClientsTableOwnershipColumn,
			).
			Values(
				b.externalIDGenerator.NewExternalID(),
				input.Name,
				input.ClientID,
				input.ClientSecret,
				input.BelongsToUser,
			),
	)
}

// BuildUpdateAPIClientQuery returns a SQL query (and args) that will update a given API client in the database.
func (b *MariaDB) BuildUpdateAPIClientQuery(ctx context.Context, input *types.APIClient) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, input.BelongsToUser)
	tracing.AttachAPIClientClientIDToSpan(span, input.ClientID)
	tracing.AttachAPIClientDatabaseIDToSpan(span, input.ID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.APIClientsTableName).
			Set(querybuilding.APIClientsTableClientIDColumn, input.ClientID).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                       input.ID,
				querybuilding.ArchivedOnColumn:               nil,
				querybuilding.APIClientsTableOwnershipColumn: input.BelongsToUser,
			}),
	)
}

// BuildArchiveAPIClientQuery returns a SQL query (and arguments) that will mark an API client as archived.
func (b *MariaDB) BuildArchiveAPIClientQuery(ctx context.Context, clientID, userID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAPIClientDatabaseIDToSpan(span, clientID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Update(querybuilding.APIClientsTableName).
			Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
			Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
			Where(squirrel.Eq{
				querybuilding.IDColumn:                       clientID,
				querybuilding.ArchivedOnColumn:               nil,
				querybuilding.APIClientsTableOwnershipColumn: userID,
			}),
	)
}

// BuildGetAuditLogEntriesForAPIClientQuery constructs a SQL query for fetching audit log entries belong to a user with a given ID.
func (b *MariaDB) BuildGetAuditLogEntriesForAPIClientQuery(ctx context.Context, clientID uint64) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAPIClientDatabaseIDToSpan(span, clientID)

	return b.buildQuery(
		span,
		b.sqlBuilder.Select(querybuilding.AuditLogEntriesTableColumns...).
			From(querybuilding.AuditLogEntriesTableName).
			Where(squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					querybuilding.AuditLogEntriesTableName,
					querybuilding.AuditLogEntriesTableContextColumn,
					clientID,
					audit.APIClientAssignmentKey,
				),
			)).
			OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

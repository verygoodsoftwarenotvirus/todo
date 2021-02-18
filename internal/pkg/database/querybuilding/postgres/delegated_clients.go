package postgres

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.DelegatedClientSQLQueryBuilder = (*Postgres)(nil)
)

// BuildGetBatchOfDelegatedClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Postgres) BuildGetBatchOfDelegatedClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.DelegatedClientsTableColumns...).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.IDColumn): endID,
		}),
	)
}

// BuildGetDelegatedClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (q *Postgres) BuildGetDelegatedClientQuery(clientID string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.DelegatedClientsTableColumns...).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.DelegatedClientsTableClientIDColumn): clientID,
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.ArchivedOnColumn):                    nil,
		}),
	)
}

// BuildGetAllDelegatedClientsCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (q *Postgres) BuildGetAllDelegatedClientsCountQuery() string {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.DelegatedClientsTableName)).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetDelegatedClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (q *Postgres) BuildGetDelegatedClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.DelegatedClientsTableName,
		querybuilding.DelegatedClientsTableOwnershipColumn,
		querybuilding.DelegatedClientsTableColumns,
		userID,
		false,
		filter,
	)
}

// BuildGetDelegatedClientByDatabaseIDQuery returns a SQL query which requests a given delegated client by its database ID.
func (q *Postgres) BuildGetDelegatedClientByDatabaseIDQuery(clientID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.DelegatedClientsTableColumns...).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.DelegatedClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.IDColumn):                             clientID,
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.ArchivedOnColumn):                     nil,
		}),
	)
}

// BuildCreateDelegatedClientQuery returns a SQL query (and args) that will create the given DelegatedClient in the database.
func (q *Postgres) BuildCreateDelegatedClientQuery(input *types.DelegatedClientCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.DelegatedClientsTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.DelegatedClientsTableNameColumn,
			querybuilding.DelegatedClientsTableClientIDColumn,
			querybuilding.DelegatedClientsTableSecretKeyColumn,
			querybuilding.DelegatedClientsTableOwnershipColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Name,
			input.ClientID,
			input.ClientSecret,
			input.BelongsToUser,
		).
		Suffix(fmt.Sprintf("RETURNING %s", querybuilding.IDColumn)),
	)
}

// BuildUpdateDelegatedClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (q *Postgres) BuildUpdateDelegatedClientQuery(input *types.DelegatedClient) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.DelegatedClientsTableName).
		Set(querybuilding.DelegatedClientsTableClientIDColumn, input.ClientID).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                             input.ID,
			querybuilding.ArchivedOnColumn:                     nil,
			querybuilding.DelegatedClientsTableOwnershipColumn: input.BelongsToUser,
		}),
	)
}

// BuildArchiveDelegatedClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *Postgres) BuildArchiveDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.DelegatedClientsTableName).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                             clientID,
			querybuilding.DelegatedClientsTableOwnershipColumn: userID,
			querybuilding.ArchivedOnColumn:                     nil,
		}),
	)
}

// BuildGetAuditLogEntriesForDelegatedClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Postgres) BuildGetAuditLogEntriesForDelegatedClientQuery(clientID uint64) (query string, args []interface{}) {
	delegatedClientIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.DelegatedClientAssignmentKey,
	)

	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{delegatedClientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

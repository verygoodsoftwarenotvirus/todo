package base

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.APIClientSQLQueryBuilder = (*QueryBuilder)(nil)
)

// BuildGetBatchOfAPIClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *QueryBuilder) BuildGetBatchOfAPIClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.APIClientsTableColumns...).
		From(querybuilding.APIClientsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.IDColumn): endID,
		}),
	)
}

// BuildGetAPIClientByClientIDQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (q *QueryBuilder) BuildGetAPIClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.APIClientsTableColumns...).
		From(querybuilding.APIClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.APIClientsTableClientIDColumn): clientID,
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.ArchivedOnColumn):              nil,
		}),
	)
}

// BuildGetAllAPIClientsCountQuery returns a SQL query for the number of OAuth2 clients
// returns the database, regardless of ownership.
func (q *QueryBuilder) BuildGetAllAPIClientsCountQuery() string {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.APIClientsTableName)).
		From(querybuilding.APIClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetAPIClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// returns the given filter's criteria (if relevant) and belong to a given account.
func (q *QueryBuilder) BuildGetAPIClientsQuery(accountID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.APIClientsTableName,
		querybuilding.APIClientsTableOwnershipColumn,
		querybuilding.APIClientsTableColumns,
		accountID,
		false,
		filter,
	)
}

// BuildGetAPIClientByDatabaseIDQuery returns a SQL query which requests a given API client by its database ID.
func (q *QueryBuilder) BuildGetAPIClientByDatabaseIDQuery(clientID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.APIClientsTableColumns...).
		From(querybuilding.APIClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.APIClientsTableOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.IDColumn):                       clientID,
			fmt.Sprintf("%s.%s", querybuilding.APIClientsTableName, querybuilding.ArchivedOnColumn):               nil,
		}),
	)
}

// BuildCreateAPIClientQuery returns a SQL query (and args) that will create the given APIClient in the database.
func (q *QueryBuilder) BuildCreateAPIClientQuery(input *types.APICientCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.APIClientsTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.APIClientsTableNameColumn,
			querybuilding.APIClientsTableClientIDColumn,
			querybuilding.APIClientsTableSecretKeyColumn,
			querybuilding.APIClientsTableOwnershipColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Name,
			input.ClientID,
			input.ClientSecret,
			input.BelongsToAccount,
		),
	)
}

// BuildUpdateAPIClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (q *QueryBuilder) BuildUpdateAPIClientQuery(input *types.APIClient) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.APIClientsTableName).
		Set(querybuilding.APIClientsTableClientIDColumn, input.ClientID).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                       input.ID,
			querybuilding.APIClientsTableOwnershipColumn: input.BelongsToAccount,
			querybuilding.ArchivedOnColumn:               nil,
		}),
	)
}

// BuildArchiveAPIClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *QueryBuilder) BuildArchiveAPIClientQuery(clientID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.APIClientsTableName).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                       clientID,
			querybuilding.ArchivedOnColumn:               nil,
			querybuilding.APIClientsTableOwnershipColumn: accountID,
		}),
	)
}

// BuildGetAuditLogEntriesForAPIClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *QueryBuilder) BuildGetAuditLogEntriesForAPIClientQuery(clientID uint64) (query string, args []interface{}) {
	apiClientIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.APIClientAssignmentKey,
	)

	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{apiClientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

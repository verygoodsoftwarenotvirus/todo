package mariadb

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.DelegatedClientSQLQueryBuilder = (*MariaDB)(nil)
)

// BuildGetBatchOfDelegatedClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *MariaDB) BuildGetBatchOfDelegatedClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
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
func (q *MariaDB) BuildGetDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.DelegatedClientsTableColumns...).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.IDColumn):                             clientID,
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.DelegatedClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.ArchivedOnColumn):                     nil,
		}),
	)
}

// BuildGetAllDelegatedClientsCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (q *MariaDB) BuildGetAllDelegatedClientsCountQuery() string {
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
func (q *MariaDB) BuildGetDelegatedClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.DelegatedClientsTableName,
		querybuilding.DelegatedClientsTableOwnershipColumn,
		querybuilding.DelegatedClientsTableColumns,
		userID,
		false,
		filter,
	)
}

// BuildCreateDelegatedClientQuery returns a SQL query (and args) that will create the given DelegatedClient in the database.
func (q *MariaDB) BuildCreateDelegatedClientQuery(input *types.DelegatedClientCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.DelegatedClientsTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.DelegatedClientsTableNameColumn,
			querybuilding.DelegatedClientsTableClientIDColumn,
			querybuilding.DelegatedClientsTableClientSecretColumn,
			querybuilding.DelegatedClientsTableOwnershipColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Name,
			input.ClientID,
			input.ClientSecret,
			input.BelongsToUser,
		),
	)
}

// BuildUpdateDelegatedClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (q *MariaDB) BuildUpdateDelegatedClientQuery(input *types.DelegatedClient) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.DelegatedClientsTableName).
		Set(querybuilding.DelegatedClientsTableClientIDColumn, input.ClientID).
		Set(querybuilding.DelegatedClientsTableClientSecretColumn, input.ClientSecret).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                             input.ID,
			querybuilding.DelegatedClientsTableOwnershipColumn: input.BelongsToUser,
		}),
	)
}

// BuildArchiveDelegatedClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *MariaDB) BuildArchiveDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.DelegatedClientsTableName).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                             clientID,
			querybuilding.DelegatedClientsTableOwnershipColumn: userID,
		}),
	)
}

// BuildGetAuditLogEntriesForDelegatedClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *MariaDB) BuildGetAuditLogEntriesForDelegatedClientQuery(clientID uint64) (query string, args []interface{}) {
	oauth2ClientIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		clientID,
		audit.DelegatedClientAssignmentKey,
	)

	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{oauth2ClientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

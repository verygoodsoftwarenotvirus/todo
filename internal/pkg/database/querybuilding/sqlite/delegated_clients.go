package sqlite

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.DelegatedClientSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildGetBatchOfDelegatedClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Sqlite) BuildGetBatchOfDelegatedClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(querybuilding.DelegatedClientsTableColumns...).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetDelegatedClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (q *Sqlite) BuildGetDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.DelegatedClientsTableColumns...).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.IDColumn):                             clientID,
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.DelegatedClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.ArchivedOnColumn):                     nil,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllDelegatedClientsCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (q *Sqlite) BuildGetAllDelegatedClientsCountQuery() string {
	var err error

	getAllDelegatedClientCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.DelegatedClientsTableName)).
		From(querybuilding.DelegatedClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.DelegatedClientsTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return getAllDelegatedClientCountQuery
}

// BuildGetDelegatedClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (q *Sqlite) BuildGetDelegatedClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
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
func (q *Sqlite) BuildCreateDelegatedClientQuery(input *types.DelegatedClientCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
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
		).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateDelegatedClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (q *Sqlite) BuildUpdateDelegatedClientQuery(input *types.DelegatedClient) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.DelegatedClientsTableName).
		Set(querybuilding.DelegatedClientsTableClientIDColumn, input.ClientID).
		Set(querybuilding.DelegatedClientsTableClientSecretColumn, input.ClientSecret).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                             input.ID,
			querybuilding.DelegatedClientsTableOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveDelegatedClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *Sqlite) BuildArchiveDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.DelegatedClientsTableName).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                             clientID,
			querybuilding.DelegatedClientsTableOwnershipColumn: userID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForDelegatedClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Sqlite) BuildGetAuditLogEntriesForDelegatedClientQuery(clientID uint64) (query string, args []interface{}) {
	var err error

	oauth2ClientIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.DelegatedClientAssignmentKey,
	)
	builder := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{oauth2ClientIDKey: clientID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

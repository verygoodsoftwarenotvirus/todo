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
	_ types.OAuth2ClientSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID.
func (q *Sqlite) BuildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	var err error

	// This query is more or less the same as the normal OAuth2 client retrieval query, only that it doesn't
	// care about ownership. It does still care about archived status
	query, args, err = q.sqlBuilder.
		Select(querybuilding.OAuth2ClientsTableColumns...).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.OAuth2ClientsTableClientIDColumn): clientID,
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.ArchivedOnColumn):                 nil,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetBatchOfOAuth2ClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Sqlite) BuildGetBatchOfOAuth2ClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(querybuilding.OAuth2ClientsTableColumns...).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (q *Sqlite) BuildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.OAuth2ClientsTableColumns...).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.IDColumn):                          clientID,
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.ArchivedOnColumn):                  nil,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllOAuth2ClientsCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (q *Sqlite) BuildGetAllOAuth2ClientsCountQuery() string {
	var err error

	getAllOAuth2ClientCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.OAuth2ClientsTableName)).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return getAllOAuth2ClientCountQuery
}

// BuildGetOAuth2ClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (q *Sqlite) BuildGetOAuth2ClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.OAuth2ClientsTableName,
		querybuilding.OAuth2ClientsTableOwnershipColumn,
		querybuilding.OAuth2ClientsTableColumns,
		userID,
		false,
		filter,
	)
}

// BuildCreateOAuth2ClientQuery returns a SQL query (and args) that will create the given OAuth2Client in the database.
func (q *Sqlite) BuildCreateOAuth2ClientQuery(input *types.OAuth2ClientCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(querybuilding.OAuth2ClientsTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.OAuth2ClientsTableNameColumn,
			querybuilding.OAuth2ClientsTableClientIDColumn,
			querybuilding.OAuth2ClientsTableClientSecretColumn,
			querybuilding.OAuth2ClientsTableScopesColumn,
			querybuilding.OAuth2ClientsTableRedirectURIColumn,
			querybuilding.OAuth2ClientsTableOwnershipColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Name,
			input.ClientID,
			input.ClientSecret,
			strings.Join(input.Scopes, querybuilding.OAuth2ClientsTableScopeSeparator),
			input.RedirectURI,
			input.BelongsToUser,
		).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateOAuth2ClientQuery returns a SQL query (and args) that will update a given OAuth2 client in the database.
func (q *Sqlite) BuildUpdateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.OAuth2ClientsTableName).
		Set(querybuilding.OAuth2ClientsTableClientIDColumn, input.ClientID).
		Set(querybuilding.OAuth2ClientsTableClientSecretColumn, input.ClientSecret).
		Set(querybuilding.OAuth2ClientsTableScopesColumn, strings.Join(input.Scopes, querybuilding.OAuth2ClientsTableScopeSeparator)).
		Set(querybuilding.OAuth2ClientsTableRedirectURIColumn, input.RedirectURI).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                          input.ID,
			querybuilding.OAuth2ClientsTableOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *Sqlite) BuildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.OAuth2ClientsTableName).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                          clientID,
			querybuilding.OAuth2ClientsTableOwnershipColumn: userID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForOAuth2ClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Sqlite) BuildGetAuditLogEntriesForOAuth2ClientQuery(clientID uint64) (query string, args []interface{}) {
	var err error

	oauth2ClientIDKey := fmt.Sprintf(
		jsonPluckQuery,
		querybuilding.AuditLogEntriesTableName,
		querybuilding.AuditLogEntriesTableContextColumn,
		audit.OAuth2ClientAssignmentKey,
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

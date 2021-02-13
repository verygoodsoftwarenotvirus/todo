package mariadb

import (
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.OAuth2ClientSQLQueryBuilder = (*MariaDB)(nil)

// BuildGetOAuth2ClientByClientIDQuery builds a SQL query for fetching an OAuth2 client by its ClientID.
func (q *MariaDB) BuildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.OAuth2ClientsTableColumns...).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.OAuth2ClientsTableClientIDColumn): clientID,
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.ArchivedOnColumn):                 nil,
		}),
	)
}

// BuildGetBatchOfOAuth2ClientsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *MariaDB) BuildGetBatchOfOAuth2ClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.OAuth2ClientsTableColumns...).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.IDColumn): endID,
		}),
	)
}

// BuildGetOAuth2ClientQuery returns a SQL query which requests a given OAuth2 client by its database ID.
func (q *MariaDB) BuildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.OAuth2ClientsTableColumns...).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.IDColumn):                          clientID,
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.OAuth2ClientsTableOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.ArchivedOnColumn):                  nil,
		}),
	)
}

// BuildGetAllOAuth2ClientsCountQuery returns a SQL query for the number of OAuth2 clients
// in the database, regardless of ownership.
func (q *MariaDB) BuildGetAllOAuth2ClientsCountQuery() string {
	return q.buildQueryOnly(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.OAuth2ClientsTableName)).
		From(querybuilding.OAuth2ClientsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.OAuth2ClientsTableName, querybuilding.ArchivedOnColumn): nil,
		}),
	)
}

// BuildGetOAuth2ClientsQuery returns a SQL query (and arguments) that will retrieve a list of OAuth2 clients that
// meet the given filter's criteria (if relevant) and belong to a given user.
func (q *MariaDB) BuildGetOAuth2ClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
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
func (q *MariaDB) BuildCreateOAuth2ClientQuery(input *types.OAuth2ClientCreationInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
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
		),
	)
}

// BuildArchiveOAuth2ClientQuery returns a SQL query (and arguments) that will mark an OAuth2 client as archived.
func (q *MariaDB) BuildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.OAuth2ClientsTableName).
		Set(querybuilding.LastUpdatedOnColumn, currentUnixTimeQuery).
		Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                          clientID,
			querybuilding.OAuth2ClientsTableOwnershipColumn: userID,
			querybuilding.ArchivedOnColumn:                  nil,
		}),
	)
}

// BuildGetAuditLogEntriesForOAuth2ClientQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *MariaDB) BuildGetAuditLogEntriesForOAuth2ClientQuery(clientID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(
			squirrel.Expr(
				fmt.Sprintf(
					jsonPluckQuery,
					querybuilding.AuditLogEntriesTableName,
					querybuilding.AuditLogEntriesTableContextColumn,
					clientID,
					audit.OAuth2ClientAssignmentKey,
				),
			),
		).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn)),
	)
}

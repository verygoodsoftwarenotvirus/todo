package sqlite

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.AccountSQLQueryBuilder = (*Sqlite)(nil)
)

// BuildGetAccountQuery constructs a SQL query for fetching an account with a given ID belong to a user with a given ID.
func (q *Sqlite) BuildGetAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(querybuilding.AccountsTableColumns...).
		From(querybuilding.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.IDColumn):                         accountID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.AccountsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.ArchivedOnColumn):                 nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllAccountsCountQuery returns a query that fetches the total number of accounts in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Sqlite) BuildGetAllAccountsCountQuery() string {
	var err error

	allAccountsCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, querybuilding.AccountsTableName)).
		From(querybuilding.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.ArchivedOnColumn): nil,
		}).
		ToSql()
	q.logQueryBuildingError(err)

	return allAccountsCountQuery
}

// BuildGetBatchOfAccountsQuery returns a query that fetches every account in the database within a bucketed range.
func (q *Sqlite) BuildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(querybuilding.AccountsTableColumns...).
		From(querybuilding.AccountsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", querybuilding.AccountsTableName, querybuilding.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAccountsQuery builds a SQL query selecting accounts that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Sqlite) BuildGetAccountsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		querybuilding.AccountsTableName,
		querybuilding.AccountsTableUserOwnershipColumn,
		querybuilding.AccountsTableColumns,
		userID,
		forAdmin,
		filter,
	)
}

// BuildCreateAccountQuery takes an account and returns a creation query for that account and the relevant arguments.
func (q *Sqlite) BuildCreateAccountQuery(input *types.AccountCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(querybuilding.AccountsTableName).
		Columns(
			querybuilding.ExternalIDColumn,
			querybuilding.AccountsTableNameColumn,
			querybuilding.AccountsTableUserOwnershipColumn,
		).
		Values(
			q.externalIDGenerator.NewExternalID(),
			input.Name,
			input.BelongsToUser,
		).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateAccountQuery takes an account and returns an update SQL query, with the relevant query parameters.
func (q *Sqlite) BuildUpdateAccountQuery(input *types.Account) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.AccountsTableName).
		Set(querybuilding.AccountsTableNameColumn, input.Name).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                         input.ID,
			querybuilding.AccountsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveAccountQuery returns a SQL query which marks a given account belonging to a given user as archived.
func (q *Sqlite) BuildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(querybuilding.AccountsTableName).
		Set(querybuilding.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(querybuilding.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			querybuilding.IDColumn:                         accountID,
			querybuilding.ArchivedOnColumn:                 nil,
			querybuilding.AccountsTableUserOwnershipColumn: userID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForAccountQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Sqlite) BuildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{}) {
	var err error

	accountIDKey := fmt.Sprintf(jsonPluckQuery, querybuilding.AuditLogEntriesTableName, querybuilding.AuditLogEntriesTableContextColumn, audit.AccountAssignmentKey)
	builder := q.sqlBuilder.
		Select(querybuilding.AuditLogEntriesTableColumns...).
		From(querybuilding.AuditLogEntriesTableName).
		Where(squirrel.Eq{accountIDKey: accountID}).
		OrderBy(fmt.Sprintf("%s.%s", querybuilding.AuditLogEntriesTableName, querybuilding.CreatedOnColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

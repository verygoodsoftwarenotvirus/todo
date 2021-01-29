package postgres

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.AccountSQLQueryBuilder = (*Postgres)(nil)

// BuildGetAccountQuery constructs a SQL query for fetching an account with a given ID belong to a user with a given ID.
func (q *Postgres) BuildGetAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.AccountsTableColumns...).
		From(queriers.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn):                         accountID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.AccountsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn):                 nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAllAccountsCountQuery returns a query that fetches the total number of accounts in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Postgres) BuildGetAllAccountsCountQuery() string {
	allAccountsCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.AccountsTableName)).
		From(queriers.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	q.logQueryBuildingError(err)

	return allAccountsCountQuery
}

// BuildGetBatchOfAccountsQuery returns a query that fetches every account in the database within a bucketed range.
func (q *Postgres) BuildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(queriers.AccountsTableColumns...).
		From(queriers.AccountsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAccountsQuery builds a SQL query selecting accounts that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Postgres) BuildGetAccountsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		queriers.AccountsTableName,
		queriers.AccountsTableUserOwnershipColumn,
		queriers.AccountsTableColumns,
		userID,
		forAdmin,
		filter,
	)
}

// BuildCreateAccountQuery takes an account and returns a creation query for that account and the relevant arguments.
func (q *Postgres) BuildCreateAccountQuery(input *types.AccountCreationInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(queriers.AccountsTableName).
		Columns(
			queriers.AccountsTableNameColumn,
			queriers.AccountsTableUserOwnershipColumn,
		).
		Values(
			input.Name,
			input.BelongsToUser,
		).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.IDColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildUpdateAccountQuery takes an account and returns an update SQL query, with the relevant query parameters.
func (q *Postgres) BuildUpdateAccountQuery(input *types.Account) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.AccountsTableName).
		Set(queriers.AccountsTableNameColumn, input.Name).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                         input.ID,
			queriers.AccountsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildArchiveAccountQuery returns a SQL query which marks a given account belonging to a given user as archived.
func (q *Postgres) BuildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.AccountsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                         accountID,
			queriers.ArchivedOnColumn:                 nil,
			queriers.AccountsTableUserOwnershipColumn: userID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// BuildGetAuditLogEntriesForAccountQuery constructs a SQL query for fetching audit log entries
// associated with a given account.
func (q *Postgres) BuildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{}) {
	accountIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.AccountAssignmentKey)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{accountIDKey: accountID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	return q.buildQuery(builder)
}

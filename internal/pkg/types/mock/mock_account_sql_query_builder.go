package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountSQLQueryBuilder = (*AccountSQLQueryBuilder)(nil)

// AccountSQLQueryBuilder is a mocked types.AccountSQLQueryBuilder for testing.
type AccountSQLQueryBuilder struct {
	mock.Mock
}

func (m *AccountSQLQueryBuilder) BuildGetAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(accountID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AccountSQLQueryBuilder) BuildGetAllAccountsCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

func (m *AccountSQLQueryBuilder) BuildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AccountSQLQueryBuilder) BuildGetAccountsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(userID, forAdmin, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AccountSQLQueryBuilder) BuildCreateAccountQuery(input *types.Account) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AccountSQLQueryBuilder) BuildUpdateAccountQuery(input *types.Account) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AccountSQLQueryBuilder) BuildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(accountID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AccountSQLQueryBuilder) BuildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

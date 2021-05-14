package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ querybuilding.AccountSQLQueryBuilder = (*AccountSQLQueryBuilder)(nil)

// AccountSQLQueryBuilder is a mocked types.AccountSQLQueryBuilder for testing.
type AccountSQLQueryBuilder struct {
	mock.Mock
}

// BuildTransferAccountOwnershipQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildTransferAccountOwnershipQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, currentOwnerID, newOwnerID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAccountQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildGetAccountQuery(ctx context.Context, accountID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllAccountsCountQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildGetAllAccountsCountQuery(ctx context.Context) string {
	returnArgs := m.Called(ctx)

	return returnArgs.String(0)
}

// BuildGetBatchOfAccountsQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildGetBatchOfAccountsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAccountsQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildGetAccountsQuery(ctx context.Context, userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, forAdmin, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildAccountCreationQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildAccountCreationQuery(ctx context.Context, input *types.AccountCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateAccountQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildUpdateAccountQuery(ctx context.Context, input *types.Account) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveAccountQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildArchiveAccountQuery(ctx context.Context, accountID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForAccountQuery implements our interface.
func (m *AccountSQLQueryBuilder) BuildGetAuditLogEntriesForAccountQuery(ctx context.Context, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

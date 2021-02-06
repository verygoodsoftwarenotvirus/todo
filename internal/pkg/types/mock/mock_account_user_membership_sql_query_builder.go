package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountUserMembershipSQLQueryBuilder = (*AccountUserMembershipSQLQueryBuilder)(nil)

// AccountUserMembershipSQLQueryBuilder is a mocked types.AccountUserMembershipSQLQueryBuilder for testing.
type AccountUserMembershipSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetAccountUserMembershipQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetAccountUserMembershipQuery(itemID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(itemID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllAccountUserMembershipCountQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetAllAccountUserMembershipsCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

// BuildGetBatchOfAccountUserMembershipsQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetBatchOfAccountUserMembershipsQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAccountUserMembershipsQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetAccountUserMembershipsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(userID, forAdmin, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateAccountUserMembershipQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildCreateAccountUserMembershipQuery(input *types.AccountUserMembershipCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForAccountUserMembershipQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetAuditLogEntriesForAccountUserMembershipQuery(itemID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(itemID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveAccountUserMembershipQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildArchiveAccountUserMembershipQuery(itemID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(itemID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

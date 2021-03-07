package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountUserMembershipSQLQueryBuilder = (*AccountUserMembershipSQLQueryBuilder)(nil)

// AccountUserMembershipSQLQueryBuilder is a mocked types.AccountUserMembershipSQLQueryBuilder for testing.
type AccountUserMembershipSQLQueryBuilder struct {
	mock.Mock
}

// BuildArchiveAccountMembershipsForUserQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildArchiveAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAccountMembershipsForUserQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildMarkAccountAsUserDefaultQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildMarkAccountAsUserDefaultQuery(userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateMembershipForNewUserQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildCreateMembershipForNewUserQuery(userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForAccountUserMembershipQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetAuditLogEntriesForAccountUserMembershipQuery(accountUserMembershipID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(accountUserMembershipID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUserIsMemberOfAccountQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildAddUserToAccountQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildAddUserToAccountQuery(accountID uint64, input *types.AddUserToAccountInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildRemoveUserFromAccountQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildTransferAccountOwnershipQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildTransferAccountOwnershipQuery(oldOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(oldOwnerID, newOwnerID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildModifyUserPermissionsQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildModifyUserPermissionsQuery(userID, accountID uint64, perms permissions.ServiceUserPermissions) (query string, args []interface{}) {
	returnArgs := m.Called(userID, accountID, perms)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.UserSQLQueryBuilder = (*UserSQLQueryBuilder)(nil)

// UserSQLQueryBuilder is a mocked types.UserSQLQueryBuilder for testing.
type UserSQLQueryBuilder struct {
	mock.Mock
}

// BuildUserHasStatusQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUserHasStatusQuery(userID uint64, statuses ...string) (query string, args []interface{}) {
	returnArgs := m.Called(userID, statuses)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUserWithUnverifiedTwoFactorSecretQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildSearchForUserByUsernameQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	returnArgs := m.Called(usernameQuery)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllUsersCountQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetAllUsersCountQuery() (query string) {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

// BuildCreateUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildCreateUserQuery(input types.UserDataStoreCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateUserPasswordQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	returnArgs := m.Called(userID, newHash)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateUserTwoFactorSecretQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUpdateUserTwoFactorSecretQuery(userID uint64, newSecret string) (query string, args []interface{}) {
	returnArgs := m.Called(userID, newSecret)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildVerifyUserTwoFactorSecretQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildSetUserStatusQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildSetUserStatusQuery(userID uint64, input types.UserReputationUpdateInput) (query string, args []interface{}) {
	returnArgs := m.Called(userID, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUsersQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUsersQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUserByUsernameQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUserByUsernameQuery(username string) (query string, args []interface{}) {
	returnArgs := m.Called(username)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

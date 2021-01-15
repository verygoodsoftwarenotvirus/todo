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

func (m *UserSQLQueryBuilder) BuildGetUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{}) {
	returnArgs := m.Called(usernameQuery)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildGetAllUsersCountQuery() (query string) {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

func (m *UserSQLQueryBuilder) BuildCreateUserQuery(input types.UserDataStoreCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildUpdateUserQuery(input *types.User) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{}) {
	returnArgs := m.Called(userID, newHash)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildArchiveUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(userID)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

func (m *UserSQLQueryBuilder) BuildSetUserStatusQuery(userID uint64, input types.AccountStatusUpdateInput) (query string, args []interface{}) {
	returnArgs := m.Called(userID, input)

	return returnArgs.String(0), returnArgs.Get(0).([]interface{})
}

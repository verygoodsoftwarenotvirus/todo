package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ querybuilding.UserSQLQueryBuilder = (*UserSQLQueryBuilder)(nil)

// UserSQLQueryBuilder is a mocked types.UserSQLQueryBuilder for testing.
type UserSQLQueryBuilder struct {
	mock.Mock
}

// BuildUserHasStatusQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUserHasStatusQuery(ctx context.Context, userID string, statuses ...string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, statuses)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUserQuery(ctx context.Context, userID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUserWithUnverifiedTwoFactorSecretQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx context.Context, userID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildSearchForUserByUsernameQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildSearchForUserByUsernameQuery(ctx context.Context, usernameQuery string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, usernameQuery)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllUsersCountQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetAllUsersCountQuery(ctx context.Context) (query string) {
	returnArgs := m.Called(ctx)

	return returnArgs.String(0)
}

// BuildUserCreationQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUserCreationQuery(ctx context.Context, input *types.UserDataStoreCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUpdateUserQuery(ctx context.Context, input *types.User) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateUserPasswordQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUpdateUserPasswordQuery(ctx context.Context, userID, newHash string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, newHash)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateUserTwoFactorSecretQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildUpdateUserTwoFactorSecretQuery(ctx context.Context, userID, newSecret string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, newSecret)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildVerifyUserTwoFactorSecretQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildVerifyUserTwoFactorSecretQuery(ctx context.Context, userID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveUserQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildArchiveUserQuery(ctx context.Context, userID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildSetUserStatusQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildSetUserStatusQuery(ctx context.Context, input *types.UserReputationUpdateInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUsersQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUsersQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetUserByUsernameQuery implements our interface.
func (m *UserSQLQueryBuilder) BuildGetUserByUsernameQuery(ctx context.Context, username string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, username)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

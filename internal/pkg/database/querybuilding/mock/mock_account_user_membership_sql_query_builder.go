package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ querybuilding.AccountUserMembershipSQLQueryBuilder = (*AccountUserMembershipSQLQueryBuilder)(nil)

// AccountUserMembershipSQLQueryBuilder is a mocked types.AccountUserMembershipSQLQueryBuilder for testing.
type AccountUserMembershipSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetDefaultAccountIDForUserQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetDefaultAccountIDForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildTransferAccountMembershipsQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildTransferAccountMembershipsQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, currentOwnerID, newOwnerID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveAccountMembershipsForUserQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildArchiveAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAccountMembershipsForUserQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildGetAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildMarkAccountAsUserDefaultQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildMarkAccountAsUserDefaultQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateMembershipForNewUserQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildCreateMembershipForNewUserQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUserIsMemberOfAccountQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildUserIsMemberOfAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildAddUserToAccountQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildAddUserToAccountQuery(ctx context.Context, accountID uint64, input *types.AddUserToAccountInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountID, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildRemoveUserFromAccountQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildRemoveUserFromAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildModifyUserPermissionsQuery implements our interface.
func (m *AccountUserMembershipSQLQueryBuilder) BuildModifyUserPermissionsQuery(ctx context.Context, userID, accountID uint64, perms permissions.ServiceUserPermissions) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, accountID, perms)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.UserDataManager = (*UserDataManager)(nil)

// UserDataManager is a mocked types.UserDataManager for testing.
type UserDataManager struct {
	mock.Mock
}

// UserHasStatus is a mock function.
func (m *UserDataManager) UserHasStatus(ctx context.Context, userID uint64, statuses ...string) (bool, error) {
	args := m.Called(ctx, userID, statuses)

	return args.Bool(0), args.Error(1)
}

// GetUser is a mock function.
func (m *UserDataManager) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*types.User), args.Error(1)
}

// GetUserWithUnverifiedTwoFactorSecret is a mock function.
func (m *UserDataManager) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*types.User), args.Error(1)
}

// MarkUserTwoFactorSecretAsVerified is a mock function.
func (m *UserDataManager) MarkUserTwoFactorSecretAsVerified(ctx context.Context, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// GetUserByUsername is a mock function.
func (m *UserDataManager) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*types.User), args.Error(1)
}

// SearchForUsersByUsername is a mock function.
func (m *UserDataManager) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*types.User, error) {
	args := m.Called(ctx, usernameQuery)
	return args.Get(0).([]*types.User), args.Error(1)
}

// GetAllUsersCount is a mock function.
func (m *UserDataManager) GetAllUsersCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetUsers is a mock function.
func (m *UserDataManager) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*types.UserList), args.Error(1)
}

// CreateUser is a mock function.
func (m *UserDataManager) CreateUser(ctx context.Context, input *types.UserDataStoreCreationInput) (*types.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.User), args.Error(1)
}

// UpdateUser is a mock function.
func (m *UserDataManager) UpdateUser(ctx context.Context, updated *types.User, changes []*types.FieldChangeSummary) error {
	return m.Called(ctx, updated, changes).Error(0)
}

// UpdateUserPassword is a mock function.
func (m *UserDataManager) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	return m.Called(ctx, userID, newHash).Error(0)
}

// ArchiveUser is a mock function.
func (m *UserDataManager) ArchiveUser(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}

// GetAuditLogEntriesForUser is a mock function.
func (m *UserDataManager) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}

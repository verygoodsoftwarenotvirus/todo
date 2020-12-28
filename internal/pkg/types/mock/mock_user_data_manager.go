package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.UserDataManager = (*UserDataManager)(nil)

// UserDataManager is a mocked types.UserDataManager for testing.
type UserDataManager struct {
	mock.Mock
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

// VerifyUserTwoFactorSecret is a mock function.
func (m *UserDataManager) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// GetUserByUsername is a mock function.
func (m *UserDataManager) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*types.User), args.Error(1)
}

// SearchForUsersByUsername is a mock function.
func (m *UserDataManager) SearchForUsersByUsername(ctx context.Context, query string) ([]types.User, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]types.User), args.Error(1)
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
func (m *UserDataManager) CreateUser(ctx context.Context, input types.UserDataStoreCreationInput) (*types.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.User), args.Error(1)
}

// UpdateUser is a mock function.
func (m *UserDataManager) UpdateUser(ctx context.Context, updated *types.User) error {
	return m.Called(ctx, updated).Error(0)
}

// UpdateUserPassword is a mock function.
func (m *UserDataManager) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	return m.Called(ctx, userID, newHash).Error(0)
}

// ArchiveUser is a mock function.
func (m *UserDataManager) ArchiveUser(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}

// LogUserCreationEvent implements our interface.
func (m *AuditLogDataManager) LogUserCreationEvent(ctx context.Context, user *types.User) {
	m.Called(ctx, user)
}

// LogUserVerifyTwoFactorSecretEvent implements our interface.
func (m *AuditLogDataManager) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUserUpdateTwoFactorSecretEvent implements our interface.
func (m *AuditLogDataManager) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUserUpdatePasswordEvent implements our interface.
func (m *AuditLogDataManager) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUserArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// GetAuditLogEntriesForUser is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}

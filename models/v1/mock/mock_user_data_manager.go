package mock

import (
	"context"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.UserDataManager = (*UserDataManager)(nil)

// UserDataManager is a mocked models.UserDataManager for testing
type UserDataManager struct {
	mock.Mock
}

// GetUser is a mock function
func (m *UserDataManager) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.User), args.Error(1)
}

// GetUserByUsername is a mock function
func (m *UserDataManager) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.User), args.Error(1)
}

// GetUserCount is a mock function
func (m *UserDataManager) GetUserCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(uint64), args.Error(1)
}

// GetUsers is a mock function
func (m *UserDataManager) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.UserList), args.Error(1)
}

// CreateUser is a mock function
func (m *UserDataManager) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.User), args.Error(1)
}

// UpdateUser is a mock function
func (m *UserDataManager) UpdateUser(ctx context.Context, updated *models.User) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveUser is a mock function
func (m *UserDataManager) ArchiveUser(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}

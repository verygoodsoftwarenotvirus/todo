package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ types.APIClientDataManager = (*APIClientDataManager)(nil)

// APIClientDataManager is a mocked types.APIClientDataManager for testing.
type APIClientDataManager struct {
	mock.Mock
}

// GetAPIClientByClientID is a mock function.
func (m *APIClientDataManager) GetAPIClientByClientID(ctx context.Context, clientID string) (*types.APIClient, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).(*types.APIClient), args.Error(1)
}

// GetAPIClientByDatabaseID is a mock function.
func (m *APIClientDataManager) GetAPIClientByDatabaseID(ctx context.Context, clientID, userID string) (*types.APIClient, error) {
	args := m.Called(ctx, clientID, userID)
	return args.Get(0).(*types.APIClient), args.Error(1)
}

// GetTotalAPIClientCount is a mock function.
func (m *APIClientDataManager) GetTotalAPIClientCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAPIClients is a mock function.
func (m *APIClientDataManager) GetAPIClients(ctx context.Context, userID string, filter *types.QueryFilter) (*types.APIClientList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*types.APIClientList), args.Error(1)
}

// CreateAPIClient is a mock function.
func (m *APIClientDataManager) CreateAPIClient(ctx context.Context, input *types.APIClientCreationInput) (*types.APIClient, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.APIClient), args.Error(1)
}

// ArchiveAPIClient is a mock function.
func (m *APIClientDataManager) ArchiveAPIClient(ctx context.Context, clientID, accountID string) error {
	return m.Called(ctx, clientID, accountID).Error(0)
}

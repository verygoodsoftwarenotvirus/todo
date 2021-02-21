package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
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
func (m *APIClientDataManager) GetAPIClientByDatabaseID(ctx context.Context, clientID, userID uint64) (*types.APIClient, error) {
	args := m.Called(ctx, clientID, userID)
	return args.Get(0).(*types.APIClient), args.Error(1)
}

// GetTotalAPIClientCount is a mock function.
func (m *APIClientDataManager) GetTotalAPIClientCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllAPIClients is a mock function.
func (m *APIClientDataManager) GetAllAPIClients(ctx context.Context, results chan []*types.APIClient, bucketSize uint16) error {
	return m.Called(ctx, results, bucketSize).Error(0)
}

// GetAPIClients is a mock function.
func (m *APIClientDataManager) GetAPIClients(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.APIClientList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*types.APIClientList), args.Error(1)
}

// CreateAPIClient is a mock function.
func (m *APIClientDataManager) CreateAPIClient(ctx context.Context, input *types.APICientCreationInput, createdByUser uint64) (*types.APIClient, error) {
	args := m.Called(ctx, input, createdByUser)
	return args.Get(0).(*types.APIClient), args.Error(1)
}

// ArchiveAPIClient is a mock function.
func (m *APIClientDataManager) ArchiveAPIClient(ctx context.Context, clientID, accountID, archivedByUser uint64) error {
	return m.Called(ctx, clientID, accountID, archivedByUser).Error(0)
}

// GetAuditLogEntriesForAPIClient is a mock function.
func (m *APIClientDataManager) GetAuditLogEntriesForAPIClient(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}

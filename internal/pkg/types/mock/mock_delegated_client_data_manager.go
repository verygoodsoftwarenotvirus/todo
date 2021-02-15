package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.DelegatedClientDataManager = (*DelegatedClientDataManager)(nil)

// DelegatedClientDataManager is a mocked types.DelegatedClientDataManager for testing.
type DelegatedClientDataManager struct {
	mock.Mock
}

// GetDelegatedClient is a mock function.
func (m *DelegatedClientDataManager) GetDelegatedClient(ctx context.Context, clientID string) (*types.DelegatedClient, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).(*types.DelegatedClient), args.Error(1)
}

// GetTotalDelegatedClientCount is a mock function.
func (m *DelegatedClientDataManager) GetTotalDelegatedClientCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllDelegatedClients is a mock function.
func (m *DelegatedClientDataManager) GetAllDelegatedClients(ctx context.Context, results chan []*types.DelegatedClient, bucketSize uint16) error {
	return m.Called(ctx, results, bucketSize).Error(0)
}

// GetDelegatedClients is a mock function.
func (m *DelegatedClientDataManager) GetDelegatedClients(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.DelegatedClientList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*types.DelegatedClientList), args.Error(1)
}

// CreateDelegatedClient is a mock function.
func (m *DelegatedClientDataManager) CreateDelegatedClient(ctx context.Context, input *types.DelegatedClientCreationInput) (*types.DelegatedClient, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.DelegatedClient), args.Error(1)
}

// UpdateDelegatedClient is a mock function.
func (m *DelegatedClientDataManager) UpdateDelegatedClient(ctx context.Context, updated *types.DelegatedClient, changes []types.FieldChangeSummary) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveDelegatedClient is a mock function.
func (m *DelegatedClientDataManager) ArchiveDelegatedClient(ctx context.Context, clientID, userID uint64) error {
	return m.Called(ctx, clientID, userID).Error(0)
}

// GetAuditLogEntriesForDelegatedClient is a mock function.
func (m *DelegatedClientDataManager) GetAuditLogEntriesForDelegatedClient(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}

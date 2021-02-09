package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.OAuth2ClientDataManager = (*OAuth2ClientDataManager)(nil)

// OAuth2ClientDataManager is a mocked types.OAuth2ClientDataManager for testing.
type OAuth2ClientDataManager struct {
	mock.Mock
}

// GetOAuth2Client is a mock function.
func (m *OAuth2ClientDataManager) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*types.OAuth2Client, error) {
	args := m.Called(ctx, clientID, userID)
	return args.Get(0).(*types.OAuth2Client), args.Error(1)
}

// GetOAuth2ClientByClientID is a mock function.
func (m *OAuth2ClientDataManager) GetOAuth2ClientByClientID(ctx context.Context, identifier string) (*types.OAuth2Client, error) {
	args := m.Called(ctx, identifier)
	return args.Get(0).(*types.OAuth2Client), args.Error(1)
}

// GetTotalOAuth2ClientCount is a mock function.
func (m *OAuth2ClientDataManager) GetTotalOAuth2ClientCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllOAuth2Clients is a mock function.
func (m *OAuth2ClientDataManager) GetAllOAuth2Clients(ctx context.Context, results chan []*types.OAuth2Client, bucketSize uint16) error {
	return m.Called(ctx, results, bucketSize).Error(0)
}

// GetOAuth2Clients is a mock function.
func (m *OAuth2ClientDataManager) GetOAuth2Clients(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.OAuth2ClientList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*types.OAuth2ClientList), args.Error(1)
}

// CreateOAuth2Client is a mock function.
func (m *OAuth2ClientDataManager) CreateOAuth2Client(ctx context.Context, input *types.OAuth2ClientCreationInput) (*types.OAuth2Client, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.OAuth2Client), args.Error(1)
}

// UpdateOAuth2Client is a mock function.
func (m *OAuth2ClientDataManager) UpdateOAuth2Client(ctx context.Context, updated *types.OAuth2Client) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveOAuth2Client is a mock function.
func (m *OAuth2ClientDataManager) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	return m.Called(ctx, clientID, userID).Error(0)
}

// GetAuditLogEntriesForOAuth2Client is a mock function.
func (m *OAuth2ClientDataManager) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}

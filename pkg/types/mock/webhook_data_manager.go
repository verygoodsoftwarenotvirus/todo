package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.WebhookDataManager = (*WebhookDataManager)(nil)

// WebhookDataManager is a mocked types.WebhookDataManager for testing.
type WebhookDataManager struct {
	mock.Mock
}

// GetWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetWebhook(ctx context.Context, webhookID, userID uint64) (*types.Webhook, error) {
	args := m.Called(ctx, webhookID, userID)
	return args.Get(0).(*types.Webhook), args.Error(1)
}

// GetAllWebhooksCount satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetAllWebhooksCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetWebhooks satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetWebhooks(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.WebhookList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*types.WebhookList), args.Error(1)
}

// GetAllWebhooks satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetAllWebhooks(ctx context.Context, results chan []*types.Webhook, bucketSize uint16) error {
	return m.Called(ctx, results, bucketSize).Error(0)
}

// CreateWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput, createdByUser uint64) (*types.Webhook, error) {
	args := m.Called(ctx, input, createdByUser)
	return args.Get(0).(*types.Webhook), args.Error(1)
}

// UpdateWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) UpdateWebhook(ctx context.Context, updated *types.Webhook, changedByUser uint64, changes []*types.FieldChangeSummary) error {
	return m.Called(ctx, updated, changedByUser, changes).Error(0)
}

// ArchiveWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) ArchiveWebhook(ctx context.Context, webhookID, accountID, archivedByUserID uint64) error {
	return m.Called(ctx, webhookID, accountID, archivedByUserID).Error(0)
}

// GetAuditLogEntriesForWebhook is a mock function.
func (m *WebhookDataManager) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, webhookID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}

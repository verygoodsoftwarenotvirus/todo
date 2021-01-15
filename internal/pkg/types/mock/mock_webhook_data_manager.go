package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

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
func (m *WebhookDataManager) GetAllWebhooks(ctx context.Context, results chan []types.Webhook, bucketSize uint16) error {
	return m.Called(ctx, results, bucketSize).Error(0)
}

// CreateWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) CreateWebhook(ctx context.Context, input *types.WebhookCreationInput) (*types.Webhook, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.Webhook), args.Error(1)
}

// UpdateWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) UpdateWebhook(ctx context.Context, updated *types.Webhook) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	return m.Called(ctx, webhookID, userID).Error(0)
}

// LogWebhookCreationEvent implements our interface.
func (m *AuditLogEntryDataManager) LogWebhookCreationEvent(ctx context.Context, webhook *types.Webhook) {
	m.Called(ctx, webhook)
}

// LogWebhookUpdateEvent implements our interface.
func (m *AuditLogEntryDataManager) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []types.FieldChangeSummary) {
	m.Called(ctx, userID, webhookID, changes)
}

// LogWebhookArchiveEvent implements our interface.
func (m *AuditLogEntryDataManager) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	m.Called(ctx, userID, webhookID)
}

// GetAuditLogEntriesForWebhook is a mock function.
func (m *AuditLogEntryDataManager) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, webhookID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}

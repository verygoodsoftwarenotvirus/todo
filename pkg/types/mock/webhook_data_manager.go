package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ types.WebhookDataManager = (*WebhookDataManager)(nil)

// WebhookDataManager is a mocked types.WebhookDataManager for testing.
type WebhookDataManager struct {
	mock.Mock
}

// WebhookExists satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error) {
	args := m.Called(ctx, webhookID, accountID)
	return args.Bool(0), args.Error(1)
}

// GetWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetWebhook(ctx context.Context, webhookID, accountID string) (*types.Webhook, error) {
	args := m.Called(ctx, webhookID, accountID)
	return args.Get(0).(*types.Webhook), args.Error(1)
}

// GetAllWebhooksCount satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetAllWebhooksCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetWebhooks satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetWebhooks(ctx context.Context, accountID string, filter *types.QueryFilter) (*types.WebhookList, error) {
	args := m.Called(ctx, accountID, filter)
	return args.Get(0).(*types.WebhookList), args.Error(1)
}

// GetAllWebhooks satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) GetAllWebhooks(ctx context.Context, results chan []*types.Webhook, bucketSize uint16) error {
	return m.Called(ctx, results, bucketSize).Error(0)
}

// CreateWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) CreateWebhook(ctx context.Context, input *types.WebhookDatabaseCreationInput) (*types.Webhook, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.Webhook), args.Error(1)
}

// UpdateWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) UpdateWebhook(ctx context.Context, updated *types.Webhook) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveWebhook satisfies our WebhookDataManager interface.
func (m *WebhookDataManager) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	return m.Called(ctx, webhookID, accountID).Error(0)
}

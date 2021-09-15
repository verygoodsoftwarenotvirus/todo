package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ querybuilding.WebhookSQLQueryBuilder = (*WebhookSQLQueryBuilder)(nil)

// WebhookSQLQueryBuilder is a mocked types.WebhookSQLQueryBuilder for testing.
type WebhookSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetWebhookQuery implements our interface.
func (m *WebhookSQLQueryBuilder) BuildGetWebhookQuery(ctx context.Context, webhookID, accountID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, webhookID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllWebhooksCountQuery implements our interface.
func (m *WebhookSQLQueryBuilder) BuildGetAllWebhooksCountQuery(ctx context.Context) string {
	return m.Called(ctx).String(0)
}

// BuildGetBatchOfWebhooksQuery implements our interface.
func (m *WebhookSQLQueryBuilder) BuildGetBatchOfWebhooksQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetWebhooksQuery implements our interface.
func (m *WebhookSQLQueryBuilder) BuildGetWebhooksQuery(ctx context.Context, accountID string, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountID, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateWebhookQuery implements our interface.
func (m *WebhookSQLQueryBuilder) BuildCreateWebhookQuery(ctx context.Context, x *types.WebhookCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, x)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateWebhookQuery implements our interface.
func (m *WebhookSQLQueryBuilder) BuildUpdateWebhookQuery(ctx context.Context, input *types.Webhook) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveWebhookQuery implements our interface.
func (m *WebhookSQLQueryBuilder) BuildArchiveWebhookQuery(ctx context.Context, webhookID, accountID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, webhookID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

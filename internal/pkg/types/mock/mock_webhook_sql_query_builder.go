package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.WebhookSQLQueryBuilder = (*WebhookSQLQueryBuilder)(nil)

// WebhookSQLQueryBuilder is a mocked types.WebhookSQLQueryBuilder for testing.
type WebhookSQLQueryBuilder struct {
	mock.Mock
}

func (m *WebhookSQLQueryBuilder) BuildGetWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(webhookID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *WebhookSQLQueryBuilder) BuildGetAllWebhooksCountQuery() string {
	return m.Called().String(0)
}

func (m *WebhookSQLQueryBuilder) BuildGetBatchOfWebhooksQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *WebhookSQLQueryBuilder) BuildGetWebhooksQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(userID, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *WebhookSQLQueryBuilder) BuildCreateWebhookQuery(x *types.Webhook) (query string, args []interface{}) {
	returnArgs := m.Called(x)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *WebhookSQLQueryBuilder) BuildUpdateWebhookQuery(input *types.Webhook) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *WebhookSQLQueryBuilder) BuildArchiveWebhookQuery(webhookID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(webhookID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *WebhookSQLQueryBuilder) BuildGetAuditLogEntriesForWebhookQuery(webhookID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(webhookID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

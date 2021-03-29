package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ querybuilding.APIClientSQLQueryBuilder = (*APIClientSQLQueryBuilder)(nil)

// APIClientSQLQueryBuilder is a mocked types.APIClientSQLQueryBuilder for testing.
type APIClientSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetBatchOfAPIClientsQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetBatchOfAPIClientsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAPIClientByClientIDQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAPIClientByClientIDQuery(ctx context.Context, clientID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAPIClientByDatabaseIDQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAPIClientByDatabaseIDQuery(ctx context.Context, clientID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, clientID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllAPIClientsCountQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAllAPIClientsCountQuery(ctx context.Context) string {
	returnArgs := m.Called(ctx)

	return returnArgs.String(0)
}

// BuildGetAPIClientsQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAPIClientsQuery(ctx context.Context, userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, userID, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildCreateAPIClientQuery(ctx context.Context, input *types.APIClientCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildUpdateAPIClientQuery(ctx context.Context, input *types.APIClient) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildArchiveAPIClientQuery(ctx context.Context, clientID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, clientID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAuditLogEntriesForAPIClientQuery(ctx context.Context, clientID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

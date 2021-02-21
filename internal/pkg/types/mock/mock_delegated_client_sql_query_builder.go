package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.APIClientSQLQueryBuilder = (*APIClientSQLQueryBuilder)(nil)

// APIClientSQLQueryBuilder is a mocked types.APIClientSQLQueryBuilder for testing.
type APIClientSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetBatchOfAPIClientsQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetBatchOfAPIClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAPIClientByClientIDQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAPIClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	returnArgs := m.Called(clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAPIClientByDatabaseIDQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAPIClientByDatabaseIDQuery(clientID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllAPIClientsCountQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAllAPIClientsCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

// BuildGetAPIClientsQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAPIClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(userID, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildCreateAPIClientQuery(input *types.APICientCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildUpdateAPIClientQuery(input *types.APIClient) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildArchiveAPIClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForAPIClientQuery implements our interface.
func (m *APIClientSQLQueryBuilder) BuildGetAuditLogEntriesForAPIClientQuery(clientID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

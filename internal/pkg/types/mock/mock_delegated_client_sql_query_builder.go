package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.DelegatedClientSQLQueryBuilder = (*DelegatedClientSQLQueryBuilder)(nil)

// DelegatedClientSQLQueryBuilder is a mocked types.DelegatedClientSQLQueryBuilder for testing.
type DelegatedClientSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetBatchOfDelegatedClientsQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildGetBatchOfDelegatedClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetDelegatedClientQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildGetDelegatedClientQuery(clientID string) (query string, args []interface{}) {
	returnArgs := m.Called(clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllDelegatedClientsCountQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildGetAllDelegatedClientsCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

// BuildGetDelegatedClientsQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildGetDelegatedClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(userID, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateDelegatedClientQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildCreateDelegatedClientQuery(input *types.DelegatedClientCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateDelegatedClientQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildUpdateDelegatedClientQuery(input *types.DelegatedClient) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveDelegatedClientQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildArchiveDelegatedClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForDelegatedClientQuery implements our interface.
func (m *DelegatedClientSQLQueryBuilder) BuildGetAuditLogEntriesForDelegatedClientQuery(clientID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

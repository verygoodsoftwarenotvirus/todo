package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AuditLogEntrySQLQueryBuilder = (*AuditLogEntrySQLQueryBuilder)(nil)

// AuditLogEntrySQLQueryBuilder is a mocked types.AuditLogEntrySQLQueryBuilder for testing.
type AuditLogEntrySQLQueryBuilder struct {
	mock.Mock
}

func (m *AuditLogEntrySQLQueryBuilder) BuildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(entryID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AuditLogEntrySQLQueryBuilder) BuildGetAllAuditLogEntriesCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

func (m *AuditLogEntrySQLQueryBuilder) BuildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AuditLogEntrySQLQueryBuilder) BuildGetAuditLogEntriesQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

func (m *AuditLogEntrySQLQueryBuilder) BuildCreateAuditLogEntryQuery(input *types.AuditLogEntry) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

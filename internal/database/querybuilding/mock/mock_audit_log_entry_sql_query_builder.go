package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ querybuilding.AuditLogEntrySQLQueryBuilder = (*AuditLogEntrySQLQueryBuilder)(nil)

// AuditLogEntrySQLQueryBuilder is a mocked types.AuditLogEntrySQLQueryBuilder for testing.
type AuditLogEntrySQLQueryBuilder struct {
	mock.Mock
}

// BuildGetAuditLogEntryQuery implements our interface.
func (m *AuditLogEntrySQLQueryBuilder) BuildGetAuditLogEntryQuery(ctx context.Context, entryID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, entryID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllAuditLogEntriesCountQuery implements our interface.
func (m *AuditLogEntrySQLQueryBuilder) BuildGetAllAuditLogEntriesCountQuery(ctx context.Context) string {
	returnArgs := m.Called(ctx)

	return returnArgs.String(0)
}

// BuildGetBatchOfAuditLogEntriesQuery implements our interface.
func (m *AuditLogEntrySQLQueryBuilder) BuildGetBatchOfAuditLogEntriesQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesQuery implements our interface.
func (m *AuditLogEntrySQLQueryBuilder) BuildGetAuditLogEntriesQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateAuditLogEntryQuery implements our interface.
func (m *AuditLogEntrySQLQueryBuilder) BuildCreateAuditLogEntryQuery(ctx context.Context, input *types.AuditLogEntryCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

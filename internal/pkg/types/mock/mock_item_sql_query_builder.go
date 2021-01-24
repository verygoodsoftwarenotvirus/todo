package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.ItemSQLQueryBuilder = (*ItemSQLQueryBuilder)(nil)

// ItemSQLQueryBuilder is a mocked types.ItemSQLQueryBuilder for testing.
type ItemSQLQueryBuilder struct {
	mock.Mock
}

// BuildItemExistsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildItemExistsQuery(itemID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(itemID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(itemID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllItemsCountQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetAllItemsCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

// BuildGetBatchOfItemsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetBatchOfItemsQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetItemsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetItemsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(userID, forAdmin, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetItemsWithIDsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetItemsWithIDsQuery(userID uint64, limit uint8, ids []uint64, forAdmin bool) (query string, args []interface{}) {
	returnArgs := m.Called(userID, limit, ids, forAdmin)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildCreateItemQuery(input *types.ItemCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(itemID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildUpdateItemQuery(input *types.Item) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildArchiveItemQuery(itemID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(itemID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

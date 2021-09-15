package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ querybuilding.ItemSQLQueryBuilder = (*ItemSQLQueryBuilder)(nil)

// ItemSQLQueryBuilder is a mocked types.ItemSQLQueryBuilder for testing.
type ItemSQLQueryBuilder struct {
	mock.Mock
}

// BuildItemExistsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildItemExistsQuery(ctx context.Context, itemID, accountID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, itemID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetItemQuery(ctx context.Context, itemID, accountID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, itemID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetTotalItemCountQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetTotalItemCountQuery(ctx context.Context) string {
	returnArgs := m.Called(ctx)

	return returnArgs.String(0)
}

// BuildGetBatchOfItemsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetBatchOfItemsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetItemsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetItemsQuery(ctx context.Context, accountID string, includeArchived bool, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountID, includeArchived, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetItemsWithIDsQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildGetItemsWithIDsQuery(ctx context.Context, accountID string, limit uint8, ids []string, restrictToAccount bool) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountID, limit, ids, restrictToAccount)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildCreateItemQuery(ctx context.Context, input *types.ItemDatabaseCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildUpdateItemQuery(ctx context.Context, input *types.Item) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveItemQuery implements our interface.
func (m *ItemSQLQueryBuilder) BuildArchiveItemQuery(ctx context.Context, itemID, accountID string) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, itemID, accountID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

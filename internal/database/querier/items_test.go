package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromItems(includeCounts bool, filteredCount uint64, items ...*types.Item) *sqlmock.Rows {
	columns := querybuilding.ItemsTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range items {
		rowValues := []driver.Value{
			x.ID,
			x.ExternalID,
			x.Name,
			x.Details,
			x.CreatedOn,
			x.LastUpdatedOn,
			x.ArchivedOn,
			x.BelongsToAccount,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(items))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanItems(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanItems(ctx, mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanItems(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_ItemExists(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildItemExistsQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := c.ItemExists(ctx, exampleItem.ID, exampleAccount.ID)
		assert.NoError(t, err)
		assert.True(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.ItemExists(ctx, 0, exampleAccount.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		actual, err := c.ItemExists(ctx, exampleItem.ID, 0)
		assert.Error(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildItemExistsQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.ItemExists(ctx, exampleItem.ID, exampleAccount.ID)
		assert.NoError(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildItemExistsQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.ItemExists(ctx, exampleItem.ID, exampleAccount.ID)
		assert.Error(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(false, 0, exampleItem))

		actual, err := c.GetItem(ctx, exampleItem.ID, exampleAccount.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetItem(ctx, 0, exampleItem.BelongsToAccount)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetItem(ctx, exampleItem.ID, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetItem(ctx, exampleItem.ID, exampleAccount.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAllItemsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(uint64(123)))

		actual, err := c.GetAllItemsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAllItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		results := make(chan []*types.Item)
		doneChan := make(chan bool, 1)
		expectedCount := uint64(20)
		exampleItemList := fakes.BuildFakeItemList()
		exampleBatchSize := uint16(1000)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAllItemsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetBatchOfItemsQuery",
			testutils.ContextMatcher,
			uint64(1),
			uint64(exampleBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(false, 0, exampleItemList.Items...))

		assert.NoError(t, c.GetAllItems(ctx, results, exampleBatchSize))

		stillQuerying := true
		for stillQuerying {
			select {
			case batch := <-results:
				assert.NotEmpty(t, batch)
				doneChan <- true
			case <-time.After(time.Second):
				t.FailNow()
			case <-doneChan:
				stillQuerying = false
			}
		}

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil results channel", func(t *testing.T) {
		t.Parallel()

		exampleBatchSize := uint16(1000)
		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.GetAllItems(ctx, nil, exampleBatchSize))
	})

	T.Run("with now rows returned", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.Item)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAllItemsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetBatchOfItemsQuery",
			testutils.ContextMatcher,
			uint64(1),
			uint64(exampleBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		assert.NoError(t, c.GetAllItems(ctx, results, exampleBatchSize))

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.Item)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAllItemsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllItems(ctx, results, exampleBatchSize)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.Item)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAllItemsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetBatchOfItemsQuery",
			testutils.ContextMatcher,
			uint64(1),
			uint64(exampleBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.NoError(t, c.GetAllItems(ctx, results, exampleBatchSize))

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.Item)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAllItemsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetBatchOfItemsQuery",
			testutils.ContextMatcher,
			uint64(1),
			uint64(exampleBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		assert.NoError(t, c.GetAllItems(ctx, results, exampleBatchSize))

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItemList := fakes.BuildFakeItemList()

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(true, exampleItemList.FilteredCount, exampleItemList.Items...))

		actual, err := c.GetItems(ctx, exampleAccount.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleAccount := fakes.BuildFakeAccount()
		exampleItemList := fakes.BuildFakeItemList()
		exampleItemList.Page = 0
		exampleItemList.Limit = 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(true, exampleItemList.FilteredCount, exampleItemList.Items...))

		actual, err := c.GetItems(ctx, exampleAccount.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetItems(ctx, 0, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccount := fakes.BuildFakeAccount()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetItems(ctx, exampleAccount.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccount := fakes.BuildFakeAccount()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetItems(ctx, exampleAccount.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetItemsWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItemList := fakes.BuildFakeItemList()

		var exampleIDs []uint64
		for _, x := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, x.ID)
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsWithIDsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			defaultLimit,
			exampleIDs,
			false,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(false, 0, exampleItemList.Items...))

		actual, err := c.GetItemsWithIDs(ctx, exampleAccount.ID, defaultLimit, exampleIDs)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList.Items, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, x := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, x.ID)
		}

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetItemsWithIDs(ctx, 0, defaultLimit, exampleIDs)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("sets limit if not present", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, x := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, x.ID)
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsWithIDsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			defaultLimit,
			exampleIDs,
			false,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(false, 0, exampleItemList.Items...))

		actual, err := c.GetItemsWithIDs(ctx, exampleAccount.ID, 0, exampleIDs)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList.Items, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, x := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, x.ID)
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsWithIDsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			defaultLimit,
			exampleIDs,
			false,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetItemsWithIDs(ctx, exampleAccount.ID, defaultLimit, exampleIDs)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, x := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, x.ID)
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetItemsWithIDsQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
			defaultLimit,
			exampleIDs,
			false,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetItemsWithIDs(ctx, exampleAccount.ID, defaultLimit, exampleIDs)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.ExternalID = ""
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildCreateItemQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeCreationQuery, fakeCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleItem.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateItem(ctx, exampleInput, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.ExternalID = ""

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateItem(ctx, nil, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()
		exampleItem.ExternalID = ""
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateItem(ctx, exampleInput, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.ExternalID = ""
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		actual, err := c.CreateItem(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New(t.Name())
		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildCreateItemQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(expectedErr)

		c.timeFunc = func() uint64 {
			return exampleItem.CreatedOn
		}

		actual, err := c.CreateItem(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error creating audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.ExternalID = ""
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildCreateItemQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeCreationQuery, fakeCreationArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		actual, err := c.CreateItem(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.ExternalID = ""
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildCreateItemQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeCreationQuery, fakeCreationArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleItem.CreatedOn
		}

		actual, err := c.CreateItem(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildUpdateItemQuery",
			testutils.ContextMatcher,
			exampleItem,
		).Return(fakeUpdateQuery, fakeUpdateArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectCommit()

		assert.NoError(t, c.UpdateItem(ctx, exampleItem, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateItem(ctx, nil, exampleUser.ID, nil))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateItem(ctx, exampleItem, 0, nil))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateItem(ctx, exampleItem, exampleUser.ID, nil))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildUpdateItemQuery",
			testutils.ContextMatcher,
			exampleItem,
		).Return(fakeUpdateQuery, fakeUpdateArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.UpdateItem(ctx, exampleItem, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildUpdateItemQuery",
			testutils.ContextMatcher,
			exampleItem,
		).Return(fakeUpdateQuery, fakeUpdateArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.UpdateItem(ctx, exampleItem, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildUpdateItemQuery",
			testutils.ContextMatcher,
			exampleItem,
		).Return(fakeUpdateQuery, fakeUpdateArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.UpdateItem(ctx, exampleItem, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildArchiveItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.NoError(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveItem(ctx, 0, exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, 0, exampleUser.ID))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccount.ID, 0))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildArchiveItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildArchiveItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildArchiveItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleItem.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAuditLogEntriesForItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()
		exampleAuditLogEntriesList := fakes.BuildFakeAuditLogEntryList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntriesList.Entries...))

		actual, err := c.GetAuditLogEntriesForItem(ctx, exampleItem.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntriesList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntriesForItem(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForItem(ctx, exampleItem.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.ItemSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForItemQuery",
			testutils.ContextMatcher,
			exampleItem.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForItem(ctx, exampleItem.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

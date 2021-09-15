package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

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

		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := c.ItemExists(ctx, exampleItem.ID, exampleAccountID)
		assert.NoError(t, err)
		assert.True(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()

		c, _ := buildTestClient(t)

		actual, err := c.ItemExists(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()

		c, db := buildTestClient(t)

		actual, err := c.ItemExists(ctx, exampleItem.ID, "")
		assert.Error(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.ItemExists(ctx, exampleItem.ID, exampleAccountID)
		assert.NoError(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.ItemExists(ctx, exampleItem.ID, exampleAccountID)
		assert.Error(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(false, 0, exampleItem))

		actual, err := c.GetItem(ctx, exampleItem.ID, exampleAccountID)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetItem(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetItem(ctx, exampleItem.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetItem(ctx, exampleItem.ID, exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetTotalItemCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(uint64(123)))

		actual, err := c.GetTotalItemCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountID := fakes.BuildFakeID()
		exampleItemList := fakes.BuildFakeItemList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(true, exampleItemList.FilteredCount, exampleItemList.Items...))

		actual, err := c.GetItems(ctx, exampleAccountID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleAccountID := fakes.BuildFakeID()
		exampleItemList := fakes.BuildFakeItemList()
		exampleItemList.Page = 0
		exampleItemList.Limit = 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromItems(true, exampleItemList.FilteredCount, exampleItemList.Items...))

		actual, err := c.GetItems(ctx, exampleAccountID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetItems(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetItems(ctx, exampleAccountID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetItems(ctx, exampleAccountID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()
		exampleItem.ID = "1"
		exampleInput := fakes.BuildFakeItemDatabaseCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleItem.ID))

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleItem.CreatedOn
		}

		actual, err := c.CreateItem(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateItem(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New(t.Name())
		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemDatabaseCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(expectedErr)

		c.timeFunc = func() uint64 {
			return exampleItem.CreatedOn
		}

		actual, err := c.CreateItem(ctx, exampleInput)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error creating audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemDatabaseCreationInputFromItem(exampleItem)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleItem.ID))

		db.ExpectRollback()

		actual, err := c.CreateItem(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleItem.ID))

		db.ExpectCommit()

		assert.NoError(t, c.UpdateItem(ctx, exampleItem))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateItem(ctx, nil))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateItem(ctx, exampleItem))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.UpdateItem(ctx, exampleItem))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry to database", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleItem.ID))

		db.ExpectRollback()

		assert.Error(t, c.UpdateItem(ctx, exampleItem))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleItem.ID))

		db.ExpectCommit()

		assert.NoError(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccountID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid item ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveItem(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, ""))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccountID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleItem.ID))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveItem(ctx, exampleItem.ID, exampleAccountID))

		mock.AssertExpectationsForObjects(t, db)
	})
}

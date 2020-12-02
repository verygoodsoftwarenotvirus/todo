package mariadb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromItems(includeCount bool, items ...*types.Item) *sqlmock.Rows {
	columns := queriers.ItemsTableColumns

	if includeCount {
		columns = append(columns, "count")
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
			x.BelongsToUser,
		}

		if includeCount {
			rowValues = append(rowValues, len(items))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromItem(x *types.Item) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.ItemsTableColumns).AddRow(
		x.ArchivedOn,
		x.Name,
		x.Details,
		x.CreatedOn,
		x.LastUpdatedOn,
		x.BelongsToUser,
		x.ID,
	)

	return exampleRows
}

func TestMariaDB_ScanItems(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := m.scanItems(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := m.scanItems(mockRows, false)
		assert.NoError(t, err)
	})
}

func TestMariaDB_buildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = ? AND items.id = ? )"
		expectedArgs := []interface{}{
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildItemExistsQuery(exampleItem.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_ItemExists(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildItemExistsQuery(exampleItem.ID, exampleItem.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := m.ItemExists(ctx, exampleItem.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.True(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildItemExistsQuery(exampleItem.ID, exampleItem.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.ItemExists(ctx, exampleItem.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.False(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"
		expectedArgs := []interface{}{
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildGetItemQuery(exampleItem.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetItem(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetItemQuery(exampleItem.ID, exampleItem.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromItems(false, exampleItem))

		actual, err := m.GetItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetItemQuery(exampleItem.ID, exampleItem.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		actualQuery := m.buildGetAllItemsCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		expectedCount := uint64(123)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := m.GetAllItemsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetBatchOfItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.id > ? AND items.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := m.buildGetBatchOfItemsQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetAllItems(T *testing.T) {
	T.Parallel()

	m, _ := buildTestService(T)
	expectedCountQuery := m.buildGetAllItemsCountQuery()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		exampleItemList := fakes.BuildFakeItemList()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedGetQuery, expectedArgs := m.buildGetBatchOfItemsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedGetQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromItems(
					false,
					&exampleItemList.Items[0],
					&exampleItemList.Items[1],
					&exampleItemList.Items[2],
				),
			)

		out := make(chan []types.Item)
		doneChan := make(chan bool, 1)

		err := m.GetAllItems(ctx, out)
		assert.NoError(t, err)

		var stillQuerying = true
		for stillQuerying {
			select {
			case batch := <-out:
				assert.NotEmpty(t, batch)
				doneChan <- true
			case <-time.After(time.Second):
				t.FailNow()
			case <-doneChan:
				stillQuerying = false
			}
		}

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		out := make(chan []types.Item)

		err := m.GetAllItems(ctx, out)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedGetQuery, expectedArgs := m.buildGetBatchOfItemsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedGetQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		out := make(chan []types.Item)

		err := m.GetAllItems(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedGetQuery, expectedArgs := m.buildGetBatchOfItemsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedGetQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		out := make(chan []types.Item)

		err := m.GetAllItems(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		exampleItem := fakes.BuildFakeItem()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedGetQuery, expectedArgs := m.buildGetBatchOfItemsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedGetQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		out := make(chan []types.Item)

		err := m.GetAllItems(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user, (SELECT COUNT(*) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ?) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ? ORDER BY items.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := m.buildGetItemsQuery(exampleUser.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		exampleUser := fakes.BuildFakeUser()
		exampleItemList := fakes.BuildFakeItemList()
		expectedQuery, expectedArgs := m.buildGetItemsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromItems(
					true,
					&exampleItemList.Items[0],
					&exampleItemList.Items[1],
					&exampleItemList.Items[2],
				),
			)

		actual, err := m.GetItems(ctx, exampleUser.ID, filter)

		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		exampleUser := fakes.BuildFakeUser()
		expectedQuery, expectedArgs := m.buildGetItemsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItems(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		exampleUser := fakes.BuildFakeUser()
		expectedQuery, expectedArgs := m.buildGetItemsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItems(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery, expectedArgs := m.buildGetItemsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := m.GetItems(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetItemsForAdminQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user, (SELECT COUNT(*) FROM items WHERE items.archived_on IS NULL AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ?) FROM items WHERE items.archived_on IS NULL AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ? ORDER BY items.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := m.buildGetItemsForAdminQuery(filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetItemsForAdmin(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		exampleItemList := fakes.BuildFakeItemList()
		expectedQuery, expectedArgs := m.buildGetItemsForAdminQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromItems(
					true,
					&exampleItemList.Items[0],
					&exampleItemList.Items[1],
					&exampleItemList.Items[2],
				),
			)

		actual, err := m.GetItemsForAdmin(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		expectedQuery, expectedArgs := m.buildGetItemsForAdminQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItemsForAdmin(ctx, filter)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		expectedQuery, expectedArgs := m.buildGetItemsForAdminQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItemsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery, expectedArgs := m.buildGetItemsForAdminQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := m.GetItemsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetItemsWithIDsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleIDs := []uint64{
			789,
			123,
			456,
		}

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.id IN (?,?,?) ORDER BY CASE items.id WHEN 789 THEN 0 WHEN 123 THEN 1 WHEN 456 THEN 2 END LIMIT 20"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleIDs[0],
			exampleIDs[1],
			exampleIDs[2],
		}
		actualQuery, actualArgs := m.buildGetItemsWithIDsQuery(exampleUser.ID, defaultLimit, exampleIDs)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetItemsWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		m, mockDB := buildTestService(t)

		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, item := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, item.ID)
		}

		expectedQuery, expectedArgs := m.buildGetItemsWithIDsQuery(exampleUser.ID, defaultLimit, exampleIDs)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromItems(
					false,
					&exampleItemList.Items[0],
					&exampleItemList.Items[1],
					&exampleItemList.Items[2],
				),
			)

		actual, err := m.GetItemsWithIDs(ctx, exampleUser.ID, defaultLimit, exampleIDs)

		assert.NoError(t, err)
		assert.Equal(t, exampleItemList.Items, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		m, mockDB := buildTestService(t)

		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, item := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, item.ID)
		}

		expectedQuery, expectedArgs := m.buildGetItemsWithIDsQuery(exampleUser.ID, defaultLimit, exampleIDs)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItemsWithIDs(ctx, exampleUser.ID, defaultLimit, exampleIDs)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		m, mockDB := buildTestService(t)

		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, item := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, item.ID)
		}

		expectedQuery, expectedArgs := m.buildGetItemsWithIDsQuery(exampleUser.ID, defaultLimit, exampleIDs)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItemsWithIDs(ctx, exampleUser.ID, defaultLimit, exampleIDs)

		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		m, mockDB := buildTestService(t)

		exampleItemList := fakes.BuildFakeItemList()
		var exampleIDs []uint64
		for _, item := range exampleItemList.Items {
			exampleIDs = append(exampleIDs, item.ID)
		}

		expectedQuery, expectedArgs := m.buildGetItemsWithIDsQuery(exampleUser.ID, defaultLimit, exampleIDs)

		exampleItem := fakes.BuildFakeItem()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := m.GetItemsWithIDs(ctx, exampleUser.ID, defaultLimit, exampleIDs)

		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES (?,?,?)"
		expectedArgs := []interface{}{
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToUser,
		}
		actualQuery, actualArgs := m.buildCreateItemQuery(exampleItem)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		expectedQuery, expectedArgs := m.buildCreateItemQuery(exampleItem)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(int64(exampleItem.ID), 1))

		mtt := &queriers.MockTimeTeller{}
		mtt.On("Now").Return(exampleItem.CreatedOn)
		m.timeTeller = mtt

		actual, err := m.CreateItem(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, mtt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		expectedQuery, expectedArgs := m.buildCreateItemQuery(exampleItem)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := m.CreateItem(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE items SET name = ?, details = ?, last_updated_on = UNIX_TIMESTAMP() WHERE belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildUpdateItemQuery(exampleItem)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery, expectedArgs := m.buildUpdateItemQuery(exampleItem)

		exampleRows := sqlmock.NewResult(int64(exampleItem.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		err := m.UpdateItem(ctx, exampleItem)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery, expectedArgs := m.buildUpdateItemQuery(exampleItem)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := m.UpdateItem(ctx, exampleItem)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE items SET last_updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildArchiveItemQuery(exampleItem.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery, expectedArgs := m.buildArchiveItemQuery(exampleItem.ID, exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("returns sql.ErrNoRows with no rows affected", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery, expectedArgs := m.buildArchiveItemQuery(exampleItem.ID, exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery, expectedArgs := m.buildArchiveItemQuery(exampleItem.ID, exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAuditLogEntriesForItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()

		expectedQuery := fmt.Sprintf("SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE JSON_CONTAINS(audit_log.context, '%d', '$.item_id') ORDER BY audit_log.id", exampleItem.ID)
		expectedArgs := []interface{}(nil)
		actualQuery, actualArgs := m.buildGetAuditLogEntriesForItemQuery(exampleItem.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetAuditLogEntriesForItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		exampleItem := fakes.BuildFakeItem()

		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries
		expectedQuery, expectedArgs := m.buildGetAuditLogEntriesForItemQuery(exampleItem.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromAuditLogEntries(
					false,
					&exampleAuditLogEntryList[0],
					&exampleAuditLogEntryList[1],
					&exampleAuditLogEntryList[2],
				),
			)

		actual, err := m.GetAuditLogEntriesForItem(ctx, exampleItem.ID)

		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		exampleItem := fakes.BuildFakeItem()

		expectedQuery, expectedArgs := p.buildGetAuditLogEntriesForItemQuery(exampleItem.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAuditLogEntriesForItem(ctx, exampleItem.ID)

		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		exampleItem := fakes.BuildFakeItem()

		expectedQuery, expectedArgs := p.buildGetAuditLogEntriesForItemQuery(exampleItem.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildErroneousMockRowFromAuditLogEntry(
					fakes.BuildFakeAuditLogEntry(),
				),
			)

		actual, err := p.GetAuditLogEntriesForItem(ctx, exampleItem.ID)

		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

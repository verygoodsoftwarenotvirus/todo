package mariadb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromItem(items ...*models.Item) *sqlmock.Rows {
	includeCount := len(items) > 1
	columns := itemsTableColumns

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
			x.UpdatedOn,
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

func buildErroneousMockRowFromItem(x *models.Item) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(itemsTableColumns).AddRow(
		x.ArchivedOn,
		x.Name,
		x.Details,
		x.CreatedOn,
		x.UpdatedOn,
		x.BelongsToUser,
		x.ID,
	)

	return exampleRows
}

func TestMariaDB_ScanItems(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		m, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := m.scanItems(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		m, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := m.scanItems(mockRows)
		assert.NoError(t, err)
	})
}

func TestMariaDB_buildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = ? AND items.id = ? )"
		expectedArgs := []interface{}{
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildItemExistsQuery(exampleItem.ID, exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_ItemExists(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = ? AND items.id = ? )"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := m.ItemExists(ctx, exampleItem.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.True(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).
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
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"
		expectedArgs := []interface{}{
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildGetItemQuery(exampleItem.ID, exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetItem(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).
			WillReturnRows(buildMockRowsFromItem(exampleItem))

		actual, err := m.GetItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		actualQuery := m.buildGetAllItemsCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
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

func TestMariaDB_buildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user, COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.created_on > ? AND items.created_on < ? AND items.updated_on > ? AND items.updated_on < ? GROUP BY items.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := m.buildGetItemsQuery(exampleUser.ID, filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetItems(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	expectedListQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user, COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? GROUP BY items.id LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		exampleItemList := fakemodels.BuildFakeItemList()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnRows(
				buildMockRowsFromItem(
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
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItems(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItems(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := m.GetItems(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES (?,?,?)"
		expectedArgs := []interface{}{
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToUser,
		}
		actualQuery, actualArgs := m.buildCreateItemQuery(exampleItem)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_CreateItem(T *testing.T) {
	T.Parallel()

	expectedCreationQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES (?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
			).WillReturnResult(sqlmock.NewResult(int64(exampleItem.ID), 1))

		mtt := &mockTimeTeller{}
		mtt.On("Now").Return(exampleItem.CreatedOn)
		m.timeTeller = mtt

		actual, err := m.CreateItem(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, mtt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
			).WillReturnError(errors.New("blah"))

		actual, err := m.CreateItem(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildUpdateItemQuery(exampleItem)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_UpdateItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		exampleRows := sqlmock.NewResult(int64(exampleItem.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnResult(exampleRows)

		err := m.UpdateItem(ctx, exampleItem)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnError(errors.New("blah"))

		err := m.UpdateItem(ctx, exampleItem)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE items SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleItem.ID,
		}
		actualQuery, actualArgs := m.buildArchiveItemQuery(exampleItem.ID, exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_ArchiveItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.ID,
				exampleItem.ID,
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("returns sql.ErrNoRows with no rows affected", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.ID,
				exampleItem.ID,
			).WillReturnResult(sqlmock.NewResult(0, 0))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.ID,
				exampleItem.ID,
			).WillReturnError(errors.New("blah"))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleUser.ID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

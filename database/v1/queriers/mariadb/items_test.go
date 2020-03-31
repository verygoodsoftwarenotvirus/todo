package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
)

func buildMockRowFromItem(x *models.Item) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(itemsTableColumns).AddRow(
		x.ID,
		x.Name,
		x.Details,
		x.CreatedOn,
		x.UpdatedOn,
		x.ArchivedOn,
		x.BelongsToUser,
	)

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

func buildFakeItem() *models.Item {
	return &models.Item{
		ID:            fake.Uint64(),
		Name:          fake.Word(),
		Details:       fake.Word(),
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		BelongsToUser: fake.Uint64(),
	}
}

func buildFakeItemCreationInput(item *models.Item) *models.ItemCreationInput {
	return &models.ItemCreationInput{
		Name:          item.Name,
		Details:       item.Details,
		BelongsToUser: item.BelongsToUser,
	}
}

func TestMariaDB_buildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleItemID := fake.Uint64()
		exampleUserID := fake.Uint64()

		expectedArgCount := 2
		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = ? AND items.id = ? )"
		actualQuery, args := m.buildItemExistsQuery(exampleItemID, exampleUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
		assert.Equal(t, exampleItemID, args[1])
	})
}

func TestMariaDB_ItemExists(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = ? AND items.id = ? )"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, exampleItemID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := m.ItemExists(ctx, exampleItemID, expectedUserID)
		assert.NoError(t, err)
		assert.True(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleItemID := fake.Uint64()
		exampleUserID := fake.Uint64()

		expectedArgCount := 2
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"
		actualQuery, args := m.buildGetItemQuery(exampleItemID, exampleUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
		assert.Equal(t, exampleItemID, args[1])
	})
}

func TestMariaDB_GetItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleItem.BelongsToUser, exampleItem.ID).
			WillReturnRows(buildMockRowFromItem(exampleItem))

		actual, err := m.GetItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleItem.BelongsToUser, exampleItem.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetItemCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		filter := models.DefaultQueryFilter()
		exampleUserID := fake.Uint64()

		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"

		actualQuery, args := m.buildGetItemCountQuery(exampleUserID, filter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
	})
}

func TestMariaDB_GetItemCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()
		expectedCount := fake.Uint64()
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := m.GetItemCount(ctx, expectedUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"

		actualQuery := m.buildGetAllItemsCountQuery()
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		expectedCount := fake.Uint64()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := m.GetAllItemsCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUserID := fake.Uint64()
		filter := models.DefaultQueryFilter()

		expectedArgCount := 1
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"
		actualQuery, args := m.buildGetItemsQuery(exampleUserID, filter)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
	})
}

func TestMariaDB_GetItems(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"
	expectedCountQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleItem := buildFakeItem()
		expectedCount := fake.Uint64()
		expected := &models.ItemList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Items: []models.Item{
				*exampleItem,
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleItem.BelongsToUser).
			WillReturnRows(buildMockRowFromItem(exampleItem))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := m.GetItems(ctx, exampleItem.BelongsToUser, filter)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItems(ctx, expectedUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItems(ctx, expectedUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleItem.BelongsToUser).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := m.GetItems(ctx, exampleItem.BelongsToUser, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying for count", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleItem.BelongsToUser).
			WillReturnRows(buildMockRowFromItem(exampleItem))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItems(ctx, exampleItem.BelongsToUser, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_GetAllItemsForUser(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleItem.BelongsToUser).
			WillReturnRows(buildMockRowFromItem(exampleItem))

		expected := []models.Item{*exampleItem}
		actual, err := m.GetAllItemsForUser(ctx, exampleItem.BelongsToUser)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleItem.BelongsToUser).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := m.GetAllItemsForUser(ctx, exampleItem.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleItem := buildFakeItem()
		expectedArgCount := 3
		expectedQuery := "INSERT INTO items (name,details,belongs_to_user,created_on) VALUES (?,?,?,UNIX_TIMESTAMP())"
		actualQuery, args := m.buildCreateItemQuery(exampleItem)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleItem.Name, args[0])
		assert.Equal(t, exampleItem.Details, args[1])
		assert.Equal(t, exampleItem.BelongsToUser, args[2])
	})
}

func TestMariaDB_CreateItem(T *testing.T) {
	T.Parallel()

	expectedCreationQuery := "INSERT INTO items (name,details,belongs_to_user,created_on) VALUES (?,?,?,UNIX_TIMESTAMP())"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()
		expectedInput := buildFakeItemCreationInput(exampleItem)

		m, mockDB := buildTestService(t)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
			).WillReturnResult(sqlmock.NewResult(int64(exampleItem.ID), 1))

		expectedTimeQuery := "SELECT items.created_on FROM items WHERE items.id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WithArgs(exampleItem.ID).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(exampleItem.CreatedOn))

		actual, err := m.CreateItem(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()
		expectedInput := buildFakeItemCreationInput(exampleItem)

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
			).WillReturnError(errors.New("blah"))

		actual, err := m.CreateItem(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleItem := buildFakeItem()
		expectedArgCount := 4
		expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to_user = ? AND id = ?"
		actualQuery, args := m.buildUpdateItemQuery(exampleItem)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleItem.Name, args[0])
		assert.Equal(t, exampleItem.Details, args[1])
		assert.Equal(t, exampleItem.BelongsToUser, args[2])
		assert.Equal(t, exampleItem.ID, args[3])
	})
}

func TestMariaDB_UpdateItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()
		exampleRows := sqlmock.NewResult(int64(exampleItem.ID), 1)

		m, mockDB := buildTestService(t)
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
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
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
		exampleItem := buildFakeItem()
		expectedArgCount := 2
		expectedQuery := "UPDATE items SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		actualQuery, args := m.buildArchiveItemQuery(exampleItem.ID, exampleItem.BelongsToUser)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleItem.BelongsToUser, args[0])
		assert.Equal(t, exampleItem.ID, args[1])
	})
}

func TestMariaDB_ArchiveItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnError(errors.New("blah"))

		err := m.ArchiveItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

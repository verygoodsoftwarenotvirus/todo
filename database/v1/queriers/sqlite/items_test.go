package sqlite

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

func TestSqlite_buildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleItem := buildFakeItem()

		expectedArgCount := 2
		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = ? AND items.id = ? )"
		actualQuery, args := s.buildItemExistsQuery(exampleItem.ID, exampleItem.BelongsToUser)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleItem.BelongsToUser, args[0])
		assert.Equal(t, exampleItem.ID, args[1])
	})
}

func TestSqlite_ItemExists(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = ? AND items.id = ? )"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleItem.BelongsToUser, exampleItem.ID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := s.ItemExists(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)
		assert.True(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleItem := buildFakeItem()

		expectedArgCount := 2
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"
		actualQuery, args := s.buildGetItemQuery(exampleItem.ID, exampleItem.BelongsToUser)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleItem.BelongsToUser, args[0])
		assert.Equal(t, exampleItem.ID, args[1])
	})
}

func TestSqlite_GetItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleItem.BelongsToUser, exampleItem.ID).
			WillReturnRows(buildMockRowFromItem(exampleItem))

		actual, err := s.GetItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleItem.BelongsToUser, exampleItem.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetItemCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleUserID := fake.Uint64()
		filter := models.DefaultQueryFilter()

		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"

		actualQuery, args := s.buildGetItemCountQuery(exampleUserID, filter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
	})
}

func TestSqlite_GetItemCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"
		expectedUserID := fake.Uint64()
		expectedCount := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := s.GetItemCount(ctx, expectedUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"

		actualQuery := s.buildGetAllItemsCountQuery()
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		expectedCount := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := s.GetAllItemsCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		filter := models.DefaultQueryFilter()
		exampleUserID := fake.Uint64()

		expectedArgCount := 1
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"
		actualQuery, args := s.buildGetItemsQuery(exampleUserID, filter)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
	})
}

func TestSqlite_GetItems(T *testing.T) {
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

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleItem.BelongsToUser).
			WillReturnRows(buildMockRowFromItem(exampleItem))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := s.GetItems(ctx, exampleItem.BelongsToUser, filter)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetItems(ctx, expectedUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetItems(ctx, expectedUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleItem.BelongsToUser).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := s.GetItems(ctx, exampleItem.BelongsToUser, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying for count", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(exampleItem))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetItems(ctx, expectedUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_GetAllItemsForUser(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := fake.Uint64()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(exampleItem))

		expected := []models.Item{*exampleItem}
		actual, err := s.GetAllItemsForUser(ctx, expectedUserID)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := fake.Uint64()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := s.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleItem := buildFakeItem()
		expectedArgCount := 3
		expectedQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES (?,?,?)"
		actualQuery, args := s.buildCreateItemQuery(exampleItem)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleItem.Name, args[0])
		assert.Equal(t, exampleItem.Details, args[1])
		assert.Equal(t, exampleItem.BelongsToUser, args[2])
	})
}

func TestSqlite_CreateItem(T *testing.T) {
	T.Parallel()

	expectedCreationQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES (?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()
		expectedInput := buildFakeItemCreationInput(exampleItem)

		s, mockDB := buildTestService(t)

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

		actual, err := s.CreateItem(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()
		expectedInput := buildFakeItemCreationInput(exampleItem)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
			).WillReturnError(errors.New("blah"))

		actual, err := s.CreateItem(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleItem := buildFakeItem()
		expectedArgCount := 4
		expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"
		actualQuery, args := s.buildUpdateItemQuery(exampleItem)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleItem.Name, args[0])
		assert.Equal(t, exampleItem.Details, args[1])
		assert.Equal(t, exampleItem.BelongsToUser, args[2])
		assert.Equal(t, exampleItem.ID, args[3])
	})
}

func TestSqlite_UpdateItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()
		exampleRows := sqlmock.NewResult(int64(exampleItem.ID), 1)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnResult(exampleRows)

		err := s.UpdateItem(ctx, exampleItem)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.Name,
				exampleItem.Details,
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnError(errors.New("blah"))

		err := s.UpdateItem(ctx, exampleItem)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleItem := buildFakeItem()
		expectedArgCount := 2
		expectedQuery := "UPDATE items SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

		actualQuery, args := s.buildArchiveItemQuery(exampleItem.ID, exampleItem.BelongsToUser)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleItem.BelongsToUser, args[0])
		assert.Equal(t, exampleItem.ID, args[1])
	})
}

func TestSqlite_ArchiveItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		exampleItem := buildFakeItem()

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleItem.BelongsToUser,
				exampleItem.ID,
			).WillReturnError(errors.New("blah"))

		err := s.ArchiveItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

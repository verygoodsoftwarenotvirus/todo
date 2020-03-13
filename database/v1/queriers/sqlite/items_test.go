package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestSqlite_buildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleItemID := uint64(123)
		exampleUserID := uint64(321)

		expectedArgCount := 2
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"
		actualQuery, args := s.buildGetItemQuery(exampleItemID, exampleUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
		assert.Equal(t, exampleItemID, args[1].(uint64))
	})
}

func TestSqlite_GetItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = ? AND items.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Item{
			ID: 123,
		}
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnRows(buildMockRowFromItem(expected))

		actual, err := s.GetItem(ctx, expected.ID, expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Item{
			ID: 123,
		}
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetItem(ctx, expected.ID, expectedUserID)
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
		exampleUserID := uint64(321)
		filter := models.DefaultQueryFilter()

		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"

		actualQuery, args := s.buildGetItemCountQuery(exampleUserID, filter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestSqlite_GetItemCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"
		expectedCount := uint64(666)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := s.GetItemCount(ctx, expectedUserID, models.DefaultQueryFilter())
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
		expectedCount := uint64(666)

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
		exampleUserID := uint64(321)

		expectedArgCount := 1
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"
		actualQuery, args := s.buildGetItemsQuery(models.DefaultQueryFilter(), exampleUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestSqlite_GetItems(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(123)
		expectedCountQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		expectedItem := &models.Item{
			ID: 321,
		}
		expectedCount := uint64(666)
		expected := &models.ItemList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Items: []models.Item{
				*expectedItem,
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expectedItem))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := s.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(123)
		expected := &models.Item{
			ID: 321,
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildErroneousMockRowFromItem(expected))

		actual, err := s.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying for count", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(123)
		expected := &models.Item{
			ID: 321,
		}
		expectedCountQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expected))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
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
		expectedUserID := uint64(123)
		expectedItem := &models.Item{
			ID: 321,
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expectedItem))

		expected := []models.Item{*expectedItem}
		actual, err := s.GetAllItemsForUser(ctx, expectedUserID)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(123)

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
		expectedUserID := uint64(123)

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
		expectedUserID := uint64(123)
		exampleItem := &models.Item{
			ID: 321,
		}

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
		expected := &models.Item{
			ID:            321,
			BelongsToUser: 123,
		}
		expectedArgCount := 3
		expectedQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES (?,?,?)"
		actualQuery, args := s.buildCreateItemQuery(expected)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.Name, args[0].(string))
		assert.Equal(t, expected.Details, args[1].(string))
		assert.Equal(t, expected.BelongsToUser, args[2].(uint64))
	})
}

func TestSqlite_CreateItem(T *testing.T) {
	T.Parallel()

	expectedCreationQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES (?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:            123,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		expectedInput := &models.ItemCreationInput{
			Name:          expected.Name,
			Details:       expected.Details,
			BelongsToUser: expected.BelongsToUser,
		}

		s, mockDB := buildTestService(t)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
			).WillReturnResult(sqlmock.NewResult(int64(expected.ID), 1))

		expectedTimeQuery := "SELECT items.created_on FROM items WHERE items.id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WithArgs(expected.ID).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(expected.CreatedOn))

		actual, err := s.CreateItem(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:            123,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		expectedInput := &models.ItemCreationInput{
			Name:          expected.Name,
			Details:       expected.Details,
			BelongsToUser: expected.BelongsToUser,
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
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
		expected := &models.Item{
			ID:            321,
			BelongsToUser: 123,
		}
		expectedArgCount := 4
		expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"
		actualQuery, args := s.buildUpdateItemQuery(expected)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.Name, args[0].(string))
		assert.Equal(t, expected.Details, args[1].(string))
		assert.Equal(t, expected.BelongsToUser, args[2].(uint64))
		assert.Equal(t, expected.ID, args[3].(uint64))
	})
}

func TestSqlite_UpdateItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:            123,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
				expected.ID,
			).WillReturnResult(exampleRows)

		err := s.UpdateItem(ctx, expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:            123,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
				expected.ID,
			).WillReturnError(errors.New("blah"))

		err := s.UpdateItem(ctx, expected)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expected := &models.Item{
			ID:            321,
			BelongsToUser: 123,
		}
		expectedArgCount := 2
		expectedQuery := "UPDATE items SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		actualQuery, args := s.buildArchiveItemQuery(expected.ID, expected.BelongsToUser)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.BelongsToUser, args[0].(uint64))
		assert.Equal(t, expected.ID, args[1].(uint64))
	})
}

func TestSqlite_ArchiveItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:            123,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.BelongsToUser,
				expected.ID,
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveItem(ctx, expected.ID, expectedUserID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		example := &models.Item{
			ID:            123,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				example.BelongsToUser,
				example.ID,
			).WillReturnError(errors.New("blah"))

		err := s.ArchiveItem(ctx, example.ID, expectedUserID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

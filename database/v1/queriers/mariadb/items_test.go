package mariadb

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
		x.BelongsTo,
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
		x.BelongsTo,
		x.ID,
	)

	return exampleRows
}

func TestMariaDB_buildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleItemID := uint64(123)
		exampleUserID := uint64(321)

		expectedArgCount := 2
		expectedQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE belongs_to = ? AND id = ?"
		actualQuery, args := m.buildGetItemQuery(exampleItemID, exampleUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
		assert.Equal(t, exampleItemID, args[1].(uint64))
	})
}

func TestMariaDB_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE belongs_to = ? AND id = ?"
		expected := &models.Item{
			ID: 123,
		}
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnRows(buildMockRowFromItem(expected))

		actual, err := m.GetItem(context.Background(), expected.ID, expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE belongs_to = ? AND id = ?"
		expected := &models.Item{
			ID: 123,
		}
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItem(context.Background(), expected.ID, expectedUserID)
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
		exampleUserID := uint64(321)

		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(id) FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		actualQuery, args := m.buildGetItemCountQuery(models.DefaultQueryFilter(), exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestMariaDB_GetItemCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expectedQuery := "SELECT COUNT(id) FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"
		expectedCount := uint64(666)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := m.GetItemCount(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expectedQuery := "SELECT COUNT(id) FROM items WHERE archived_on IS NULL"

		actualQuery := m.buildGetAllItemsCountQuery()
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM items WHERE archived_on IS NULL"
		expectedCount := uint64(666)

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
		exampleUserID := uint64(321)

		expectedArgCount := 1
		expectedQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"
		actualQuery, args := m.buildGetItemsQuery(models.DefaultQueryFilter(), exampleUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestMariaDB_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(123)
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"
		expectedCountQuery := "SELECT COUNT(id) FROM items WHERE archived_on IS NULL"
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

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expectedItem))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := m.GetItems(context.Background(), models.DefaultQueryFilter(), expectedUserID)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedUserID := uint64(123)
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetItems(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		expectedUserID := uint64(123)
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItems(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		expectedUserID := uint64(123)
		expected := &models.Item{
			ID: 321,
		}
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildErroneousMockRowFromItem(expected))

		actual, err := m.GetItems(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying for count", func(t *testing.T) {
		expectedUserID := uint64(123)
		expected := &models.Item{
			ID: 321,
		}
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"
		expectedCountQuery := "SELECT COUNT(id) FROM items WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expected))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetItems(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_GetAllItemsForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(123)
		expectedItem := &models.Item{
			ID: 321,
		}
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expectedItem))

		expected := []models.Item{*expectedItem}
		actual, err := m.GetAllItemsForUser(context.Background(), expectedUserID)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedUserID := uint64(123)
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetAllItemsForUser(context.Background(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedUserID := uint64(123)
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetAllItemsForUser(context.Background(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response", func(t *testing.T) {
		expectedUserID := uint64(123)
		exampleItem := &models.Item{
			ID: 321,
		}
		expectedListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := m.GetAllItemsForUser(context.Background(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expected := &models.Item{
			ID:        321,
			BelongsTo: 123,
		}
		expectedArgCount := 3
		expectedQuery := "INSERT INTO items (name,details,belongs_to,created_on) VALUES (?,?,?,UNIX_TIMESTAMP())"
		actualQuery, args := m.buildCreateItemQuery(expected)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.Name, args[0].(string))
		assert.Equal(t, expected.Details, args[1].(string))
		assert.Equal(t, expected.BelongsTo, args[2].(uint64))
	})
}

func TestMariaDB_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:        123,
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.ItemCreationInput{
			Name:      expected.Name,
			Details:   expected.Details,
			BelongsTo: expected.BelongsTo,
		}

		m, mockDB := buildTestService(t)

		expectedCreationQuery := "INSERT INTO items (name,details,belongs_to,created_on) VALUES (?,?,?,UNIX_TIMESTAMP())"
		mockDB.ExpectExec(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsTo,
			).WillReturnResult(sqlmock.NewResult(int64(expected.ID), 1))

		expectedTimeQuery := "SELECT created_on FROM items WHERE id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WithArgs(expected.ID).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(expected.CreatedOn))

		actual, err := m.CreateItem(context.Background(), expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:        123,
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.ItemCreationInput{
			Name:      expected.Name,
			Details:   expected.Details,
			BelongsTo: expected.BelongsTo,
		}
		expectedQuery := "INSERT INTO items (name,details,belongs_to,created_on) VALUES (?,?,?,UNIX_TIMESTAMP())"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsTo,
			).WillReturnError(errors.New("blah"))

		actual, err := m.CreateItem(context.Background(), expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expected := &models.Item{
			ID:        321,
			BelongsTo: 123,
		}
		expectedArgCount := 4
		expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to = ? AND id = ?"
		actualQuery, args := m.buildUpdateItemQuery(expected)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.Name, args[0].(string))
		assert.Equal(t, expected.Details, args[1].(string))
		assert.Equal(t, expected.BelongsTo, args[2].(uint64))
		assert.Equal(t, expected.ID, args[3].(uint64))
	})
}

func TestMariaDB_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:        123,
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to = ? AND id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsTo,
				expected.ID,
			).WillReturnResult(exampleRows)

		err := m.UpdateItem(context.Background(), expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:        123,
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE items SET name = ?, details = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to = ? AND id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsTo,
				expected.ID,
			).WillReturnError(errors.New("blah"))

		err := m.UpdateItem(context.Background(), expected)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expected := &models.Item{
			ID:        321,
			BelongsTo: 123,
		}
		expectedArgCount := 2
		expectedQuery := "UPDATE items SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"
		actualQuery, args := m.buildArchiveItemQuery(expected.ID, expected.BelongsTo)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.BelongsTo, args[0].(uint64))
		assert.Equal(t, expected.ID, args[1].(uint64))
	})
}

func TestMariaDB_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Item{
			ID:        123,
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE items SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.BelongsTo,
				expected.ID,
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveItem(context.Background(), expected.ID, expectedUserID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		expectedUserID := uint64(321)
		example := &models.Item{
			ID:        123,
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE items SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				example.BelongsTo,
				example.ID,
			).WillReturnError(errors.New("blah"))

		err := m.ArchiveItem(context.Background(), example.ID, expectedUserID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

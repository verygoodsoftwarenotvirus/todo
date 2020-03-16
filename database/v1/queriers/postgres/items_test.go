package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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

func TestPostgres_buildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expectedArgCount := 2
		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = $1 AND items.id = $2 )"
		actualQuery, args := p.buildItemExistsQuery(exampleItemID, expectedUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
		assert.Equal(t, exampleItemID, args[1].(uint64))
	})
}

func TestPostgres_ItemExists(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.belongs_to_user = $1 AND items.id = $2 )"

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := true
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expectedItemID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(expected))

		actual, err := p.ItemExists(ctx, expectedItemID, expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expectedArgCount := 2
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = $1 AND items.id = $2"
		actualQuery, args := p.buildGetItemQuery(exampleItemID, expectedUserID)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
		assert.Equal(t, exampleItemID, args[1].(uint64))
	})
}

func TestPostgres_GetItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.belongs_to_user = $1 AND items.id = $2"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Item{
			ID: fake.Uint64(),
		}
		expectedUserID := fake.Uint64()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnRows(buildMockRowFromItem(expected))

		actual, err := p.GetItem(ctx, expected.ID, expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Item{
			ID: fake.Uint64(),
		}
		expectedUserID := fake.Uint64()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetItem(ctx, expected.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetItemCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedUserID := fake.Uint64()
		filter := models.DefaultQueryFilter()

		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = $1 LIMIT 20"

		actualQuery, args := p.buildGetItemCountQuery(expectedUserID, filter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestPostgres_GetItemCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = $1 LIMIT 20"
		expectedCount := uint64(666)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := p.GetItemCount(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"

		actualQuery := p.buildGetAllItemsCountQuery()
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		expectedCount := uint64(666)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := p.GetAllItemsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedUserID := fake.Uint64()
		filter := models.DefaultQueryFilter()

		expectedArgCount := 1
		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = $1 LIMIT 20"
		actualQuery, args := p.buildGetItemsQuery(expectedUserID, filter)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestPostgres_GetItems(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = $1 LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expectedCountQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		expectedItem := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
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

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expectedItem))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := p.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
		}

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildErroneousMockRowFromItem(expected))

		actual, err := p.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying for count", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
		}
		expectedCountQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expected))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetItems(ctx, expectedUserID, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_GetAllItemsForUser(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT items.id, items.name, items.details, items.created_on, items.updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = $1"

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expectedItem := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
		}

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildMockRowFromItem(expectedItem))

		expected := []models.Item{*expectedItem}
		actual, err := p.GetAllItemsForUser(ctx, expectedUserID)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedUserID := fake.Uint64()

		exampleItem := &models.Item{
			ID: fake.Uint64(),
		}

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(buildErroneousMockRowFromItem(exampleItem))

		actual, err := p.GetAllItemsForUser(ctx, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
		}

		expectedArgCount := 3
		expectedQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES ($1,$2,$3) RETURNING id, created_on"
		actualQuery, args := p.buildCreateItemQuery(expected)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.Name, args[0].(string))
		assert.Equal(t, expected.Details, args[1].(string))
		assert.Equal(t, expected.BelongsToUser, args[2].(uint64))
	})
}

func TestPostgres_CreateItem(T *testing.T) {
	T.Parallel()

	expectedCreationQuery := "INSERT INTO items (name,details,belongs_to_user) VALUES ($1,$2,$3) RETURNING id, created_on"

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		expectedInput := &models.ItemCreationInput{
			Name:          expected.Name,
			Details:       expected.Details,
			BelongsToUser: expected.BelongsToUser,
		}
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(expected.ID, expected.CreatedOn)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
			).WillReturnRows(exampleRows)

		actual, err := p.CreateItem(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		expectedInput := &models.ItemCreationInput{
			Name:          expected.Name,
			Details:       expected.Details,
			BelongsToUser: expected.BelongsToUser,
		}

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCreationQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
			).WillReturnError(errors.New("blah"))

		actual, err := p.CreateItem(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
		}

		expectedArgCount := 4
		expectedQuery := "UPDATE items SET name = $1, details = $2, updated_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $3 AND id = $4 RETURNING updated_on"
		actualQuery, args := p.buildUpdateItemQuery(expected)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.Name, args[0].(string))
		assert.Equal(t, expected.Details, args[1].(string))
		assert.Equal(t, expected.BelongsToUser, args[2].(uint64))
		assert.Equal(t, expected.ID, args[3].(uint64))
	})
}

func TestPostgres_UpdateItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET name = $1, details = $2, updated_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $3 AND id = $4 RETURNING updated_on"

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		exampleRows := sqlmock.NewRows([]string{"updated_on"}).AddRow(uint64(time.Now().Unix()))

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
				expected.ID,
			).WillReturnRows(exampleRows)

		err := p.UpdateItem(ctx, expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.Name,
				expected.Details,
				expected.BelongsToUser,
				expected.ID,
			).WillReturnError(errors.New("blah"))

		err := p.UpdateItem(ctx, expected)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
		}

		expectedArgCount := 2
		expectedQuery := "UPDATE items SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1 AND id = $2 RETURNING archived_on"
		actualQuery, args := p.buildArchiveItemQuery(expected.ID, expected.BelongsToUser)

		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.BelongsToUser, args[0].(uint64))
		assert.Equal(t, expected.ID, args[1].(uint64))
	})
}

func TestPostgres_ArchiveItem(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE items SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1 AND id = $2 RETURNING archived_on"

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		expected := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				expected.BelongsToUser,
				expected.ID,
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.ArchiveItem(ctx, expected.ID, expectedUserID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		p, mockDB := buildTestService(t)
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expectedUserID := fake.Uint64()

		example := &models.Item{
			ID:            expectedItemID,
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				example.BelongsToUser,
				example.ID,
			).WillReturnError(errors.New("blah"))

		err := p.ArchiveItem(ctx, example.ID, expectedUserID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

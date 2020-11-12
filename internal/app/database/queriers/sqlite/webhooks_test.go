package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromWebhook(webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := webhooksTableColumns
	exampleRows := sqlmock.NewRows(columns)

	for _, w := range webhooks {
		rowValues := []driver.Value{
			w.ID,
			w.Name,
			w.ContentType,
			w.URL,
			w.Method,
			strings.Join(w.Events, eventsSeparator),
			strings.Join(w.DataTypes, typesSeparator),
			strings.Join(w.Topics, topicsSeparator),
			w.CreatedOn,
			w.LastUpdatedOn,
			w.ArchivedOn,
			w.BelongsToUser,
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromWebhook(w *types.Webhook) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(webhooksTableColumns).AddRow(
		w.ArchivedOn,
		w.BelongsToUser,
		w.Name,
		w.ContentType,
		w.URL,
		w.Method,
		strings.Join(w.Events, eventsSeparator),
		strings.Join(w.DataTypes, typesSeparator),
		strings.Join(w.Topics, topicsSeparator),
		w.CreatedOn,
		w.LastUpdatedOn,
		w.ID,
	)

	return exampleRows
}

func TestSqlite_ScanWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, err := s.scanWebhooks(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, err := s.scanWebhooks(mockRows)
		assert.NoError(t, err)
	})
}

func TestSqlite_buildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = ? AND webhooks.id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := s.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromWebhook(exampleWebhook))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"
		actualQuery := s.buildGetAllWebhooksCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllWebhooksCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := s.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllWebhooksCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooksCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL"
		actualQuery := s.buildGetAllWebhooksQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()
		exampleWebhookList.Limit = 0

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllWebhooksQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(
				buildMockRowsFromWebhook(
					&exampleWebhookList.Webhooks[0],
					&exampleWebhookList.Webhooks[1],
					&exampleWebhookList.Webhooks[2],
				),
			)

		actual, err := s.GetAllWebhooks(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllWebhooksQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllWebhooksQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllWebhooksQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? AND webhooks.created_on > ? AND webhooks.created_on < ? AND webhooks.last_updated_on > ? AND webhooks.last_updated_on < ? ORDER BY webhooks.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := s.buildGetWebhooksQuery(exampleUser.ID, filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromWebhook(
					&exampleWebhookList.Webhooks[0],
					&exampleWebhookList.Webhooks[1],
					&exampleWebhookList.Webhooks[2],
				),
			)

		actual, err := s.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES (?,?,?,?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		}
		actualQuery, actualArgs := s.buildCreateWebhookQuery(exampleWebhook)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildCreateWebhookQuery(exampleWebhook)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		stt := &mockTimeTeller{}
		stt.On("Now").Return(exampleWebhook.CreatedOn)
		s.timeTeller = stt

		actual, err := s.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, stt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error interacting with database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildCreateWebhookQuery(exampleWebhook)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.CreateWebhook(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, last_updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := s.buildUpdateWebhookQuery(exampleWebhook)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildUpdateWebhookQuery(exampleWebhook)

		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		err := s.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildUpdateWebhookQuery(exampleWebhook)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := s.UpdateWebhook(ctx, exampleWebhook)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET last_updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := s.buildArchiveWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildArchiveWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

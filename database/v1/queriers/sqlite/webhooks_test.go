package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromWebhook(webhooks ...*models.Webhook) *sqlmock.Rows {
	includeCount := len(webhooks) > 1
	columns := webhooksTableColumns

	if includeCount {
		columns = append(columns, "count")
	}
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
			w.UpdatedOn,
			w.ArchivedOn,
			w.BelongsToUser,
		}

		if includeCount {
			rowValues = append(rowValues, len(webhooks))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromWebhook(w *models.Webhook) *sqlmock.Rows {
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
		w.UpdatedOn,
		w.ID,
	)

	return exampleRows
}

func TestSqlite_buildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = ? AND webhooks.id = ?"
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

	expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = ? AND webhooks.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnRows(buildMockRowsFromWebhook(exampleWebhook))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
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
		s, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"
		actualQuery := s.buildGetAllWebhooksCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleCount := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := s.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
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
		s, _ := buildTestService(t)

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL"
		actualQuery := s.buildGetAllWebhooksQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()
		exampleWebhookList.Limit = 0

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
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
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
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
		s, _ := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()
		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user, COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? AND webhooks.created_on > ? AND webhooks.created_on < ? AND webhooks.updated_on > ? AND webhooks.updated_on < ? GROUP BY webhooks.id LIMIT 20 OFFSET 180"
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
	expectedListQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user, COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? GROUP BY webhooks.id LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleUser.ID).
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
		ctx := context.Background()
		filter := models.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildWebhookCreationQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
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
		actualQuery, actualArgs := s.buildWebhookCreationQuery(exampleWebhook)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_CreateWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES (?,?,?,?,?,?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		).WillReturnResult(exampleRows)

		mtt := &mockTimeTeller{}
		mtt.On("Now").Return(exampleWebhook.CreatedOn)
		s.timeTeller = mtt

		actual, err := s.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error interacting with database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		).WillReturnError(errors.New("blah"))

		actual, err := s.CreateWebhook(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"
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

	expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnResult(exampleRows)

		err := s.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnError(errors.New("blah"))

		err := s.UpdateWebhook(ctx, exampleWebhook)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
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
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()
		expectedQuery := "UPDATE webhooks SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

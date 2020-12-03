package mariadb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
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

func buildMockRowsFromWebhooks(includeCount bool, webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := queriers.WebhooksTableColumns

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
			strings.Join(w.Events, queriers.WebhooksTableEventsSeparator),
			strings.Join(w.DataTypes, queriers.WebhooksTableDataTypesSeparator),
			strings.Join(w.Topics, queriers.WebhooksTableTopicsSeparator),
			w.CreatedOn,
			w.LastUpdatedOn,
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

func buildErroneousMockRowFromWebhook(w *types.Webhook) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.WebhooksTableColumns).AddRow(
		w.ArchivedOn,
		w.BelongsToUser,
		w.Name,
		w.ContentType,
		w.URL,
		w.Method,
		strings.Join(w.Events, queriers.WebhooksTableEventsSeparator),
		strings.Join(w.DataTypes, queriers.WebhooksTableDataTypesSeparator),
		strings.Join(w.Topics, queriers.WebhooksTableTopicsSeparator),
		w.CreatedOn,
		w.LastUpdatedOn,
		w.ID,
	)

	return exampleRows
}

func TestMariaDB_ScanWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := q.scanWebhooks(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := q.scanWebhooks(mockRows, false)
		assert.NoError(t, err)
	})
}

func TestMariaDB_buildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = ? AND webhooks.id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(false, exampleWebhook))

		actual, err := q.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := q.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"
		actualQuery := q.buildGetAllWebhooksCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		q, mockDB := buildTestService(t)
		expectedQuery := q.buildGetAllWebhooksCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := q.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		expectedQuery := q.buildGetAllWebhooksCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := q.GetAllWebhooksCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetBatchOfWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.id > ? AND webhooks.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.buildGetBatchOfWebhooksQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	_q, _ := buildTestService(T)
	expectedCountQuery := _q.buildGetAllWebhooksCountQuery()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleWebhookList := fakes.BuildFakeWebhookList()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.buildGetBatchOfWebhooksQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromWebhooks(
					false,
					&exampleWebhookList.Webhooks[0],
					&exampleWebhookList.Webhooks[1],
					&exampleWebhookList.Webhooks[2],
				),
			)

		out := make(chan []types.Webhook)
		doneChan := make(chan bool, 1)

		err := q.GetAllWebhooks(ctx, out)
		assert.NoError(t, err)

		stillQuerying := true
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

		q, mockDB := buildTestService(t)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		out := make(chan []types.Webhook)

		err := q.GetAllWebhooks(ctx, out)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.buildGetBatchOfWebhooksQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		out := make(chan []types.Webhook)

		err := q.GetAllWebhooks(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.buildGetBatchOfWebhooksQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		out := make(chan []types.Webhook)

		err := q.GetAllWebhooks(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleWebhook := fakes.BuildFakeWebhook()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.buildGetBatchOfWebhooksQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		out := make(chan []types.Webhook)

		err := q.GetAllWebhooks(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_user, (SELECT COUNT(*) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ?) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? AND webhooks.created_on > ? AND webhooks.created_on < ? AND webhooks.last_updated_on > ? AND webhooks.last_updated_on < ? ORDER BY webhooks.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.buildGetWebhooksQuery(exampleUser.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleWebhookList := fakes.BuildFakeWebhookList()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromWebhooks(
					true,
					&exampleWebhookList.Webhooks[0],
					&exampleWebhookList.Webhooks[1],
					&exampleWebhookList.Webhooks[2],
				),
			)

		actual, err := q.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetWebhooksQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := q.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES (?,?,?,?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, queriers.WebhooksTableEventsSeparator),
			strings.Join(exampleWebhook.DataTypes, queriers.WebhooksTableDataTypesSeparator),
			strings.Join(exampleWebhook.Topics, queriers.WebhooksTableTopicsSeparator),
			exampleWebhook.BelongsToUser,
		}
		actualQuery, actualArgs := q.buildCreateWebhookQuery(exampleWebhook)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildCreateWebhookQuery(exampleWebhook)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		mtt := &queriers.MockTimeTeller{}
		mtt.On("Now").Return(exampleWebhook.CreatedOn)
		q.timeTeller = mtt

		actual, err := q.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, mtt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error interacting with database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildCreateWebhookQuery(exampleWebhook)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.CreateWebhook(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, last_updated_on = UNIX_TIMESTAMP() WHERE belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, queriers.WebhooksTableEventsSeparator),
			strings.Join(exampleWebhook.DataTypes, queriers.WebhooksTableDataTypesSeparator),
			strings.Join(exampleWebhook.Topics, queriers.WebhooksTableTopicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.buildUpdateWebhookQuery(exampleWebhook)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildUpdateWebhookQuery(exampleWebhook)

		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		err := q.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildUpdateWebhookQuery(exampleWebhook)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := q.UpdateWebhook(ctx, exampleWebhook)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET last_updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.buildArchiveWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildArchiveWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := q.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

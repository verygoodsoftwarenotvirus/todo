package superclient

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromWebhooks(includeCounts bool, filteredCount uint64, webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := queriers.WebhooksTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
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

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(webhooks))
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

func TestClient_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetWebhookQuery", exampleWebhook.ID, exampleWebhook.BelongsToUser).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(false, 0, exampleWebhook))

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, db.ExpectationsWereMet(), "not all database query expectations were met")
		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		expected := uint64(123)
		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetAllWebhooksCountQuery").
			Return(fakeQuery, []interface{}{})

		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expected))

		actual, err := c.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		results := make(chan []*types.Webhook)
		doneChan := make(chan bool, 1)
		expectedCount := uint64(20)
		exampleWebhookList := fakes.BuildFakeWebhookList()
		exampleBatchSize := uint16(1000)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetAllWebhooksCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetBatchOfWebhooksQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)

		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(false, 0, exampleWebhookList.Webhooks...))

		err := c.GetAllWebhooks(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		stillQuerying := true
		for stillQuerying {
			select {
			case batch := <-results:
				assert.NotEmpty(t, batch)
				doneChan <- true
			case <-time.After(time.Second):
				t.FailNow()
			case <-doneChan:
				stillQuerying = false
			}
		}

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()

		results := make(chan []*types.Webhook)
		exampleBatchSize := uint16(1000)

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetAllWebhooksCountQuery").
			Return(fakeQuery, []interface{}{})

		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		err := c.GetAllWebhooks(ctx, results, exampleBatchSize)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		results := make(chan []*types.Webhook)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetAllWebhooksCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetBatchOfWebhooksQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllWebhooks(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()

		results := make(chan []*types.Webhook)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetAllWebhooksCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetBatchOfWebhooksQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllWebhooks(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		results := make(chan []*types.Webhook)
		expectedCount := uint64(20)
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleBatchSize := uint16(1000)

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetAllWebhooksCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.
			On("BuildGetBatchOfWebhooksQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllWebhooks(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleWebhookList := fakes.BuildFakeWebhookList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On("BuildGetWebhooksQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(
				true,
				exampleWebhookList.FilteredCount,
				exampleWebhookList.Webhooks...,
			))

		actual, err := c.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		exampleWebhookList := fakes.BuildFakeWebhookList()
		exampleWebhookList.Page = 0
		exampleWebhookList.Limit = 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On("BuildGetWebhooksQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(
				true,
				exampleWebhookList.FilteredCount,
				exampleWebhookList.Webhooks...,
			))

		actual, err := c.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleRows := newSuccessfulDatabaseResult(exampleWebhook.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On("BuildCreateWebhookQuery", exampleInput).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(exampleRows)

		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		var expected error

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On("BuildUpdateWebhookQuery", exampleWebhook).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		exampleRows := newSuccessfulDatabaseResult(exampleWebhook.ID)
		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(exampleRows)

		actual := c.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		var expected error

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On("BuildArchiveWebhookQuery", exampleWebhook.ID, exampleWebhook.BelongsToUser).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		exampleRows := newSuccessfulDatabaseResult(exampleWebhook.ID)
		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(exampleRows)

		actual := c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAuditLogEntriesForWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		auditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On("BuildGetAuditLogEntriesForWebhookQuery", exampleWebhook.ID).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(
				false,
				auditLogEntries...,
			))

		actual, err := c.GetAuditLogEntriesForWebhook(ctx, exampleWebhook.ID)
		assert.NoError(t, err)
		assert.Equal(t, auditLogEntries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

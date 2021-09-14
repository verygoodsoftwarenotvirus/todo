package queriers

import (
	"context"
	"database/sql/driver"
	"errors"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromWebhooks(includeCounts bool, filteredCount uint64, webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := querybuilding.WebhooksTableColumns

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
			strings.Join(w.Events, querybuilding.WebhooksTableEventsSeparator),
			strings.Join(w.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
			strings.Join(w.Topics, querybuilding.WebhooksTableTopicsSeparator),
			w.CreatedOn,
			w.LastUpdatedOn,
			w.ArchivedOn,
			w.BelongsToAccount,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(webhooks))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowsFromWebhooks(includeCounts bool, filteredCount uint64, webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := querybuilding.WebhooksTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, w := range webhooks {
		rowValues := []driver.Value{
			w.ArchivedOn,
			strings.Join(w.Events, querybuilding.WebhooksTableEventsSeparator),
			strings.Join(w.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
			strings.Join(w.Topics, querybuilding.WebhooksTableTopicsSeparator),
			w.ID,
			w.Name,
			w.ContentType,
			w.URL,
			w.Method,
			w.CreatedOn,
			w.LastUpdatedOn,
			w.BelongsToAccount,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(webhooks))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanWebhooks(ctx, mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanWebhooks(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(false, 0, exampleWebhook))

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, exampleAccount.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetWebhook(ctx, "", exampleAccount.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRowsFromWebhooks(false, 0, exampleWebhook))

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, exampleAccount.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint64(123)
		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetAllWebhooksCountQuery", testutils.ContextMatcher).Return(fakeQuery, []interface{}{})

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

func TestQuerier_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleAccount := fakes.BuildFakeAccount()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleWebhookList := fakes.BuildFakeWebhookList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetWebhooksQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(
				true,
				exampleWebhookList.FilteredCount,
				exampleWebhookList.Webhooks...,
			))

		actual, err := c.GetWebhooks(ctx, exampleAccount.ID, filter)
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
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetWebhooksQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(
				true,
				exampleWebhookList.FilteredCount,
				exampleWebhookList.Webhooks...,
			))

		actual, err := c.GetWebhooks(ctx, exampleAccount.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		filter := types.DefaultQueryFilter()
		c, _ := buildTestClient(t)

		actual, err := c.GetWebhooks(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetWebhooksQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetWebhooks(ctx, exampleAccount.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous database response", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetWebhooksQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetWebhooks(ctx, exampleAccount.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildCreateWebhookQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder
		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateWebhook(ctx, nil, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder
		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing creation query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildCreateWebhookQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder
		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildCreateWebhookQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder
		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildCreateWebhookQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder
		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildUpdateWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.NoError(t, c.UpdateWebhook(ctx, exampleWebhook, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateWebhook(ctx, nil, exampleUser.ID, nil))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateWebhook(ctx, exampleWebhook, "", nil))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateWebhook(ctx, exampleWebhook, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildUpdateWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.UpdateWebhook(ctx, exampleWebhook, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildUpdateWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.UpdateWebhook(ctx, exampleWebhook, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildUpdateWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.UpdateWebhook(ctx, exampleWebhook, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildArchiveWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		actual := c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleAccount.ID, exampleUser.ID)
		assert.NoError(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveWebhook(ctx, "", exampleWebhook.BelongsToAccount, exampleUser.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhook.ID, "", exampleUser.ID))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToAccount, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildArchiveWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildArchiveWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildArchiveWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAuditLogEntriesForWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		auditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
		).Return(fakeQuery, fakeArgs)
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

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntriesForWebhook(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForWebhook(ctx, exampleWebhook.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.WebhookSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForWebhookQuery",
			testutils.ContextMatcher,
			exampleWebhook.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForWebhook(ctx, exampleWebhook.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

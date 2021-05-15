package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func prepareForAuditLogEntryCreation(t *testing.T, exampleAuditLogEntry *types.AuditLogEntryCreationInput, mockQueryBuilder *database.MockSQLQueryBuilder, db sqlmock.Sqlmock) {
	t.Helper()

	fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
	mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
		"BuildCreateAuditLogEntryQuery",
		testutil.ContextMatcher,
		exampleAuditLogEntry,
	).Return(fakeQuery, fakeArgs)

	db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
		WithArgs(interfaceToDriverValue(fakeArgs)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func buildMockRowsFromAuditLogEntries(includeCount bool, auditLogEntries ...*types.AuditLogEntry) *sqlmock.Rows {
	columns := querybuilding.AuditLogEntriesTableColumns

	if includeCount {
		columns = append(columns, "count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range auditLogEntries {
		rowValues := []driver.Value{
			x.ID,
			x.ExternalID,
			x.EventType,
			x.Context,
			x.CreatedOn,
		}

		if includeCount {
			rowValues = append(rowValues, len(auditLogEntries))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := q.scanAuditLogEntries(ctx, mockRows, false)
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

		_, _, err := q.scanAuditLogEntries(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAuditLogEntryQuery",
			testutil.ContextMatcher,
			exampleAuditLogEntry.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntry))

		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid entry ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntry(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAuditLogEntryQuery",
			testutil.ContextMatcher,
			exampleAuditLogEntry.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAllAuditLogEntriesCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAllAuditLogEntriesCountQuery", testutil.ContextMatcher).Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllAuditLogEntriesCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAllAuditLogEntriesCountQuery", testutil.ContextMatcher).Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAllAuditLogEntriesCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAllAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.AuditLogEntry)
		doneChan := make(chan bool, 1)
		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries
		exampleBatchSize := uint16(1000)
		expectedStart, expectedEnd := uint64(1), uint64(1001)
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeCountQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAllAuditLogEntriesCountQuery",
			testutil.ContextMatcher,
		).Return(fakeCountQuery)

		db.ExpectQuery(formatQueryForSQLMock(fakeCountQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(123))

		fakeSelectQuery, fakeSelectArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetBatchOfAuditLogEntriesQuery",
			testutil.ContextMatcher,
			expectedStart, expectedEnd,
		).Return(fakeSelectQuery, fakeSelectArgs)

		db.ExpectQuery(formatQueryForSQLMock(fakeSelectQuery)).
			WithArgs(interfaceToDriverValue(fakeSelectArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntries...))

		c.sqlQueryBuilder = mockQueryBuilder

		assert.NoError(t, c.GetAllAuditLogEntries(ctx, results, exampleBatchSize))

		var stillQuerying = true
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

	T.Run("with nil results channel", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleBatchSize := uint16(1000)
		c, db := buildTestClient(t)

		assert.Error(t, c.GetAllAuditLogEntries(ctx, nil, exampleBatchSize))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with now rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.AuditLogEntry)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAllAuditLogEntriesCountQuery", testutil.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetBatchOfAuditLogEntriesQuery",
			testutil.ContextMatcher,
			uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		assert.NoError(t, c.GetAllAuditLogEntries(ctx, results, exampleBatchSize))

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.AuditLogEntry)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAllAuditLogEntriesCountQuery", testutil.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllAuditLogEntries(ctx, results, exampleBatchSize)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.AuditLogEntry)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAllAuditLogEntriesCountQuery", testutil.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetBatchOfAuditLogEntriesQuery",
			testutil.ContextMatcher,
			uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.NoError(t, c.GetAllAuditLogEntries(ctx, results, exampleBatchSize))

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.AuditLogEntry)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAllAuditLogEntriesCountQuery", testutil.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetBatchOfAuditLogEntriesQuery",
			testutil.ContextMatcher,
			uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		assert.NoError(t, c.GetAllAuditLogEntries(ctx, results, exampleBatchSize))

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAuditLogEntriesQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(true, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()
		exampleAuditLogEntryList.Page = 0
		exampleAuditLogEntryList.Limit = 0
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAuditLogEntriesQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(true, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAuditLogEntriesQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildGetAuditLogEntriesQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_createAuditLogEntryInTransaction(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, c.createAuditLogEntryInTransaction(ctx, tx, exampleInput))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		assert.Error(t, c.createAuditLogEntryInTransaction(ctx, tx, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil querier", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.createAuditLogEntryInTransaction(ctx, nil, exampleInput))
	})

	T.Run("obligatory but with helper method", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		prepareForAuditLogEntryCreation(t, exampleInput, mockQueryBuilder, db)

		c.sqlQueryBuilder = mockQueryBuilder

		assert.NoError(t, c.createAuditLogEntryInTransaction(ctx, tx, exampleInput))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		assert.Error(t, c.createAuditLogEntryInTransaction(ctx, tx, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil querier", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		assert.Error(t, c.createAuditLogEntryInTransaction(ctx, nil, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.createAuditLogEntryInTransaction(ctx, tx, exampleInput))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_createAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		c.createAuditLogEntry(ctx, tx, exampleInput)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("obligatory but with helper method", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		prepareForAuditLogEntryCreation(t, exampleInput, mockQueryBuilder, db)

		c.sqlQueryBuilder = mockQueryBuilder

		c.createAuditLogEntry(ctx, tx, exampleInput)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		c.createAuditLogEntry(ctx, tx, nil)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil querier", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.createAuditLogEntry(ctx, nil, exampleInput)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		c.createAuditLogEntry(ctx, tx, exampleInput)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}
package superclient

import (
	"context"
	"database/sql/driver"
	"errors"
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

func buildMockRowsFromAuditLogEntries(includeCount bool, auditLogEntries ...*types.AuditLogEntry) *sqlmock.Rows {
	columns := queriers.AuditLogEntriesTableColumns

	if includeCount {
		columns = append(columns, "count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range auditLogEntries {
		rowValues := []driver.Value{
			x.ID,
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

func buildErroneousMockRowFromAuditLogEntry(x *types.AuditLogEntry) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.AuditLogEntriesTableColumns).AddRow(
		x.CreatedOn,
		x.ID,
		x.EventType,
		x.Context,
	)

	return exampleRows
}

func TestSqlite_ScanAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := q.scanAuditLogEntries(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := q.scanAuditLogEntries(mockRows, false)
		assert.Error(t, err)
	})
}

func TestClient_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildGetAuditLogEntryQuery", exampleAuditLogEntry.ID).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntry))

		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func prepareForAuditLogEntryCreation(t *testing.T, exampleAuditLogEntry *types.AuditLogEntry, mockQueryBuilder *database.MockSQLQueryBuilder, db sqlmock.Sqlmock) {
	t.Helper()

	fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
	mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildCreateAuditLogEntryQuery", mock.MatchedBy(matchAuditLogEntry(t, exampleAuditLogEntry))).Return(fakeQuery, fakeArgs)

	db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
		WithArgs(interfaceToDriverValue(fakeArgs)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func TestClient_createAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildCreateAuditLogEntryQuery", mock.MatchedBy(matchAuditLogEntry(t, exampleAuditLogEntry))).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		c.createAuditLogEntry(ctx, exampleInput)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("obligatory but with helper method", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.createAuditLogEntry(ctx, exampleInput)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildGetAllAuditLogEntriesCountQuery").Return(fakeCountQuery)

		db.ExpectQuery(formatQueryForSQLMock(fakeCountQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(123))

		fakeSelectQuery, fakeSelectArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildGetBatchOfAuditLogEntriesQuery", expectedStart, expectedEnd).Return(fakeSelectQuery, fakeSelectArgs)

		db.ExpectQuery(formatQueryForSQLMock(fakeSelectQuery)).
			WithArgs(interfaceToDriverValue(fakeSelectArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntries...))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllAuditLogEntries(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

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
}

func TestClient_GetAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildGetAuditLogEntriesQuery", filter).Return(fakeQuery, fakeArgs)
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
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildGetAuditLogEntriesQuery", filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(true, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var auditLogCreationInputMatcher interface{} = mock.MatchedBy(func(*types.AuditLogEntryCreationInput) bool {
	return true
})

func buildMockRowsFromAuditLogEntries(includeCount bool, auditLogEntries ...*types.AuditLogEntry) *sqlmock.Rows {
	columns := querybuilding.AuditLogEntriesTableColumns

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

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntry))

		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid entry ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntry(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAllAuditLogEntriesCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllAuditLogEntriesCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAllAuditLogEntriesCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(true, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(true, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, c.createAuditLogEntryInTransaction(ctx, tx, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		assert.NoError(t, c.createAuditLogEntryInTransaction(ctx, tx, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.createAuditLogEntryInTransaction(ctx, tx, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		c.createAuditLogEntry(ctx, tx, exampleInput)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("obligatory but with helper method", func(t *testing.T) {
		t.Parallel()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		exampleInput := fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		c.createAuditLogEntry(ctx, tx, exampleInput)

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectBegin()

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		c.createAuditLogEntry(ctx, tx, exampleInput)

		mock.AssertExpectationsForObjects(t, db)
	})
}

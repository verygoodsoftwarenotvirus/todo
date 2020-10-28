package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromAuditLogEntries(auditLogEntries ...*models.AuditLogEntry) *sqlmock.Rows {
	columns := auditLogEntriesTableColumns

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range auditLogEntries {
		rowValues := []driver.Value{
			x.ID,
			x.EventType,
			x.Context,
			x.CreatedOn,
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromAuditLogEntry(x *models.AuditLogEntry) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(auditLogEntriesTableColumns).AddRow(
		x.CreatedOn,
		x.ID,
		x.EventType,
		x.Context,
	)

	return exampleRows
}

func TestPostgres_ScanAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, err := p.scanAuditLogEntries(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, err := p.scanAuditLogEntries(mockRows)
		assert.NoError(t, err)
	})
}

func TestPostgres_buildGetAuditLogEntryQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.id = $1"
		expectedArgs := []interface{}{
			exampleAuditLogEntry.ID,
		}
		actualQuery, actualArgs := p.buildGetAuditLogEntryQuery(exampleAuditLogEntry.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetAuditLogEntryQuery(exampleAuditLogEntry.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnRows(buildMockRowsFromAuditLogEntries(exampleAuditLogEntry))

		actual, err := p.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		p, mockDB := buildTestService(t)

		expectedQuery, expectedArgs := p.buildGetAuditLogEntryQuery(exampleAuditLogEntry.ID)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllAuditLogEntriesCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(audit_log.id) FROM audit_log WHERE audit_log.archived_on IS NULL"
		actualQuery := p.buildGetAllAuditLogEntriesCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllAuditLogEntriesCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		expectedCount := uint64(123)

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(p.buildGetAllAuditLogEntriesCountQuery())).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := p.GetAllAuditLogEntriesCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetBatchOfAuditLogEntriesQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.id > $1 AND audit_log.id < $2"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := p.buildGetBatchOfAuditLogEntriesQuery(beginID, endID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetAllAuditLogEntries(T *testing.T) {
	T.Parallel()

	p, _ := buildTestService(T)
	expectedCountQuery := p.buildGetAllAuditLogEntriesCountQuery()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		p, mockDB := buildTestService(t)
		exampleAuditLogEntryList := fakemodels.BuildFakeAuditLogEntryList()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := p.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnRows(
				buildMockRowsFromAuditLogEntries(
					&exampleAuditLogEntryList.AuditLogEntries[0],
					&exampleAuditLogEntryList.AuditLogEntries[1],
					&exampleAuditLogEntryList.AuditLogEntries[2],
				),
			)

		out := make(chan []models.AuditLogEntry)
		doneChan := make(chan bool, 1)

		err := p.GetAllAuditLogEntries(ctx, out)
		assert.NoError(t, err)

		var stillQuerying = true
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

		p, mockDB := buildTestService(t)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		out := make(chan []models.AuditLogEntry)

		err := p.GetAllAuditLogEntries(ctx, out)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		p, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := p.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnError(sql.ErrNoRows)

		out := make(chan []models.AuditLogEntry)

		err := p.GetAllAuditLogEntries(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		p, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := p.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnError(errors.New("blah"))

		out := make(chan []models.AuditLogEntry)

		err := p.GetAllAuditLogEntries(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		p, mockDB := buildTestService(t)
		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := p.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnRows(buildErroneousMockRowFromAuditLogEntry(exampleAuditLogEntry))

		out := make(chan []models.AuditLogEntry)

		err := p.GetAllAuditLogEntries(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAuditLogEntriesQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.archived_on IS NULL AND audit_log.created_on > $1 AND audit_log.created_on < $2 AND audit_log.last_updated_on > $3 AND audit_log.last_updated_on < $4 ORDER BY audit_log.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := p.buildGetAuditLogEntriesQuery(filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		exampleAuditLogEntryList := fakemodels.BuildFakeAuditLogEntryList()
		expectedQuery, expectedArgs := p.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnRows(
				buildMockRowsFromAuditLogEntries(
					&exampleAuditLogEntryList.AuditLogEntries[0],
					&exampleAuditLogEntryList.AuditLogEntries[1],
					&exampleAuditLogEntryList.AuditLogEntries[2],
				),
			)

		actual, err := p.GetAuditLogEntries(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		expectedQuery, expectedArgs := p.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		expectedQuery, expectedArgs := p.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		expectedQuery, expectedArgs := p.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnRows(buildErroneousMockRowFromAuditLogEntry(exampleAuditLogEntry))

		actual, err := p.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildCreateAuditLogEntryQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		expectedQuery := "INSERT INTO audit_log (event_type,context) VALUES ($1,$2) RETURNING id, created_on"
		expectedArgs := []interface{}{
			exampleAuditLogEntry.EventType,
			exampleAuditLogEntry.Context,
		}
		actualQuery, actualArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_CreateAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		exampleInput := fakemodels.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleAuditLogEntry.ID, exampleAuditLogEntry.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnRows(exampleRows)

		actual, err := p.CreateAuditLogEntry(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		exampleInput := fakemodels.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				interfaceToDriverValue(expectedArgs)...,
			).
			WillReturnError(errors.New("blah"))

		actual, err := p.CreateAuditLogEntry(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

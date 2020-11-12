package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/converters"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/fake"

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

func TestSqlite_ScanAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, err := s.scanAuditLogEntries(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, err := s.scanAuditLogEntries(mockRows)
		assert.NoError(t, err)
	})
}

func TestSqlite_buildGetAuditLogEntryQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.id = ?"
		expectedArgs := []interface{}{
			exampleAuditLogEntry.ID,
		}
		actualQuery, actualArgs := s.buildGetAuditLogEntryQuery(exampleAuditLogEntry.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetAuditLogEntryQuery(exampleAuditLogEntry.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(exampleAuditLogEntry))

		actual, err := s.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		s, mockDB := buildTestService(t)

		expectedQuery, expectedArgs := s.buildGetAuditLogEntryQuery(exampleAuditLogEntry.ID)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllAuditLogEntriesCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(audit_log.id) FROM audit_log"
		actualQuery := s.buildGetAllAuditLogEntriesCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllAuditLogEntriesCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedCount := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(s.buildGetAllAuditLogEntriesCountQuery())).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := s.GetAllAuditLogEntriesCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetBatchOfAuditLogEntriesQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.id > ? AND audit_log.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := s.buildGetBatchOfAuditLogEntriesQuery(beginID, endID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetAllAuditLogEntries(T *testing.T) {
	T.Parallel()

	s, _ := buildTestService(T)
	expectedCountQuery := s.buildGetAllAuditLogEntriesCountQuery()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		exampleAuditLogEntryList := fakemodels.BuildFakeAuditLogEntryList()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := s.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromAuditLogEntries(
					&exampleAuditLogEntryList.Entries[0],
					&exampleAuditLogEntryList.Entries[1],
					&exampleAuditLogEntryList.Entries[2],
				),
			)

		out := make(chan []models.AuditLogEntry)
		doneChan := make(chan bool, 1)

		err := s.GetAllAuditLogEntries(ctx, out)
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

		s, mockDB := buildTestService(t)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		out := make(chan []models.AuditLogEntry)

		err := s.GetAllAuditLogEntries(ctx, out)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := s.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		out := make(chan []models.AuditLogEntry)

		err := s.GetAllAuditLogEntries(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := s.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		out := make(chan []models.AuditLogEntry)

		err := s.GetAllAuditLogEntries(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := s.buildGetBatchOfAuditLogEntriesQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromAuditLogEntry(exampleAuditLogEntry))

		out := make(chan []models.AuditLogEntry)

		err := s.GetAllAuditLogEntries(ctx, out)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAuditLogEntriesQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.created_on > ? AND audit_log.created_on < ? AND audit_log.last_updated_on > ? AND audit_log.last_updated_on < ? ORDER BY audit_log.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := s.buildGetAuditLogEntriesQuery(filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		exampleAuditLogEntryList := fakemodels.BuildFakeAuditLogEntryList()
		expectedQuery, expectedArgs := s.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromAuditLogEntries(
					&exampleAuditLogEntryList.Entries[0],
					&exampleAuditLogEntryList.Entries[1],
					&exampleAuditLogEntryList.Entries[2],
				),
			)

		actual, err := s.GetAuditLogEntries(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		expectedQuery, expectedArgs := s.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		expectedQuery, expectedArgs := s.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning item", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		filter := models.DefaultQueryFilter()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		expectedQuery, expectedArgs := s.buildGetAuditLogEntriesQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromAuditLogEntry(exampleAuditLogEntry))

		actual, err := s.GetAuditLogEntries(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateAuditLogEntryQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		expectedQuery := "INSERT INTO audit_log (event_type,context) VALUES (?,?)"
		expectedArgs := []interface{}{
			exampleAuditLogEntry.EventType,
			exampleAuditLogEntry.Context,
		}
		actualQuery, actualArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_createAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		exampleInput := fakemodels.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.createAuditLogEntry(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		exampleInput := fakemodels.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		s.createAuditLogEntry(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogSuccessfulLoginEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildSuccessfulLoginEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogSuccessfulLoginEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogUnsuccessfulLoginBadPasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUnsuccessfulLoginBadPasswordEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogUnsuccessfulLoginBadPasswordEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogUnsuccessfulLoginBad2FATokenEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogUnsuccessfulLoginBad2FATokenEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogLogoutEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildLogoutEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogLogoutEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogWebhookCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleAuditLogEntryInput := audit.BuildWebhookCreationEventEntry(exampleWebhook)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogWebhookCreationEvent(ctx, exampleWebhook)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogWebhookUpdateEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		exampleChanges := []models.FieldChangeSummary{}
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleAuditLogEntryInput := audit.BuildWebhookUpdateEventEntry(exampleWebhook.BelongsToUser, exampleWebhook.ID, exampleChanges)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogWebhookUpdateEvent(ctx, exampleWebhook.BelongsToUser, exampleWebhook.ID, exampleChanges)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogWebhookArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleAuditLogEntryInput := audit.BuildWebhookArchiveEventEntry(exampleWebhook.BelongsToUser, exampleWebhook.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogWebhookArchiveEvent(ctx, exampleWebhook.BelongsToUser, exampleWebhook.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

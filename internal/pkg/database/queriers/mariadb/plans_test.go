package mariadb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromPlans(includeCount bool, plans ...*types.Plan) *sqlmock.Rows {
	columns := queriers.PlansTableColumns

	if includeCount {
		columns = append(columns, "count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range plans {
		rowValues := []driver.Value{
			x.ID,
			x.Name,
			x.Description,
			x.Price,
			x.Period.String(),
			x.CreatedOn,
			x.LastUpdatedOn,
			x.ArchivedOn,
		}

		if includeCount {
			rowValues = append(rowValues, len(plans))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromPlan(x *types.Plan) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.PlansTableColumns).AddRow(
		x.Name,
		x.ID,
		x.Description,
		x.Price,
		x.Period.String(),
		x.CreatedOn,
		x.LastUpdatedOn,
		x.ArchivedOn,
	)

	return exampleRows
}

func TestMariaDB_ScanPlans(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := q.scanPlans(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := q.scanPlans(mockRows, false)
		assert.NoError(t, err)
	})
}

func TestMariaDB_buildGetPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "SELECT plans.id, plans.name, plans.description, plans.price, plans.period, plans.created_on, plans.last_updated_on, plans.archived_on FROM plans WHERE plans.id = ?"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.buildGetPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetPlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakePlan()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetPlanQuery(examplePlan.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromPlans(false, examplePlan))

		actual, err := q.GetPlan(ctx, examplePlan.ID)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakePlan()

		q, mockDB := buildTestService(t)

		expectedQuery, expectedArgs := q.buildGetPlanQuery(examplePlan.ID)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetPlan(ctx, examplePlan.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllPlansCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(plans.id) FROM plans WHERE plans.archived_on IS NULL"
		actualQuery := q.buildGetAllPlansCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_GetAllPlansCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedCount := uint64(123)

		q, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(q.buildGetAllPlansCountQuery())).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := q.GetAllPlansCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetPlansQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT plans.id, plans.name, plans.description, plans.price, plans.period, plans.created_on, plans.last_updated_on, plans.archived_on, (SELECT COUNT(*) FROM plans WHERE plans.archived_on IS NULL AND plans.created_on > ? AND plans.created_on < ? AND plans.last_updated_on > ? AND plans.last_updated_on < ?) FROM plans WHERE plans.archived_on IS NULL AND plans.created_on > ? AND plans.created_on < ? AND plans.last_updated_on > ? AND plans.last_updated_on < ? ORDER BY plans.created_on LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.buildGetPlansQuery(filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetPlans(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		examplePlanList := fakes.BuildFakePlanList()
		expectedQuery, expectedArgs := q.buildGetPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromPlans(
					true,
					&examplePlanList.Plans[0],
					&examplePlanList.Plans[1],
					&examplePlanList.Plans[2],
				),
			)

		actual, err := q.GetPlans(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, examplePlanList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		expectedQuery, expectedArgs := q.buildGetPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetPlans(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing read query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		expectedQuery, expectedArgs := q.buildGetPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.GetPlans(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error scanning plan", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.buildGetPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromPlan(examplePlan))

		actual, err := q.GetPlans(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreatePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "INSERT INTO plans (name,description,price,period) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			examplePlan.Name,
			examplePlan.Description,
			examplePlan.Price,
			examplePlan.Period.String(),
		}
		actualQuery, actualArgs := q.buildCreatePlanQuery(examplePlan)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_CreatePlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		expectedQuery, expectedArgs := q.buildCreatePlanQuery(examplePlan)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(int64(examplePlan.ID), 1))

		mtt := &queriers.MockTimeTeller{}
		mtt.On("Now").Return(examplePlan.CreatedOn)
		q.timeTeller = mtt

		actual, err := q.CreatePlan(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		mock.AssertExpectationsForObjects(t, mtt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		q, mockDB := buildTestService(t)
		examplePlan := fakes.BuildFakePlan()

		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		expectedQuery, expectedArgs := q.buildCreatePlanQuery(examplePlan)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.CreatePlan(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdatePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "UPDATE plans SET name = ?, description = ?, price = ?, period = ?, last_updated_on = UNIX_TIMESTAMP() WHERE id = ?"
		expectedArgs := []interface{}{
			examplePlan.Name,
			examplePlan.Description,
			examplePlan.Price,
			examplePlan.Period.String(),
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.buildUpdatePlanQuery(examplePlan)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_UpdatePlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		q, mockDB := buildTestService(t)
		examplePlan := fakes.BuildFakePlan()
		expectedQuery, expectedArgs := q.buildUpdatePlanQuery(examplePlan)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(int64(examplePlan.ID), 1))

		err := q.UpdatePlan(ctx, examplePlan)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		q, mockDB := buildTestService(t)
		examplePlan := fakes.BuildFakePlan()
		expectedQuery, expectedArgs := q.buildUpdatePlanQuery(examplePlan)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := q.UpdatePlan(ctx, examplePlan)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchivePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "UPDATE plans SET last_updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.buildArchivePlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_ArchivePlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.buildArchivePlanQuery(examplePlan.ID)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := q.ArchivePlan(ctx, examplePlan.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("returns sql.ErrNoRows with no rows affected", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.buildArchivePlanQuery(examplePlan.ID)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := q.ArchivePlan(ctx, examplePlan.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.buildArchivePlanQuery(examplePlan.ID)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := q.ArchivePlan(ctx, examplePlan.ID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_LogPlanCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakePlan()
		exampleAuditLogEntryInput := audit.BuildPlanCreationEventEntry(exampleInput)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		q.LogPlanCreationEvent(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_LogPlanUpdateEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleUser := fakes.BuildFakeUser()
		exampleChanges := []types.FieldChangeSummary{}
		exampleInput := fakes.BuildFakePlan()
		exampleAuditLogEntryInput := audit.BuildPlanUpdateEventEntry(exampleUser.ID, exampleInput.ID, exampleChanges)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		q.LogPlanUpdateEvent(ctx, exampleUser.ID, exampleInput.ID, exampleChanges)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_LogPlanArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakePlan()
		exampleAuditLogEntryInput := audit.BuildPlanArchiveEventEntry(exampleUser.ID, exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		q.LogPlanArchiveEvent(ctx, exampleUser.ID, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAuditLogEntriesForPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := fmt.Sprintf("SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE JSON_CONTAINS(audit_log.context, '%d', '$.plan_id') ORDER BY audit_log.created_on", examplePlan.ID)
		expectedArgs := []interface{}(nil)
		actualQuery, actualArgs := q.buildGetAuditLogEntriesForPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

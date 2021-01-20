package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromPlans(includeCounts bool, filteredCount uint64, plans ...*types.AccountSubscriptionPlan) *sqlmock.Rows {
	columns := queriers.AccountSubscriptionPlansTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
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

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(plans))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromPlan(x *types.AccountSubscriptionPlan) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.AccountSubscriptionPlansTableColumns).AddRow(
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

func TestPostgres_ScanPlans(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanAccountSubscriptionPlans(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanAccountSubscriptionPlans(mockRows, false)
		assert.Error(t, err)
	})
}

func TestPostgres_buildGetPlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.id = $1"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountSubscriptionPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetPlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakePlan()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildGetAccountSubscriptionPlanQuery(examplePlan.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromPlans(false, 0, examplePlan))

		actual, err := q.GetAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakePlan()

		q, mockDB := buildTestService(t)

		expectedQuery, expectedArgs := q.BuildGetAccountSubscriptionPlanQuery(examplePlan.ID)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllPlansCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL"
		actualQuery := q.BuildGetAllAccountSubscriptionPlansCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllPlansCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedCount := uint64(123)

		q, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(q.BuildGetAllAccountSubscriptionPlansCountQuery())).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := q.GetAllAccountSubscriptionPlansCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetPlansQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT account_subscription_plans.id, account_subscription_plans.name, account_subscription_plans.description, account_subscription_plans.price, account_subscription_plans.period, account_subscription_plans.created_on, account_subscription_plans.last_updated_on, account_subscription_plans.archived_on, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL) as total_count, (SELECT COUNT(account_subscription_plans.id) FROM account_subscription_plans WHERE account_subscription_plans.archived_on IS NULL AND account_subscription_plans.created_on > $1 AND account_subscription_plans.created_on < $2 AND account_subscription_plans.last_updated_on > $3 AND account_subscription_plans.last_updated_on < $4) as filtered_count FROM account_subscription_plans WHERE account_subscription_plans.created_on > $5 AND account_subscription_plans.created_on < $6 AND account_subscription_plans.last_updated_on > $7 AND account_subscription_plans.last_updated_on < $8 GROUP BY account_subscription_plans.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.BuildGetAccountSubscriptionPlansQuery(filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetPlans(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		examplePlanList := fakes.BuildFakePlanList()
		expectedQuery, expectedArgs := q.BuildGetAccountSubscriptionPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromPlans(
					true,
					examplePlanList.FilteredCount,
					examplePlanList.AccountSubscriptionPlans...,
				),
			)

		actual, err := q.GetAccountSubscriptionPlans(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, examplePlanList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		filter := types.DefaultQueryFilter()

		expectedQuery, expectedArgs := q.BuildGetAccountSubscriptionPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetAccountSubscriptionPlans(ctx, filter)
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

		expectedQuery, expectedArgs := q.BuildGetAccountSubscriptionPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.GetAccountSubscriptionPlans(ctx, filter)
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

		expectedQuery, expectedArgs := q.BuildGetAccountSubscriptionPlansQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromPlan(examplePlan))

		actual, err := q.GetAccountSubscriptionPlans(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildCreatePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "INSERT INTO account_subscription_plans (name,description,price,period) VALUES ($1,$2,$3,$4) RETURNING id, created_on"
		expectedArgs := []interface{}{
			examplePlan.Name,
			examplePlan.Description,
			examplePlan.Price,
			examplePlan.Period.String(),
		}
		actualQuery, actualArgs := q.BuildCreateAccountSubscriptionPlanQuery(examplePlan)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_CreatePlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		expectedQuery, expectedArgs := q.BuildCreateAccountSubscriptionPlanQuery(examplePlan)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(examplePlan.ID, examplePlan.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		actual, err := q.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		q, mockDB := buildTestService(t)
		examplePlan := fakes.BuildFakePlan()

		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		expectedQuery, expectedArgs := q.BuildCreateAccountSubscriptionPlanQuery(examplePlan)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdatePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "UPDATE account_subscription_plans SET name = $1, description = $2, price = $3, period = $4, last_updated_on = extract(epoch FROM NOW()) WHERE id = $5 RETURNING last_updated_on"
		expectedArgs := []interface{}{
			examplePlan.Name,
			examplePlan.Description,
			examplePlan.Price,
			examplePlan.Period.String(),
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateAccountSubscriptionPlanQuery(examplePlan)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_UpdatePlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.BuildUpdateAccountSubscriptionPlanQuery(examplePlan)

		exampleRows := sqlmock.NewRows([]string{"last_updated_on"}).AddRow(uint64(time.Now().Unix()))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		err := q.UpdateAccountSubscriptionPlan(ctx, examplePlan)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.BuildUpdateAccountSubscriptionPlanQuery(examplePlan)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := q.UpdateAccountSubscriptionPlan(ctx, examplePlan)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchivePlanQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery := "UPDATE account_subscription_plans SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $1 RETURNING archived_on"
		expectedArgs := []interface{}{
			examplePlan.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountSubscriptionPlanQuery(examplePlan.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_ArchivePlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.BuildArchiveAccountSubscriptionPlanQuery(examplePlan.ID)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := q.ArchiveAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("returns sql.ErrNoRows with no rows affected", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.BuildArchiveAccountSubscriptionPlanQuery(examplePlan.ID)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := q.ArchiveAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		examplePlan := fakes.BuildFakePlan()

		expectedQuery, expectedArgs := q.BuildArchiveAccountSubscriptionPlanQuery(examplePlan.ID)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := q.ArchiveAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogPlanCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakePlan()
		exampleAuditLogEntryInput := audit.BuildAccountSubscriptionPlanCreationEventEntry(exampleInput)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.BuildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogAccountSubscriptionPlanCreationEvent(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogPlanUpdateEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleUser := fakes.BuildFakeUser()
		exampleChanges := []types.FieldChangeSummary{}
		exampleInput := fakes.BuildFakePlan()
		exampleAuditLogEntryInput := audit.BuildAccountSubscriptionPlanUpdateEventEntry(exampleUser.ID, exampleInput.ID, exampleChanges)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.BuildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.AccountSubscriptionLogPlanUpdateEvent(ctx, exampleUser.ID, exampleInput.ID, exampleChanges)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogPlanArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakePlan()
		exampleAuditLogEntryInput := audit.BuildAccountSubscriptionPlanArchiveEventEntry(exampleUser.ID, exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.BuildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.AccountSubscriptionLogPlanArchiveEvent(ctx, exampleUser.ID, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

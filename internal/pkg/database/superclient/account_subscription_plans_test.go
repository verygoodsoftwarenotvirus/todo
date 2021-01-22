package superclient

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAccountSubscriptionPlans(includeCounts bool, filteredCount uint64, plans ...*types.AccountSubscriptionPlan) *sqlmock.Rows {
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

func buildErroneousMockRowFromAccountSubscriptionPlan(x *types.AccountSubscriptionPlan) *sqlmock.Rows {
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

func TestSqlite_ScanPlans(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestClient(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanAccountSubscriptionPlans(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestClient(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanAccountSubscriptionPlans(mockRows, false)
		assert.Error(t, err)
	})
}

func TestClient_GetPlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		examplePlan := fakes.BuildFakePlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.
			On("BuildGetAccountSubscriptionPlanQuery", examplePlan.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountSubscriptionPlans(false, 0, examplePlan))

		actual, err := c.GetAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllPlansCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.
			On("BuildGetAllAccountSubscriptionPlansCountQuery").
			Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllAccountSubscriptionPlansCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetPlans(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		examplePlanList := fakes.BuildFakePlanList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.
			On("BuildGetAccountSubscriptionPlansQuery", filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountSubscriptionPlans(true, examplePlanList.FilteredCount, examplePlanList.AccountSubscriptionPlans...))

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, examplePlanList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		examplePlanList := fakes.BuildFakePlanList()
		examplePlanList.Page, examplePlanList.Limit = 0, 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.
			On("BuildGetAccountSubscriptionPlansQuery", filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountSubscriptionPlans(true, examplePlanList.FilteredCount, examplePlanList.AccountSubscriptionPlans...))

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, examplePlanList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_CreatePlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.
			On("BuildCreateAccountSubscriptionPlanQuery", mock.MatchedBy(matchAccountSubscriptionPlan(t, examplePlan))).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(examplePlan.ID))

		stt := &queriers.MockTimeTeller{}
		stt.On("Now").Return(examplePlan.CreatedOn)
		c.timeTeller = stt

		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_UpdatePlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		examplePlan := fakes.BuildFakePlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.
			On("BuildUpdateAccountSubscriptionPlanQuery", examplePlan).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(examplePlan.ID))

		err := c.UpdateAccountSubscriptionPlan(ctx, examplePlan)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_ArchivePlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		examplePlan := fakes.BuildFakePlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.
			On("BuildArchiveAccountSubscriptionPlanQuery", examplePlan.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(examplePlan.ID))

		err := c.ArchiveAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

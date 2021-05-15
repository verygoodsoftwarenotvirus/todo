package querier

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAccountSubscriptionPlans(includeCounts bool, filteredCount uint64, plans ...*types.AccountSubscriptionPlan) *sqlmock.Rows {
	columns := querybuilding.AccountSubscriptionPlansTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range plans {
		rowValues := []driver.Value{
			x.ID,
			x.ExternalID,
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

func TestQuerier_ScanPlans(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanAccountSubscriptionPlans(ctx, mockRows, false)
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

		_, _, _, err := q.scanAccountSubscriptionPlans(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_GetPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountSubscriptionPlans(false, 0, exampleAccountSubscriptionPlan))

		actual, err := c.GetAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountSubscriptionPlan, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAccountSubscriptionPlan(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		expectedError := errors.New(t.Name())

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(expectedError)

		actual, err := c.GetAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedError))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid time in database", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		exampleRows := sqlmock.NewRows(querybuilding.AccountSubscriptionPlansTableColumns).AddRow(
			exampleAccountSubscriptionPlan.ID,
			exampleAccountSubscriptionPlan.ExternalID,
			exampleAccountSubscriptionPlan.Name,
			exampleAccountSubscriptionPlan.Description,
			exampleAccountSubscriptionPlan.Price,
			"this doesn't parse lol",
			exampleAccountSubscriptionPlan.CreatedOn,
			exampleAccountSubscriptionPlan.LastUpdatedOn,
			exampleAccountSubscriptionPlan.ArchivedOn,
		)

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(exampleRows)

		actual, err := c.GetAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAllAccountSubscriptionPlansCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAllAccountSubscriptionPlansCountQuery", testutil.ContextMatcher).Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllAccountSubscriptionPlansCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New(t.Name())

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAllAccountSubscriptionPlansCountQuery", testutil.ContextMatcher).Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(expectedError)

		actual, err := c.GetAllAccountSubscriptionPlansCount(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedError))
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAccountSubscriptionPlans(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountSubscriptionPlanList := fakes.BuildFakeAccountSubscriptionPlanList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAccountSubscriptionPlansQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountSubscriptionPlans(true, exampleAccountSubscriptionPlanList.FilteredCount, exampleAccountSubscriptionPlanList.AccountSubscriptionPlans...))

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountSubscriptionPlanList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleAccountSubscriptionPlanList := fakes.BuildFakeAccountSubscriptionPlanList()
		exampleAccountSubscriptionPlanList.Page, exampleAccountSubscriptionPlanList.Limit = 0, 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAccountSubscriptionPlansQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountSubscriptionPlans(true, exampleAccountSubscriptionPlanList.FilteredCount, exampleAccountSubscriptionPlanList.AccountSubscriptionPlans...))

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountSubscriptionPlanList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New(t.Name())
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAccountSubscriptionPlansQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(expectedError)

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedError))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response to query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAccountSubscriptionPlansQuery",
			testutil.ContextMatcher,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_CreateAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleAccountSubscriptionPlan.ExternalID = ""
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildCreateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleAccountSubscriptionPlan.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountSubscriptionPlan, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAccountSubscriptionPlan(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleAccountSubscriptionPlan.ExternalID = ""
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("error executing query", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New(t.Name())
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildCreateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(expectedError)

		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedError))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error creating transaction", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleAccountSubscriptionPlan.ExternalID = ""
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildCreateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccountSubscriptionPlan.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleAccountSubscriptionPlan.ExternalID = ""
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(exampleAccountSubscriptionPlan)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildCreateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleAccountSubscriptionPlan.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_UpdateAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildUpdateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.UpdateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan, 0, nil)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		err := c.UpdateAccountSubscriptionPlan(ctx, nil, 0, nil)
		assert.Error(t, err)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		err := c.UpdateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan, 0, nil)
		assert.Error(t, err)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New(t.Name())
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildUpdateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(expectedError)

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.UpdateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan, 0, nil)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedError))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error creating audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildUpdateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.UpdateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan, 0, nil)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildUpdateAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.UpdateAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan, 0, nil)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_ArchiveAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildArchiveAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.ArchiveAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID, 0)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		err := c.ArchiveAccountSubscriptionPlan(ctx, 0, 0)
		assert.Error(t, err)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		err := c.ArchiveAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID, 0)
		assert.Error(t, err)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New(t.Name())
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildArchiveAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(expectedError)

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.ArchiveAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID, 0)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedError))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildArchiveAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.ArchiveAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID, 0)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildArchiveAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountSubscriptionPlan.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.ArchiveAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID, 0)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAuditLogEntriesForAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntriesForAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntriesForAccountSubscriptionPlan(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AccountSubscriptionPlanSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery",
			testutil.ContextMatcher,
			exampleAccountSubscriptionPlan.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForAccountSubscriptionPlan(ctx, exampleAccountSubscriptionPlan.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}
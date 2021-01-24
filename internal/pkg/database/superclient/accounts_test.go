package superclient

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAccounts(includeCounts bool, filteredCount uint64, accounts ...*types.Account) *sqlmock.Rows {
	columns := queriers.AccountsTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range accounts {
		rowValues := []driver.Value{
			x.ID,
			x.Name,
			x.PlanID,
			x.PersonalAccount,
			x.CreatedOn,
			x.LastUpdatedOn,
			x.ArchivedOn,
			x.BelongsToUser,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(accounts))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromAccount(x *types.Account) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.AccountsTableColumns).AddRow(
		x.ArchivedOn,
		x.Name,
		x.PlanID,
		x.PersonalAccount,
		x.CreatedOn,
		x.LastUpdatedOn,
		x.BelongsToUser,
		x.ID,
	)

	return exampleRows
}

func TestClient_GetAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildGetAccountQuery", exampleAccount.ID, exampleUser.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(false, 0, exampleAccount))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleAccount.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllAccountsCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildGetAllAccountsCountQuery").
			Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllAccountsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllAccounts(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.Account)
		expectedCount := uint64(20)
		doneChan := make(chan bool, 1)
		exampleBatchSize := uint16(1000)
		exampleAccountList := fakes.BuildFakeAccountList()

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildGetAllAccountsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildGetBatchOfAccountsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(false, 0, exampleAccountList.Accounts...))

		err := c.GetAllAccounts(ctx, results, exampleBatchSize)
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

func TestClient_GetAccounts(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountList := fakes.BuildFakeAccountList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildGetAccountsQuery", exampleUser.ID, false, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleAccountList := fakes.BuildFakeAccountList()
		exampleAccountList.Page, exampleAccountList.Limit = 0, 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildGetAccountsQuery", exampleUser.ID, false, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_CreateAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildCreateAccountQuery", exampleInput).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_UpdateAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildUpdateAccountQuery", exampleAccount).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		err := c.UpdateAccount(ctx, exampleAccount)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_ArchiveAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.
			On("BuildArchiveAccountQuery", exampleAccount.ID, exampleUser.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		err := c.ArchiveAccount(ctx, exampleAccount.ID, exampleAccount.BelongsToUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

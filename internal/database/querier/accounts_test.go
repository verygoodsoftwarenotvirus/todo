package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAccounts(includeCounts bool, filteredCount uint64, accounts ...*types.Account) *sqlmock.Rows {
	columns := append(querybuilding.AccountsTableColumns, querybuilding.AccountsUserMembershipTableColumns...)

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range accounts {
		for _, y := range x.Members {
			rowValues := []driver.Value{
				x.ID,
				x.ExternalID,
				x.Name,
				x.BillingStatus,
				x.ContactEmail,
				x.ContactPhone,
				x.PaymentProcessorCustomerID,
				x.SubscriptionPlanID,
				x.CreatedOn,
				x.LastUpdatedOn,
				x.ArchivedOn,
				x.BelongsToUser,
				y.ID,
				y.BelongsToUser,
				y.BelongsToAccount,
				strings.Join(y.AccountRoles, accountMemberRolesSeparator),
				y.DefaultAccount,
				y.CreatedOn,
				x.LastUpdatedOn,
				y.ArchivedOn,
			}

			if includeCounts {
				rowValues = append(rowValues, filteredCount, len(accounts))
			}

			exampleRows.AddRow(rowValues...)
		}
	}

	return exampleRows
}

func TestQuerier_ScanAccounts(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanAccounts(ctx, mockRows, false)
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

		_, _, _, err := q.scanAccounts(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_GetAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(false, 0, exampleAccount))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, _ := buildTestClient(t)

		actual, err := c.GetAccount(ctx, 0, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, _ := buildTestClient(t)

		actual, err := c.GetAccount(ctx, exampleAccount.ID, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with no returned accounts", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		columns := append(querybuilding.AccountsTableColumns, querybuilding.AccountsUserMembershipTableColumns...)

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows(columns))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAllAccountsCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleCount := uint64(123)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAllAccountsCountQuery", testutils.ContextMatcher).Return(fakeQuery)
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

func TestQuerier_GetAllAccounts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.Account)
		expectedCount := uint64(20)
		doneChan := make(chan bool, 1)
		exampleAccountList := fakes.BuildFakeAccountList()

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAllAccountsCountQuery", testutils.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetBatchOfAccountsQuery",
			testutils.ContextMatcher,
			uint64(1), uint64(defaultBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(false, 0, exampleAccountList.Accounts...))

		err := c.GetAllAccounts(ctx, results, 0)
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

	T.Run("with nil results channel", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		c, _ := buildTestClient(t)

		assert.Error(t, c.GetAllAccounts(ctx, nil, 0))
	})

	T.Run("with now rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.Account)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAllAccountsCountQuery", testutils.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetBatchOfAccountsQuery",
			testutils.ContextMatcher,
			uint64(1), uint64(exampleBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		err := c.GetAllAccounts(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.Account)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAllAccountsCountQuery", testutils.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllAccounts(ctx, results, exampleBatchSize)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.Account)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAllAccountsCountQuery", testutils.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetBatchOfAccountsQuery",
			testutils.ContextMatcher,
			uint64(1), uint64(exampleBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(errors.New("blah"))

		err := c.GetAllAccounts(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		results := make(chan []*types.Account)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAllAccountsCountQuery", testutils.ContextMatcher).Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetBatchOfAccountsQuery",
			testutils.ContextMatcher,
			uint64(1), uint64(exampleBatchSize+1),
		).Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		err := c.GetAllAccounts(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAccounts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		exampleAccountList := fakes.BuildFakeAccountList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			exampleUser.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAccounts(ctx, 0, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleUser := fakes.BuildFakeUser()
		exampleAccountList := fakes.BuildFakeAccountList()
		exampleAccountList.Page, exampleAccountList.Limit = 0, 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			exampleUser.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			exampleUser.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			exampleUser.ID,
			false,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAccountsForAdmin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountList := fakes.BuildFakeAccountList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			uint64(0),
			true,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccountsForAdmin(ctx, filter)
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
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			uint64(0),
			true,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccountsForAdmin(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			uint64(0),
			true,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccountsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAccountsQuery",
			testutils.ContextMatcher,
			uint64(0),
			true,
			filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccountsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_CreateAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ExternalID = ""
		exampleAccount.BillingStatus = types.UnpaidAccountBillingStatus
		exampleAccount.PaymentProcessorCustomerID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
		exampleAccountAdditionInput := &types.AddUserToAccountInput{
			Reason:       "account creation",
			UserID:       exampleUser.ID,
			AccountID:    exampleAccount.ID,
			AccountRoles: []string{authorization.AccountAdminRole.String()},
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildAccountCreationQuery",
			testutils.ContextMatcher,
			exampleCreationInput,
		).Return(fakeCreationQuery, fakeCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationAuditLogEntryArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(123))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.On(
			"BuildAddUserToAccountQuery",
			testutils.ContextMatcher,
			exampleAccountAdditionInput,
		).Return(fakeAccountAdditionQuery, fakeAccountAdditionArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeAccountAdditionAuditLogEntryQuery, fakeAccountAdditionAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeAccountAdditionAuditLogEntryQuery, fakeAccountAdditionAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionAuditLogEntryArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(123))

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccount(ctx, exampleCreationInput, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ExternalID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAccount(ctx, exampleCreationInput, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAccount(ctx, nil, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ExternalID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		actual, err := c.CreateAccount(ctx, exampleCreationInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership{}
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildAccountCreationQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing account creation audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ExternalID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildAccountCreationQuery",
			testutils.ContextMatcher,
			exampleCreationInput,
		).Return(fakeCreationQuery, fakeCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationAuditLogEntryArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccount(ctx, exampleCreationInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing user addition to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ExternalID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
		exampleAccountAdditionInput := &types.AddUserToAccountInput{
			Reason:       "account creation",
			UserID:       exampleUser.ID,
			AccountID:    exampleAccount.ID,
			AccountRoles: []string{authorization.AccountAdminRole.String()},
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildAccountCreationQuery",
			testutils.ContextMatcher,
			exampleCreationInput,
		).Return(fakeCreationQuery, fakeCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationAuditLogEntryArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(123))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.On(
			"BuildAddUserToAccountQuery",
			testutils.ContextMatcher,
			exampleAccountAdditionInput,
		).Return(fakeAccountAdditionQuery, fakeAccountAdditionArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccount(ctx, exampleCreationInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing user membership addition audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ExternalID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
		exampleAccountAdditionInput := &types.AddUserToAccountInput{
			Reason:       "account creation",
			UserID:       exampleUser.ID,
			AccountID:    exampleAccount.ID,
			AccountRoles: []string{authorization.AccountAdminRole.String()},
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildAccountCreationQuery",
			testutils.ContextMatcher,
			exampleCreationInput,
		).Return(fakeCreationQuery, fakeCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationAuditLogEntryArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(123))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.On(
			"BuildAddUserToAccountQuery",
			testutils.ContextMatcher,
			exampleAccountAdditionInput,
		).Return(fakeAccountAdditionQuery, fakeAccountAdditionArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeAccountAdditionAuditLogEntryQuery, fakeAccountAdditionAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeAccountAdditionAuditLogEntryQuery, fakeAccountAdditionAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionAuditLogEntryArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccount(ctx, exampleCreationInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ExternalID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
		exampleAccountAdditionInput := &types.AddUserToAccountInput{
			Reason:       "account creation",
			UserID:       exampleUser.ID,
			AccountID:    exampleAccount.ID,
			AccountRoles: []string{authorization.AccountAdminRole.String()},
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildAccountCreationQuery",
			testutils.ContextMatcher,
			exampleCreationInput,
		).Return(fakeCreationQuery, fakeCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeCreationAuditLogEntryQuery, fakeCreationAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationAuditLogEntryArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(123))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.On(
			"BuildAddUserToAccountQuery",
			testutils.ContextMatcher,
			exampleAccountAdditionInput,
		).Return(fakeAccountAdditionQuery, fakeAccountAdditionArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		fakeAccountAdditionAuditLogEntryQuery, fakeAccountAdditionAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{}),
		).Return(fakeAccountAdditionAuditLogEntryQuery, fakeAccountAdditionAuditLogEntryArgs).Once()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionAuditLogEntryQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionAuditLogEntryArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(123))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAccount(ctx, exampleCreationInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_UpdateAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildUpdateAccountQuery",
			testutils.ContextMatcher,
			exampleAccount,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		assert.NoError(t, c.UpdateAccount(ctx, exampleAccount, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateAccount(ctx, nil, exampleUser.ID, nil))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount, 0, nil))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildUpdateAccountQuery",
			testutils.ContextMatcher,
			exampleAccount,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildUpdateAccountQuery",
			testutils.ContextMatcher,
			exampleAccount,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildUpdateAccountQuery",
			testutils.ContextMatcher,
			exampleAccount,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount, exampleUser.ID, nil))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_ArchiveAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildArchiveAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		assert.NoError(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAccount(ctx, 0, exampleUser.ID, exampleUser.ID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, 0, exampleUser.ID))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildArchiveAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildArchiveAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildArchiveAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID, exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAuditLogEntriesForAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntriesForAccount(ctx, exampleAccount.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntriesForAccount(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForAccount(ctx, exampleAccount.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAccountQuery",
			testutils.ContextMatcher,
			exampleAccount.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForAccount(ctx, exampleAccount.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

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

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(false, 0, exampleAccount))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, _ := buildTestClient(t)

		actual, err := c.GetAccount(ctx, "", exampleUser.ID)
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

		actual, err := c.GetAccount(ctx, exampleAccount.ID, "")
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

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with no returned accounts", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		columns := append(querybuilding.AccountsTableColumns, querybuilding.AccountsUserMembershipTableColumns...)

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows(columns))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAllAccountsCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleCount := uint64(123)

		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllAccountsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAccounts(ctx, "", filter)
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

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccountsForAdmin(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleAccountList := fakes.BuildFakeAccountList()
		exampleAccountList.Page, exampleAccountList.Limit = 0, 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccountsForAdmin(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccountsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccountsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_CreateAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BillingStatus = types.UnpaidAccountBillingStatus
		exampleAccount.PaymentProcessorCustomerID = ""
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleCreationInput)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual.ID)
		actual.ID = ""

		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAccount(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		actual, err := c.CreateAccount(ctx, exampleCreationInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership{}
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing account creation audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleCreationInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing user addition to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleCreationInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing user membership addition audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleCreationInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleCreationInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeCreationQuery, fakeCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountAdditionQuery, fakeAccountAdditionArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountAdditionQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountAdditionArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleCreationInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.NoError(t, c.UpdateAccount(ctx, exampleAccount))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateAccount(ctx, nil))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount))

		mock.AssertExpectationsForObjects(t, db)
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

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.NoError(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAccount(ctx, "", exampleUser.ID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})
}

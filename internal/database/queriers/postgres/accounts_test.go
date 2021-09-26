package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func buildMockRowsFromAccounts(includeCounts bool, filteredCount uint64, accounts ...*types.Account) *sqlmock.Rows {
	columns := append(accountsTableColumns, accountsUserMembershipTableColumns...)

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

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		c, db := buildTestClient(t)

		args := []interface{}{
			exampleUserID,
			exampleAccount.ID,
		}

		db.ExpectQuery(formatQueryForSQLMock(getAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromAccounts(false, 0, exampleAccount))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		c, _ := buildTestClient(t)

		actual, err := c.GetAccount(ctx, "", exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		c, _ := buildTestClient(t)

		actual, err := c.GetAccount(ctx, exampleAccount.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		c, db := buildTestClient(t)

		args := []interface{}{
			exampleUserID,
			exampleAccount.ID,
		}

		db.ExpectQuery(formatQueryForSQLMock(getAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		c, db := buildTestClient(t)

		args := []interface{}{
			exampleUserID,
			exampleAccount.ID,
		}

		db.ExpectQuery(formatQueryForSQLMock(getAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with no returned accounts", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		c, db := buildTestClient(t)

		columns := append(accountsTableColumns, accountsUserMembershipTableColumns...)

		args := []interface{}{
			exampleUserID,
			exampleAccount.ID,
		}

		db.ExpectQuery(formatQueryForSQLMock(getAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(sqlmock.NewRows(columns))

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleUserID)
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

		db.ExpectQuery(formatQueryForSQLMock(getAllAccountsCountQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllAccountsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectQuery(formatQueryForSQLMock(getAllAccountsCountQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAllAccountsCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAccounts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountList := fakes.BuildFakeAccountList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildGetAccountsQuery(ctx, exampleUserID, false, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUserID, filter)
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
		exampleUserID := fakes.BuildFakeID()
		exampleAccountList := fakes.BuildFakeAccountList()
		exampleAccountList.Page, exampleAccountList.Limit = 0, 0

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildGetAccountsQuery(ctx, exampleUserID, false, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromAccounts(true, exampleAccountList.FilteredCount, exampleAccountList.Accounts...))

		actual, err := c.GetAccounts(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUserID := fakes.BuildFakeID()
		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildGetAccountsQuery(ctx, exampleUserID, false, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccounts(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUserID := fakes.BuildFakeID()
		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildGetAccountsQuery(ctx, exampleUserID, false, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccounts(ctx, exampleUserID, filter)
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

		query, args := c.buildGetAccountsQuery(ctx, "", true, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
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

		query, args := c.buildGetAccountsQuery(ctx, "", true, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
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

		query, args := c.buildGetAccountsQuery(ctx, "", true, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
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

		query, args := c.buildGetAccountsQuery(ctx, "", true, filter)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
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

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BillingStatus = types.UnpaidAccountBillingStatus
		exampleAccount.PaymentProcessorCustomerID = ""
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUserID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		accountCreationArgs := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			types.UnpaidAccountBillingStatus,
			exampleInput.ContactEmail,
			exampleInput.ContactPhone,
			exampleInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		addUserToAccountArgs := []interface{}{
			&idMatcher{},
			exampleInput.BelongsToUser,
			&idMatcher{},
			strings.Join([]string{authorization.AccountAdminRole.String()}, accountMemberRolesSeparator),
		}

		db.ExpectExec(formatQueryForSQLMock(addUserToAccountDuringCreationQuery)).
			WithArgs(interfaceToDriverValue(addUserToAccountArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleInput)
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

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUserID
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

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUserID
		exampleAccount.Members = []*types.AccountUserMembership{}
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		accountCreationArgs := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			types.UnpaidAccountBillingStatus,
			exampleInput.ContactEmail,
			exampleInput.ContactPhone,
			exampleInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
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

	T.Run("with error writing user addition to database", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUserID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()
		accountCreationArgs := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			types.UnpaidAccountBillingStatus,
			exampleInput.ContactEmail,
			exampleInput.ContactPhone,
			exampleInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		addUserToAccountArgs := []interface{}{
			&idMatcher{},
			exampleInput.BelongsToUser,
			&idMatcher{},
			strings.Join([]string{authorization.AccountAdminRole.String()}, accountMemberRolesSeparator),
		}

		db.ExpectExec(formatQueryForSQLMock(addUserToAccountDuringCreationQuery)).
			WithArgs(interfaceToDriverValue(addUserToAccountArgs)...).
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

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.ID = ""
		exampleAccount.BelongsToUser = exampleUserID
		exampleAccount.Members = []*types.AccountUserMembership(nil)
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()
		accountCreationArgs := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			types.UnpaidAccountBillingStatus,
			exampleInput.ContactEmail,
			exampleInput.ContactPhone,
			exampleInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		addUserToAccountArgs := []interface{}{
			&idMatcher{},
			exampleInput.BelongsToUser,
			&idMatcher{},
			strings.Join([]string{authorization.AccountAdminRole.String()}, accountMemberRolesSeparator),
		}

		db.ExpectExec(formatQueryForSQLMock(addUserToAccountDuringCreationQuery)).
			WithArgs(interfaceToDriverValue(addUserToAccountArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleAccount.CreatedOn
		}

		actual, err := c.CreateAccount(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleAccount.Name,
			exampleAccount.ContactEmail,
			exampleAccount.ContactPhone,
			exampleAccount.BelongsToUser,
			exampleAccount.ID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		assert.NoError(t, c.UpdateAccount(ctx, exampleAccount))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateAccount(ctx, nil))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleAccount.Name,
			exampleAccount.ContactEmail,
			exampleAccount.ContactPhone,
			exampleAccount.BelongsToUser,
			exampleAccount.ID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateAccount(ctx, exampleAccount))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_ArchiveAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleUserID,
			exampleAccountID,
		}

		db.ExpectExec(formatQueryForSQLMock(archiveAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		assert.NoError(t, c.ArchiveAccount(ctx, exampleAccountID, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAccount(ctx, "", exampleUserID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccountID, ""))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleUserID,
			exampleAccountID,
		}

		db.ExpectExec(formatQueryForSQLMock(archiveAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccountID, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})
}

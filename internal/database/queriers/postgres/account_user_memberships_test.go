package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAccountUserMemberships(memberships ...*types.AccountUserMembership) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(accountsUserMembershipTableColumns)

	for _, x := range memberships {
		rowValues := []driver.Value{
			&x.ID,
			&x.BelongsToUser,
			&x.BelongsToAccount,
			strings.Join(x.AccountRoles, accountMemberRolesSeparator),
			&x.DefaultAccount,
			&x.CreatedOn,
			&x.LastUpdatedOn,
			&x.ArchivedOn,
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildInvalidMockRowsFromAccountUserMemberships(memberships ...*types.AccountUserMembership) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(accountsUserMembershipTableColumns)

	for _, x := range memberships {
		rowValues := []driver.Value{
			&x.DefaultAccount,
			&x.BelongsToUser,
			&x.BelongsToAccount,
			strings.Join(x.AccountRoles, accountMemberRolesSeparator),
			&x.CreatedOn,
			&x.LastUpdatedOn,
			&x.ArchivedOn,
			&x.ID,
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanAccountUserMemberships(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := q.scanAccountUserMemberships(ctx, mockRows)
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

		_, _, err := q.scanAccountUserMemberships(ctx, mockRows)
		assert.Error(t, err)
	})
}

func TestQuerier_BuildSessionContextDataForUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.Members[0].DefaultAccount = true

		examplePermsMap := map[string]*types.UserAccountMembershipInfo{}
		for _, membership := range exampleAccount.Members {
			examplePermsMap[membership.BelongsToAccount] = &types.UserAccountMembershipInfo{
				AccountName:  exampleAccount.Name,
				AccountID:    membership.BelongsToAccount,
				AccountRoles: membership.AccountRoles,
			}
		}

		exampleAccountPermissionsMap := map[string]authorization.AccountRolePermissionsChecker{}
		for _, membership := range exampleAccount.Members {
			exampleAccountPermissionsMap[membership.BelongsToAccount] = authorization.NewAccountRolePermissionChecker(membership.AccountRoles...)
		}

		c, db := buildTestClient(t)

		userRetrievalArgs := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(userRetrievalArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		getAccountMembershipsForUserArgs := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getAccountMembershipsForUserQuery)).
			WithArgs(interfaceToDriverValue(getAccountMembershipsForUserArgs)...).
			WillReturnRows(buildMockRowsFromAccountUserMemberships(exampleAccount.Members...))

		expectedActiveAccountID := exampleAccount.Members[0].BelongsToAccount

		expected := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                exampleUser.ID,
				Reputation:            exampleUser.ServiceAccountStatus,
				ReputationExplanation: exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(exampleUser.ServiceRoles...),
			},
			AccountPermissions: exampleAccountPermissionsMap,
			ActiveAccountID:    expectedActiveAccountID,
		}

		actual, err := c.BuildSessionContextDataForUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual, "expected and actual RequestContextData do not match")
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.BuildSessionContextDataForUser(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error retrieving user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		examplePermsMap := map[string]*types.UserAccountMembershipInfo{}
		for _, membership := range exampleAccount.Members {
			examplePermsMap[membership.BelongsToAccount] = &types.UserAccountMembershipInfo{
				AccountName:  exampleAccount.Name,
				AccountID:    membership.BelongsToAccount,
				AccountRoles: membership.AccountRoles,
			}
		}

		c, db := buildTestClient(t)

		userRetrievalArgs := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(userRetrievalArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.BuildSessionContextDataForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error retrieving account memberships", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		examplePermsMap := map[string]*types.UserAccountMembershipInfo{}
		for _, membership := range exampleAccount.Members {
			examplePermsMap[membership.BelongsToAccount] = &types.UserAccountMembershipInfo{
				AccountName:  exampleAccount.Name,
				AccountID:    membership.BelongsToAccount,
				AccountRoles: membership.AccountRoles,
			}
		}

		exampleAccountPermissionsMap := map[string]authorization.AccountRolePermissionsChecker{}
		for _, membership := range exampleAccount.Members {
			exampleAccountPermissionsMap[membership.BelongsToAccount] = authorization.NewAccountRolePermissionChecker(membership.AccountRoles...)
		}

		c, db := buildTestClient(t)

		userRetrievalArgs := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(userRetrievalArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		getAccountMembershipsForUserArgs := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getAccountMembershipsForUserQuery)).
			WithArgs(interfaceToDriverValue(getAccountMembershipsForUserArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.BuildSessionContextDataForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error scanning account user memberships", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		examplePermsMap := map[string]*types.UserAccountMembershipInfo{}
		for _, membership := range exampleAccount.Members {
			examplePermsMap[membership.BelongsToAccount] = &types.UserAccountMembershipInfo{
				AccountName:  exampleAccount.Name,
				AccountID:    membership.BelongsToAccount,
				AccountRoles: membership.AccountRoles,
			}
		}

		exampleAccountPermissionsMap := map[string]authorization.AccountRolePermissionsChecker{}
		for _, membership := range exampleAccount.Members {
			exampleAccountPermissionsMap[membership.BelongsToAccount] = authorization.NewAccountRolePermissionChecker(membership.AccountRoles...)
		}

		c, db := buildTestClient(t)

		userRetrievalArgs := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(userRetrievalArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		getAccountMembershipsForUserArgs := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getAccountMembershipsForUserQuery)).
			WithArgs(interfaceToDriverValue(getAccountMembershipsForUserArgs)...).
			WillReturnRows(buildInvalidMockRowsFromAccountUserMemberships(exampleAccount.Members...))

		actual, err := c.BuildSessionContextDataForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetDefaultAccountIDForUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()
		expected := exampleAccountID

		c, db := buildTestClient(t)

		args := []interface{}{exampleUserID, true}

		db.ExpectQuery(formatQueryForSQLMock(getDefaultAccountIDForUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(exampleAccountID))

		actual, err := c.GetDefaultAccountIDForUser(ctx, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, db.ExpectationsWereMet())
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetDefaultAccountIDForUser(ctx, "")
		assert.Error(t, err)
		assert.Zero(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()

		c, db := buildTestClient(t)

		args := []interface{}{exampleUserID, true}

		db.ExpectQuery(formatQueryForSQLMock(getDefaultAccountIDForUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetDefaultAccountIDForUser(ctx, exampleUserID)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, db.ExpectationsWereMet())
	})
}

func TestQuerier_MarkAccountAsUserDefault(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, db := buildTestClient(t)

		markAccountAsUserDefaultArgs := []interface{}{
			exampleUserID,
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(markAccountAsUserDefaultQuery)).
			WithArgs(interfaceToDriverValue(markAccountAsUserDefaultArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		assert.NoError(t, c.MarkAccountAsUserDefault(ctx, exampleUserID, exampleAccountID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, "", exampleAccount.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()

		c, _ := buildTestClient(t)

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, ""))
	})

	T.Run("with error marking account as default", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, db := buildTestClient(t)

		markAccountAsUserDefaultArgs := []interface{}{
			exampleUserID,
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(markAccountAsUserDefaultQuery)).
			WithArgs(interfaceToDriverValue(markAccountAsUserDefaultArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUserID, exampleAccountID))
	})
}

func TestQuerier_UserIsMemberOfAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, db := buildTestClient(t)

		userIsMemberOfAccountArgs := []interface{}{
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectQuery(formatQueryForSQLMock(userIsMemberOfAccountQuery)).
			WithArgs(interfaceToDriverValue(userIsMemberOfAccountArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(true))

		actual, err := c.UserIsMemberOfAccount(ctx, exampleUserID, exampleAccountID)
		assert.True(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccountID := fakes.BuildFakeID()

		c, _ := buildTestClient(t)

		actual, err := c.UserIsMemberOfAccount(ctx, "", exampleAccountID)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()

		c, _ := buildTestClient(t)

		actual, err := c.UserIsMemberOfAccount(ctx, exampleUserID, "")
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with error performing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, db := buildTestClient(t)

		userIsMemberOfAccountArgs := []interface{}{
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectQuery(formatQueryForSQLMock(userIsMemberOfAccountQuery)).
			WithArgs(interfaceToDriverValue(userIsMemberOfAccountArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.UserIsMemberOfAccount(ctx, exampleUserID, exampleAccountID)
		assert.False(t, actual)
		assert.Error(t, err)
	})
}

func TestQuerier_ModifyUserPermissions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, db := buildTestClient(t)

		fakeArgs := []interface{}{
			strings.Join(exampleInput.NewRoles, accountMemberRolesSeparator),
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(modifyUserPermissionsQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		assert.NoError(t, c.ModifyUserPermissions(ctx, exampleAccountID, exampleUserID, exampleInput))
	})

	T.Run("with invalid account id", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, _ := buildTestClient(t)

		assert.Error(t, c.ModifyUserPermissions(ctx, "", exampleUserID, exampleInput))
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, _ := buildTestClient(t)

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccountID, exampleUserID, nil))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, db := buildTestClient(t)

		fakeArgs := []interface{}{
			strings.Join(exampleInput.NewRoles, accountMemberRolesSeparator),
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(modifyUserPermissionsQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccountID, exampleUserID, exampleInput))
	})
}

func TestQuerier_TransferAccountOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccountID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferArgs := []interface{}{
			exampleInput.NewOwner,
			exampleInput.CurrentOwner,
			exampleAccountID,
		}

		db.ExpectExec(formatQueryForSQLMock(transferAccountOwnershipQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		fakeAccountMembershipsTransferArgs := []interface{}{
			exampleInput.NewOwner,
			exampleAccountID,
			exampleInput.CurrentOwner,
		}

		db.ExpectExec(formatQueryForSQLMock(transferAccountMembershipQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		db.ExpectCommit()

		assert.NoError(t, c.TransferAccountOwnership(ctx, exampleAccountID, exampleInput))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, _ := buildTestClient(t)

		assert.Error(t, c.TransferAccountOwnership(ctx, "", exampleInput))
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, nil))
	})

	T.Run("with error starting transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleInput))
	})

	T.Run("with error writing account transfer", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferArgs := []interface{}{
			exampleInput.NewOwner,
			exampleInput.CurrentOwner,
			exampleAccount.ID,
		}

		db.ExpectExec(formatQueryForSQLMock(transferAccountOwnershipQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleInput))
	})

	T.Run("with error writing membership transfers", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccountID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferArgs := []interface{}{
			exampleInput.NewOwner,
			exampleInput.CurrentOwner,
			exampleAccountID,
		}

		db.ExpectExec(formatQueryForSQLMock(transferAccountOwnershipQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		fakeAccountMembershipsTransferArgs := []interface{}{
			exampleInput.NewOwner,
			exampleAccountID,
			exampleInput.CurrentOwner,
		}

		db.ExpectExec(formatQueryForSQLMock(transferAccountMembershipQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsTransferArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccountID, exampleInput))
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccountID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferArgs := []interface{}{
			exampleInput.NewOwner,
			exampleInput.CurrentOwner,
			exampleAccountID,
		}

		db.ExpectExec(formatQueryForSQLMock(transferAccountOwnershipQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		fakeAccountMembershipsTransferArgs := []interface{}{
			exampleInput.NewOwner,
			exampleAccountID,
			exampleInput.CurrentOwner,
		}

		db.ExpectExec(formatQueryForSQLMock(transferAccountMembershipQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccountID, exampleInput))
	})
}

func TestQuerier_AddUserToAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.BelongsToAccount = exampleAccount.ID

		exampleInput := &types.AddUserToAccountInput{
			Reason:       t.Name(),
			AccountID:    exampleAccount.ID,
			UserID:       exampleAccount.BelongsToUser,
			AccountRoles: []string{accountMemberRolesSeparator},
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		addUserToAccountArgs := []interface{}{
			exampleInput.ID,
			exampleInput.UserID,
			exampleInput.AccountID,
			strings.Join(exampleInput.AccountRoles, accountMemberRolesSeparator),
		}

		db.ExpectExec(formatQueryForSQLMock(addUserToAccountQuery)).
			WithArgs(interfaceToDriverValue(addUserToAccountArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountUserMembership.ID))

		assert.NoError(t, c.AddUserToAccount(ctx, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.AddUserToAccount(ctx, nil))
	})

	T.Run("with error writing add query", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.BelongsToAccount = exampleAccount.ID

		exampleInput := &types.AddUserToAccountInput{
			Reason:       t.Name(),
			AccountID:    exampleAccount.ID,
			UserID:       exampleAccount.BelongsToUser,
			AccountRoles: []string{accountMemberRolesSeparator},
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		addUserToAccountArgs := []interface{}{
			exampleInput.ID,
			exampleInput.UserID,
			exampleInput.AccountID,
			strings.Join(exampleInput.AccountRoles, accountMemberRolesSeparator),
		}

		db.ExpectExec(formatQueryForSQLMock(addUserToAccountQuery)).
			WithArgs(interfaceToDriverValue(addUserToAccountArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.AddUserToAccount(ctx, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_RemoveUserFromAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, db := buildTestClient(t)

		args := []interface{}{
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(removeUserFromAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountID))

		assert.NoError(t, c.RemoveUserFromAccount(ctx, exampleUserID, exampleAccountID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(ctx, "", exampleAccount.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()

		c, _ := buildTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, ""))
	})

	T.Run("with error writing removal to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, db := buildTestClient(t)

		args := []interface{}{
			exampleAccountID,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(removeUserFromAccountQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUserID, exampleAccountID))
	})
}

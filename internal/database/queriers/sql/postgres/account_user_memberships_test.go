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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func buildMockRowsFromAccountUserMemberships(memberships ...*types.AccountUserMembership) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(querybuilding.AccountsUserMembershipTableColumns)

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
	exampleRows := sqlmock.NewRows(querybuilding.AccountsUserMembershipTableColumns)

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

		fakeUserRetrievalQuery, fakeUserRetrievalArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeUserRetrievalQuery)).
			WithArgs(interfaceToDriverValue(fakeUserRetrievalArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		fakeAccountMembershipsQuery, fakeAccountMembershipsArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeAccountMembershipsQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsArgs)...).
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

		fakeUserRetrievalQuery, fakeUserRetrievalArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeUserRetrievalQuery)).
			WithArgs(interfaceToDriverValue(fakeUserRetrievalArgs)...).
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

		fakeUserRetrievalQuery, fakeUserRetrievalArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeUserRetrievalQuery)).
			WithArgs(interfaceToDriverValue(fakeUserRetrievalArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		fakeAccountMembershipsQuery, fakeAccountMembershipsArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeAccountMembershipsQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsArgs)...).
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

		fakeUserRetrievalQuery, fakeUserRetrievalArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeUserRetrievalQuery)).
			WithArgs(interfaceToDriverValue(fakeUserRetrievalArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		fakeAccountMembershipsQuery, fakeAccountMembershipsArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeAccountMembershipsQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsArgs)...).
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
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		expected := exampleAccount.ID

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(exampleAccount.ID))

		actual, err := c.GetDefaultAccountIDForUser(ctx, exampleUser.ID)
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
		exampleUser := fakes.BuildFakeUser()

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetDefaultAccountIDForUser(ctx, exampleUser.ID)
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
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.NoError(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, "", exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()

		c, _ := buildTestClient(t)

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, "", exampleUser.ID))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with error marking account as default", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID))
	})
}

func TestQuerier_UserIsMemberOfAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(true))

		actual, err := c.UserIsMemberOfAccount(ctx, exampleUser.ID, exampleAccount.ID)
		assert.True(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		actual, err := c.UserIsMemberOfAccount(ctx, "", exampleAccount.ID)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()

		c, _ := buildTestClient(t)

		actual, err := c.UserIsMemberOfAccount(ctx, exampleUser.ID, "")
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with error performing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.UserIsMemberOfAccount(ctx, exampleUser.ID, exampleAccount.ID)
		assert.False(t, actual)
		assert.Error(t, err)
	})
}

func TestQuerier_ModifyUserPermissions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.NoError(t, c.ModifyUserPermissions(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, _ := buildTestClient(t)

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccount.ID, "", exampleUser.ID, exampleInput))
	})

	T.Run("with invalid account id", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, _ := buildTestClient(t)

		assert.Error(t, c.ModifyUserPermissions(ctx, "", exampleUser.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID, nil))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccount.ID, exampleUser.ID, exampleUser.ID, exampleInput))
	})
}

func TestQuerier_TransferAccountOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferQuery, fakeAccountTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountMembershipsTransferQuery, fakeAccountMembershipsTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountMembershipsTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.NoError(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, _ := buildTestClient(t)

		assert.Error(t, c.TransferAccountOwnership(ctx, "", exampleUser.ID, exampleInput))
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleUser.ID, nil))
	})

	T.Run("with error starting transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with error writing account transfer", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferQuery, fakeAccountTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with error writing membership transfers", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferQuery, fakeAccountTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountMembershipsTransferQuery, fakeAccountMembershipsTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountMembershipsTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsTransferArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with error writing membership transfers audit log entry", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferQuery, fakeAccountTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountMembershipsTransferQuery, fakeAccountMembershipsTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountMembershipsTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleUser.ID, exampleInput))
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeAccountTransferQuery, fakeAccountTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		fakeAccountMembershipsTransferQuery, fakeAccountMembershipsTransferArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountMembershipsTransferQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountMembershipsTransferArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.TransferAccountOwnership(ctx, exampleAccount.ID, exampleUser.ID, exampleInput))
	})
}

func TestQuerier_AddUserToAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
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

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountUserMembership.ID))

		db.ExpectCommit()

		assert.NoError(t, c.AddUserToAccount(ctx, exampleInput, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
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
		c, _ := buildTestClient(t)

		assert.Error(t, c.AddUserToAccount(ctx, exampleInput, ""))
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.BelongsToAccount = exampleAccount.ID

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.AddUserToAccount(ctx, nil, exampleUser.ID))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
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

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.AddUserToAccount(ctx, exampleInput, exampleUser.ID))
	})

	T.Run("with error writing add query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
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

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.AddUserToAccount(ctx, exampleInput, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
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

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountUserMembership.ID))

		db.ExpectRollback()

		assert.Error(t, c.AddUserToAccount(ctx, exampleInput, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
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

		db.ExpectBegin()

		fakeUpdateQuery, fakeUpdateArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUpdateQuery)).
			WithArgs(interfaceToDriverValue(fakeUpdateArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccountUserMembership.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.AddUserToAccount(ctx, exampleInput, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_RemoveUserFromAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.NoError(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID, t.Name()))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(ctx, "", exampleAccount.ID, exampleUser.ID, t.Name()))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()

		c, _ := buildTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, "", exampleUser.ID, t.Name()))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, _ := buildTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, exampleAccount.ID, "", t.Name()))
	})

	T.Run("with empty reason", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID, t.Name()))
	})

	T.Run("with error writing removal to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID, t.Name()))
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectRollback()

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID, t.Name()))
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, exampleAccount.ID, exampleUser.ID, t.Name()))
	})
}

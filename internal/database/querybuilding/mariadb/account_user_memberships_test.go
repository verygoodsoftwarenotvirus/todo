package mariadb

import (
	"context"
	"math"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
)

func TestSqlite_BuildGetDefaultAccountIDForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT accounts.id FROM accounts JOIN account_user_memberships ON account_user_memberships.belongs_to_account = accounts.id WHERE account_user_memberships.belongs_to_user = ? AND account_user_memberships.default_account = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
			true,
		}
		actualQuery, actualArgs := q.BuildGetDefaultAccountIDForUserQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildUserIsMemberOfAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "SELECT EXISTS ( SELECT account_user_memberships.id FROM account_user_memberships WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_user = ? )"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUserIsMemberOfAccountQuery(ctx, exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildAddUserToAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := &types.AddUserToAccountInput{
			UserID:       exampleUser.ID,
			AccountID:    exampleAccount.ID,
			Reason:       t.Name(),
			AccountRoles: []string{authorization.AccountMemberRole.String()},
		}

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account,account_role,user_account_permissions) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			exampleInput.UserID,
			exampleAccount.ID,
			exampleInput.AccountRoles,
			exampleInput.UserAccountPermissions,
		}
		actualQuery, actualArgs := q.BuildAddUserToAccountQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildRemoveUserFromAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "DELETE FROM account_user_memberships WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_account = ? AND account_user_memberships.belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildRemoveUserFromAccountQuery(ctx, exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildArchiveAccountMembershipsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE account_user_memberships SET archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountMembershipsForUserQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildCreateMembershipForNewUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account,default_account,account_role,user_account_permissions) VALUES (?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
			true,
			authorization.AccountAdminRole.String(),
			math.MaxInt64,
		}
		actualQuery, actualArgs := q.BuildCreateMembershipForNewUserQuery(ctx, exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetAccountMembershipsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT account_user_memberships.id, account_user_memberships.belongs_to_user, account_user_memberships.belongs_to_account, account_user_memberships.account_role, account_user_memberships.user_account_permissions, account_user_memberships.default_account, account_user_memberships.created_on, account_user_memberships.last_updated_on, account_user_memberships.archived_on, accounts.name FROM account_user_memberships JOIN accounts ON accounts.id = account_user_memberships.belongs_to_account WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountMembershipsForUserQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildMarkAccountAsUserDefaultQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET default_account = (belongs_to_user = ? AND belongs_to_account = ?) WHERE archived_on IS NULL AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildMarkAccountAsUserDefaultQuery(ctx, exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildModifyUserPermissionsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleRole := authorization.AccountMemberRole.String()
		exampleAccount := fakes.BuildFakeAccount()
		examplePermissions := testutil.BuildMaxUserPerms()

		expectedQuery := "UPDATE account_user_memberships SET user_account_permissions = ?, account_role = ? WHERE belongs_to_account = ? AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			examplePermissions,
			exampleRole,
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildModifyUserPermissionsQuery(ctx, exampleUser.ID, exampleAccount.ID, examplePermissions, exampleRole)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildTransferAccountOwnershipQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleOldOwner := fakes.BuildFakeUser()
		exampleNewOwner := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE accounts SET belongs_to_user = ? WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleNewOwner.ID,
			exampleOldOwner.ID,
			exampleAccount.ID,
		}
		actualQuery, actualArgs := q.BuildTransferAccountOwnershipQuery(ctx, exampleOldOwner.ID, exampleNewOwner.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildTransferAccountMembershipsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleOldOwner := fakes.BuildFakeUser()
		exampleNewOwner := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET belongs_to_user = ? WHERE archived_on IS NULL AND belongs_to_account = ? AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleNewOwner.ID,
			exampleAccount.ID,
			exampleOldOwner.ID,
		}
		actualQuery, actualArgs := q.BuildTransferAccountMembershipsQuery(ctx, exampleOldOwner.ID, exampleNewOwner.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

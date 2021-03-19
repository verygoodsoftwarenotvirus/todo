package sqlite

import (
	"math"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
)

func TestSqlite_BuildMarkAccountAsUserPrimaryQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET default_account = (belongs_to_user = ? AND belongs_to_account = ?) WHERE archived_on IS NULL AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildMarkAccountAsUserPrimaryQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildUserIsMemberOfAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "SELECT EXISTS ( SELECT account_user_memberships.id FROM account_user_memberships WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_user = ? )"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUserIsMemberOfAccountQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildAddUserToAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := &types.AddUserToAccountInput{
			UserID: exampleUser.ID,
		}

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account,user_account_permissions) VALUES (?,?,?)"
		expectedArgs := []interface{}{
			exampleInput.UserID,
			exampleAccount.ID,
			exampleInput.UserAccountPermissions,
		}
		actualQuery, actualArgs := q.BuildAddUserToAccountQuery(exampleAccount.ID, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildRemoveUserFromAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "DELETE FROM account_user_memberships WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_account = ? AND account_user_memberships.belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildRemoveUserFromAccountQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildArchiveAccountMembershipsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE account_user_memberships SET archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountMembershipsForUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildCreateMembershipForNewUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account,default_account,user_account_permissions) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
			true,
			math.MaxUint32,
		}
		actualQuery, actualArgs := q.BuildCreateMembershipForNewUserQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetAccountMembershipsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT account_user_memberships.id, account_user_memberships.belongs_to_user, account_user_memberships.belongs_to_account, account_user_memberships.user_account_permissions, account_user_memberships.default_account, account_user_memberships.created_on, account_user_memberships.archived_on FROM account_user_memberships WHERE archived_on IS NULL AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountMembershipsForUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildMarkAccountAsUserDefaultQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET default_account = (belongs_to_user = ? AND belongs_to_account = ?) WHERE archived_on IS NULL AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildMarkAccountAsUserDefaultQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildModifyUserPermissionsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		examplePermissions := testutil.BuildMaxUserPerms()

		expectedQuery := "UPDATE account_user_memberships SET user_account_permissions = ? WHERE belongs_to_account = ? AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			examplePermissions,
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildModifyUserPermissionsQuery(exampleUser.ID, exampleAccount.ID, examplePermissions)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildTransferAccountOwnershipQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleOldOwner := fakes.BuildFakeUser()
		exampleNewOwner := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE accounts SET belongs_to_user = ? WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleNewOwner.ID,
			exampleOldOwner.ID,
			exampleAccount.ID,
		}
		actualQuery, actualArgs := q.BuildTransferAccountOwnershipQuery(exampleOldOwner.ID, exampleNewOwner.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildTransferAccountMembershipsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleOldOwner := fakes.BuildFakeUser()
		exampleNewOwner := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET belongs_to_user = ? WHERE archived_on IS NULL AND belongs_to_account = ? AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			exampleNewOwner.ID,
			exampleAccount.ID,
			exampleOldOwner.ID,
		}
		actualQuery, actualArgs := q.BuildTransferAccountMembershipsQuery(exampleOldOwner.ID, exampleNewOwner.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

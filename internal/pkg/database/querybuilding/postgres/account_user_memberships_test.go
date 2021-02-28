package postgres

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildArchiveAccountMembershipsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE account_user_memberships SET archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountMembershipsForUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAccountMembershipsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT account_user_memberships.id, account_user_memberships.belongs_to_user, account_user_memberships.belongs_to_account, account_user_memberships.user_account_permissions, account_user_memberships.default_account, account_user_memberships.created_on, account_user_memberships.archived_on FROM account_user_memberships WHERE archived_on IS NULL AND belongs_to_user = $1"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountMembershipsForUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildMarkAccountAsUserDefaultQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET default_account = (belongs_to_user = $1 AND belongs_to_account = $2) WHERE archived_on IS NULL AND belongs_to_user = $3"
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

func TestPostgres_BuildUserIsMemberOfAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "SELECT EXISTS ( SELECT account_user_memberships.id FROM account_user_memberships WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_user = $1 )"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUserIsMemberOfAccountQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildAddUserToAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account) VALUES ($1,$2)"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
		}
		actualQuery, actualArgs := q.BuildAddUserToAccountQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildRemoveUserFromAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "DELETE FROM account_user_memberships WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_account = $1 AND account_user_memberships.belongs_to_user = $2"
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

func TestPostgres_BuildCreateMembershipForNewUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account,default_account) VALUES ($1,$2,$3)"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
			true,
		}
		actualQuery, actualArgs := q.BuildCreateMembershipForNewUserQuery(exampleUser.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildModifyUserPermissionsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		examplePermissions := testutil.BuildMaxUserPerms()

		expectedQuery := "UPDATE account_user_memberships SET user_account_permissions = $1 WHERE belongs_to_account = $2 AND belongs_to_user = $3"
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

func TestPostgres_BuildTransferAccountOwnershipQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET belongs_to_user = $1 WHERE archived_on IS NULL AND belongs_to_account = $2 AND belongs_to_user = $3"
		expectedArgs := []interface{}{
			exampleAccount.BelongsToUser,
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildTransferAccountOwnershipQuery(exampleUser.ID, exampleAccount.BelongsToUser, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

package mariadb

import (
	"context"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestMariaDB_BuildGetDefaultAccountIDForUserQuery(T *testing.T) {
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

		expectedQuery := "SELECT EXISTS ( SELECT account_user_memberships.id FROM account_user_memberships WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_account = ? AND account_user_memberships.belongs_to_user = ? )"
		expectedArgs := []interface{}{
			exampleAccount.ID,
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

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account,account_roles) VALUES (?,?,?)"
		expectedArgs := []interface{}{
			exampleInput.UserID,
			exampleAccount.ID,
			strings.Join(exampleInput.AccountRoles, accountMemberRolesSeparator),
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

		expectedQuery := "INSERT INTO account_user_memberships (belongs_to_user,belongs_to_account,default_account,account_roles) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
			true,
			authorization.AccountAdminRole.String(),
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

		expectedQuery := "SELECT account_user_memberships.id, account_user_memberships.belongs_to_user, account_user_memberships.belongs_to_account, account_user_memberships.account_roles, account_user_memberships.default_account, account_user_memberships.created_on, account_user_memberships.last_updated_on, account_user_memberships.archived_on FROM account_user_memberships JOIN accounts ON accounts.id = account_user_memberships.belongs_to_account WHERE account_user_memberships.archived_on IS NULL AND account_user_memberships.belongs_to_user = ?"
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
		exampleRoles := []string{authorization.AccountMemberRole.String()}
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET account_roles = ? WHERE belongs_to_account = ? AND belongs_to_user = ?"
		expectedArgs := []interface{}{
			strings.Join(exampleRoles, accountMemberRolesSeparator),
			exampleAccount.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildModifyUserPermissionsQuery(ctx, exampleUser.ID, exampleAccount.ID, exampleRoles)

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

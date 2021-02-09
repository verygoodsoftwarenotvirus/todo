package postgres

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildMarkAccountAsUserDefaultQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "UPDATE account_user_memberships SET is_primary_user_account = (belongs_to_user = $1 AND belongs_to_account = $2) WHERE belongs_to_user = $3 AND archived_on IS NOT NULL"
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

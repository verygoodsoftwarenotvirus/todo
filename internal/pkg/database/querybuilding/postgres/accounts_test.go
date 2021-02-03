package postgres

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildGetAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		expectedQuery := "SELECT accounts.id, accounts.external_id, accounts.name, accounts.plan_id, accounts.is_personal_account, accounts.created_on, accounts.last_updated_on, accounts.archived_on, accounts.belongs_to_user FROM accounts WHERE accounts.archived_on IS NULL AND accounts.belongs_to_user = $1 AND accounts.id = $2"
		expectedArgs := []interface{}{
			exampleAccount.BelongsToUser,
			exampleAccount.ID,
		}
		actualQuery, actualArgs := q.BuildGetAccountQuery(exampleAccount.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAllAccountsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(accounts.id) FROM accounts WHERE accounts.archived_on IS NULL"
		actualQuery := q.BuildGetAllAccountsCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_BuildGetBatchOfAccountsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT accounts.id, accounts.external_id, accounts.name, accounts.plan_id, accounts.is_personal_account, accounts.created_on, accounts.last_updated_on, accounts.archived_on, accounts.belongs_to_user FROM accounts WHERE accounts.id > $1 AND accounts.id < $2"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfAccountsQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAccountsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT accounts.id, accounts.external_id, accounts.name, accounts.plan_id, accounts.is_personal_account, accounts.created_on, accounts.last_updated_on, accounts.archived_on, accounts.belongs_to_user, (SELECT COUNT(accounts.id) FROM accounts WHERE accounts.archived_on IS NULL AND accounts.belongs_to_user = $1) as total_count, (SELECT COUNT(accounts.id) FROM accounts WHERE accounts.archived_on IS NULL AND accounts.belongs_to_user = $2 AND accounts.created_on > $3 AND accounts.created_on < $4 AND accounts.last_updated_on > $5 AND accounts.last_updated_on < $6) as filtered_count FROM accounts WHERE accounts.archived_on IS NULL AND accounts.belongs_to_user = $7 AND accounts.created_on > $8 AND accounts.created_on < $9 AND accounts.last_updated_on > $10 AND accounts.last_updated_on < $11 GROUP BY accounts.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			exampleUser.ID,
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.BuildGetAccountsQuery(exampleUser.ID, false, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildCreateAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleAccount.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO accounts (external_id,name,belongs_to_user) VALUES ($1,$2,$3) RETURNING id"
		expectedArgs := []interface{}{
			exampleAccount.ExternalID,
			exampleAccount.Name,
			exampleAccount.BelongsToUser,
		}
		actualQuery, actualArgs := q.BuildCreateAccountQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestPostgres_BuildUpdateAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE accounts SET name = $1, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $2 AND id = $3"
		expectedArgs := []interface{}{
			exampleAccount.Name,
			exampleAccount.BelongsToUser,
			exampleAccount.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateAccountQuery(exampleAccount)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildArchiveAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE accounts SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1 AND id = $2"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleAccount.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAccountQuery(exampleAccount.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAuditLogEntriesForAccountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAccount := fakes.BuildFakeAccount()

		expectedQuery := "SELECT audit_log.id, audit_log.external_id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.context->'account_id' = $1 ORDER BY audit_log.created_on"
		expectedArgs := []interface{}{
			exampleAccount.ID,
		}
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForAccountQuery(exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

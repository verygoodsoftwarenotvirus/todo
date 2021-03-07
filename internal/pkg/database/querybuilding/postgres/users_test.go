package postgres

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildUserIsBannedQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT EXISTS ( SELECT users.id FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND (users.reputation = $2 OR users.reputation = $3) )"
		expectedArgs := []interface{}{
			exampleUser.ID,
			types.BannedAccountStatus,
			types.TerminatedAccountStatus,
		}
		actualQuery, actualArgs := q.BuildUserIsBannedQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUserWithUnverifiedTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on, (SELECT COUNT(*) FROM users WHERE users.archived_on IS NULL AND items.created_on > $1 AND items.created_on < $2 AND items.last_updated_on > $3 AND items.last_updated_on < $4) FROM users WHERE users.archived_on IS NULL AND users.created_on > $5 AND users.created_on < $6 AND users.last_updated_on > $7 AND users.last_updated_on < $8 ORDER BY users.created_on LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.BuildGetUsersQuery(filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}
		actualQuery, actualArgs := q.BuildGetUserByUsernameQuery(exampleUser.Username)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildSearchForUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUsername := fakes.BuildFakeUser().Username

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.username ILIKE $1 AND users.archived_on IS NULL AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			fmt.Sprintf("%s%%", exampleUsername),
		}
		actualQuery, actualArgs := q.BuildSearchForUserByUsernameQuery(exampleUsername)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAllUsersCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"
		actualQuery := q.BuildGetAllUsersCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_BuildTestUserCreationQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildTestUserCreationConfig()
		exampleInput.IsServiceAdmin = true

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleUser.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO users (external_id,username,hashed_password,salt,two_factor_secret,reputation,site_admin_permissions,two_factor_secret_verified_on) VALUES ($1,$2,$3,$4,$5,$6,$7,extract(epoch FROM NOW())) RETURNING id"
		expectedArgs := []interface{}{
			exampleUser.ExternalID,
			exampleInput.Username,
			exampleInput.HashedPassword,
			[]byte("aaaaaaaaaaaaaaaa"),
			"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			types.GoodStandingAccountStatus,
			math.MaxUint32,
		}
		actualQuery, actualArgs := q.BuildTestUserCreationQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestPostgres_BuildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleUser.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO users (external_id,username,hashed_password,salt,two_factor_secret,reputation,site_admin_permissions) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id"
		expectedArgs := []interface{}{
			exampleUser.ExternalID,
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleUser.TwoFactorSecret,
			types.UnverifiedAccountStatus,
			0,
		}
		actualQuery, actualArgs := q.BuildCreateUserQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestPostgres_BuildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, salt = $3, two_factor_secret = $4, two_factor_secret_verified_on = $5, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $6"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateUserQuery(exampleUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = $1, requires_password_change = $2, password_last_changed_on = extract(epoch FROM NOW()), last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $3"
		expectedArgs := []interface{}{
			exampleUser.HashedPassword,
			false,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildUpdateUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = $1, two_factor_secret = $2 WHERE archived_on IS NULL AND id = $3"
		expectedArgs := []interface{}{
			nil,
			exampleUser.TwoFactorSecret,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateUserTwoFactorSecretQuery(exampleUser.ID, exampleUser.TwoFactorSecret)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = extract(epoch FROM NOW()), reputation = $1 WHERE archived_on IS NULL AND id = $2"
		expectedArgs := []interface{}{
			types.GoodStandingAccountStatus,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildSetUserStatusQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := types.UserReputationUpdateInput{
			TargetAccountID: exampleUser.ID,
			NewReputation:   "new",
			Reason:          "test",
		}

		expectedQuery := "UPDATE users SET reputation = $1, reputation_explanation = $2 WHERE archived_on IS NULL AND id = $3"
		expectedArgs := []interface{}{
			exampleInput.NewReputation,
			exampleInput.Reason,
			exampleInput.TargetAccountID,
		}
		actualQuery, actualArgs := q.BuildSetUserStatusQuery(exampleUser.ID, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $1"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAuditLogEntriesForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT audit_log.id, audit_log.external_id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE (audit_log.context->'user_id' = $1 OR audit_log.context->'performed_by' = $2) ORDER BY audit_log.created_on"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

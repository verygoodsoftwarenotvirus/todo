package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestSqlite_BuildUserIsBannedQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT EXISTS ( SELECT users.id FROM users WHERE users.archived_on IS NULL AND users.id = ? AND (users.reputation = ? OR users.reputation = ?) )"
		expectedArgs := []interface{}{
			exampleUser.ID,
			string(types.BannedUserReputation),
			string(types.TerminatedUserReputation),
		}
		actualQuery, actualArgs := q.BuildUserHasStatusQuery(ctx, exampleUser.ID, string(types.BannedUserReputation), string(types.TerminatedUserReputation))

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = ? AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetUserQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetUserWithUnverifiedTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = ? AND users.two_factor_secret_verified_on IS NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on, (SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL) as total_count, (SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL AND users.created_on > ? AND users.created_on < ? AND users.last_updated_on > ? AND users.last_updated_on < ?) as filtered_count FROM users WHERE users.archived_on IS NULL AND users.created_on > ? AND users.created_on < ? AND users.last_updated_on > ? AND users.last_updated_on < ? GROUP BY users.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.BuildGetUsersQuery(ctx, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.external_id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = ? AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}
		actualQuery, actualArgs := q.BuildGetUserByUsernameQuery(ctx, exampleUser.Username)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetAllUsersCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"
		actualQuery := q.BuildGetAllUsersCountQuery(ctx)

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_BuildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On(
			"NewExternalID").Return(exampleUser.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO users (external_id,username,hashed_password,two_factor_secret,reputation,site_admin_permissions) VALUES (?,?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleUser.ExternalID,
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			types.UnverifiedAccountStatus,
			0,
		}
		actualQuery, actualArgs := q.BuildCreateUserQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestSqlite_BuildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, avatar_src = ?, two_factor_secret = ?, two_factor_secret_verified_on = ?, last_updated_on = (strftime('%s','now')) WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.AvatarSrc,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateUserQuery(ctx, exampleUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = ?, requires_password_change = ?, password_last_changed_on = (strftime('%s','now')), last_updated_on = (strftime('%s','now')) WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			exampleUser.HashedPassword,
			false,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateUserPasswordQuery(ctx, exampleUser.ID, exampleUser.HashedPassword)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildUpdateUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = ?, two_factor_secret = ? WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			nil,
			exampleUser.TwoFactorSecret,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateUserTwoFactorSecretQuery(ctx, exampleUser.ID, exampleUser.TwoFactorSecret)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = (strftime('%s','now')), reputation = ? WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			types.GoodStandingAccountStatus,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildVerifyUserTwoFactorSecretQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveUserQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

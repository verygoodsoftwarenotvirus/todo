package postgres

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildUserHasStatusQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleStatuses := []string{
			string(types.BannedUserAccountStatus),
			string(types.TerminatedUserReputation),
		}

		expectedQuery := "SELECT EXISTS ( SELECT users.id FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND (users.reputation = $2 OR users.reputation = $3) )"
		expectedArgs := []interface{}{
			exampleUser.ID,
			string(types.BannedUserAccountStatus),
			string(types.TerminatedUserReputation),
		}
		actualQuery, actualArgs := q.BuildUserHasStatusQuery(ctx, exampleUser.ID, exampleStatuses...)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetUserQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUserWithUnverifiedTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on, (SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL) as total_count, (SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL AND users.created_on > $1 AND users.created_on < $2 AND users.last_updated_on > $3 AND users.last_updated_on < $4) as filtered_count FROM users WHERE users.archived_on IS NULL AND users.created_on > $5 AND users.created_on < $6 AND users.last_updated_on > $7 AND users.last_updated_on < $8 GROUP BY users.id LIMIT 20 OFFSET 180"
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

func TestPostgres_BuildTestUserCreationQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleInput := &types.TestUserCreationConfig{
			ID:             fakes.BuildFakeID(),
			Username:       "username",
			Password:       "password",
			HashedPassword: "hashashashash",
			IsServiceAdmin: true,
		}

		expectedQuery := "INSERT INTO users (id,username,hashed_password,two_factor_secret,reputation,service_roles,two_factor_secret_verified_on) VALUES ($1,$2,$3,$4,$5,$6,extract(epoch FROM NOW()))"
		expectedArgs := []interface{}{
			exampleInput.ID,
			exampleInput.Username,
			exampleInput.HashedPassword,
			querybuilding.DefaultTestUserTwoFactorSecret,
			types.GoodStandingAccountStatus,
			authorization.ServiceAdminRole.String(),
		}
		actualQuery, actualArgs := q.BuildTestUserCreationQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}
		actualQuery, actualArgs := q.BuildGetUserByUsernameQuery(ctx, exampleUser.Username)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildSearchForUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.username ILIKE $1 AND users.archived_on IS NULL AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			fmt.Sprintf("%s%%", exampleUser.Username),
		}
		actualQuery, actualArgs := q.BuildSearchForUserByUsernameQuery(ctx, exampleUser.Username)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAllUsersCountQuery(T *testing.T) {
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

func TestPostgres_BuildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		expectedQuery := "INSERT INTO users (id,username,hashed_password,two_factor_secret,reputation,service_roles) VALUES ($1,$2,$3,$4,$5,$6)"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			types.UnverifiedAccountStatus,
			authorization.ServiceUserRole.String(),
		}
		actualQuery, actualArgs := q.BuildUserCreationQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, avatar_src = $3, two_factor_secret = $4, two_factor_secret_verified_on = $5, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $6"
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

func TestPostgres_BuildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = $1, requires_password_change = $2, password_last_changed_on = extract(epoch FROM NOW()), last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $3"
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

func TestPostgres_BuildSetUserStatusQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserReputationUpdateInputFromUser(exampleUser)

		expectedQuery := "UPDATE users SET reputation = $1, reputation_explanation = $2 WHERE archived_on IS NULL AND id = $3"
		expectedArgs := []interface{}{
			exampleInput.NewReputation,
			exampleInput.Reason,
			exampleInput.TargetUserID,
		}
		actualQuery, actualArgs := q.BuildSetUserStatusQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildUpdateUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = $1, two_factor_secret = $2 WHERE archived_on IS NULL AND id = $3"
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

func TestPostgres_BuildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = extract(epoch FROM NOW()), reputation = $1 WHERE archived_on IS NULL AND id = $2"
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

func TestPostgres_BuildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND id = $1"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveUserQuery(ctx, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

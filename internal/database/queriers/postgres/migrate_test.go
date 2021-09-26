package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func TestQuerier_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)

		exampleTestUserConfig := &types.TestUserCreationConfig{
			Username:       exampleUser.Username,
			Password:       exampleUser.HashedPassword,
			HashedPassword: exampleUser.HashedPassword,
			IsServiceAdmin: true,
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		// called by c.IsReady()
		db.ExpectPing()

		c.migrateOnce.Do(func() {})

		// expect TestUser to be queried for
		testUserExistenceArgs := []interface{}{exampleTestUserConfig.Username}
		db.ExpectQuery(formatQueryForSQLMock(testUserExistenceQuery)).
			WithArgs(interfaceToDriverValue(testUserExistenceArgs)...).
			WillReturnError(sql.ErrNoRows)

		db.ExpectBegin()

		// expect TestUser to be created
		testUserCreationArgs := []interface{}{
			&idMatcher{},
			exampleTestUserConfig.Username,
			exampleTestUserConfig.HashedPassword,
			defaultTestUserTwoFactorSecret,
			types.GoodStandingAccountStatus,
			authorization.ServiceAdminRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(testUserCreationQuery)).
			WithArgs(interfaceToDriverValue(testUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create account for created TestUser
		accountCreationInput := types.AccountCreationInputForNewUser(exampleUser)
		accountCreationArgs := []interface{}{
			&idMatcher{},
			accountCreationInput.Name,
			types.UnpaidAccountBillingStatus,
			accountCreationInput.ContactEmail,
			accountCreationInput.ContactPhone,
			&idMatcher{},
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		// create account user membership for created user

		createAccountMembershipForNewUserArgs := []interface{}{
			&idMatcher{},
			&idMatcher{},
			&idMatcher{},
			true,
			authorization.AccountAdminRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(createAccountMembershipForNewUserQuery)).
			WithArgs(interfaceToDriverValue(createAccountMembershipForNewUserArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		err := c.Migrate(ctx, 1, exampleTestUserConfig)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with failure executing creation query", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleTestUserConfig := &types.TestUserCreationConfig{
			Username:       exampleUser.Username,
			Password:       exampleUser.HashedPassword,
			HashedPassword: exampleUser.HashedPassword,
			IsServiceAdmin: true,
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		// called by c.IsReady()
		db.ExpectPing()

		c.migrateOnce.Do(func() {})

		// expect TestUser to be queried for
		testUserExistenceArgs := []interface{}{exampleTestUserConfig.Username}
		db.ExpectQuery(formatQueryForSQLMock(testUserExistenceQuery)).
			WithArgs(interfaceToDriverValue(testUserExistenceArgs)...).
			WillReturnError(sql.ErrNoRows)

		db.ExpectBegin()

		// expect TestUser to be created
		testUserCreationArgs := []interface{}{
			&idMatcher{},
			exampleTestUserConfig.Username,
			exampleTestUserConfig.HashedPassword,
			defaultTestUserTwoFactorSecret,
			types.GoodStandingAccountStatus,
			authorization.ServiceAdminRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(testUserCreationQuery)).
			WithArgs(interfaceToDriverValue(testUserCreationArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		err := c.Migrate(ctx, 1, exampleTestUserConfig)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db)
	})
}

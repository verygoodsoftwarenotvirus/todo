package queriers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		exampleAccountCreationInput := &types.AccountCreationInput{
			Name:          fmt.Sprintf("%s_default", exampleUser.Username),
			BelongsToUser: exampleUser.ID,
		}

		exampleInput := &types.TestUserCreationConfig{
			Username:       exampleUser.Username,
			Password:       exampleUser.HashedPassword,
			HashedPassword: exampleUser.HashedPassword,
			IsServiceAdmin: true,
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		// called by c.IsReady()
		db.ExpectPing()

		migrationFuncCalled := false

		// expect BuildMigrationFunc to be called
		mockQueryBuilder.On(
			"BuildMigrationFunc",
			mock.IsType(&sql.DB{})).
			Return(func() {
				migrationFuncCalled = true
			})

		// expect TestUser to be queried for
		fakeTestUserExistenceQuery, fakeTestUserExistenceArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.UserSQLQueryBuilder.On(
			"BuildGetUserByUsernameQuery",
			testutils.ContextMatcher,
			exampleInput.Username,
		).Return(fakeTestUserExistenceQuery, fakeTestUserExistenceArgs)

		db.ExpectQuery(formatQueryForSQLMock(fakeTestUserExistenceQuery)).
			WithArgs(interfaceToDriverValue(fakeTestUserExistenceArgs)...).
			WillReturnError(sql.ErrNoRows)

		db.ExpectBegin()

		// expect TestUser to be created
		fakeTestUserCreationQuery, fakeTestUserCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.On(
			"BuildTestUserCreationQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeTestUserCreationQuery, fakeTestUserCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeTestUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeTestUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created TestUser
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.MatchedBy(testutils.AuditLogEntryCreationInputMatcher(audit.UserCreationEvent))).
			Return(firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs)

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created TestUser
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On(
			"BuildAccountCreationQuery",
			testutils.ContextMatcher,
			mock.MatchedBy(accountCreationInputMatcher(t, exampleAccountCreationInput)),
		).Return(fakeAccountCreationQuery, fakeAccountCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.MatchedBy(testutils.AuditLogEntryCreationInputMatcher(audit.AccountCreationEvent))).
			Return(secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs)

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account user membership for created user
		fakeMembershipCreationQuery, fakeMembershipCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.On(
			"BuildCreateMembershipForNewUserQuery",
			testutils.ContextMatcher,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
		).Return(fakeMembershipCreationQuery, fakeMembershipCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeMembershipCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeMembershipCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On(
			"BuildCreateAuditLogEntryQuery",
			testutils.ContextMatcher,
			mock.MatchedBy(testutils.AuditLogEntryCreationInputMatcher(audit.UserAddedToAccountEvent))).
			Return(thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs)

		db.ExpectExec(formatQueryForSQLMock(thirdFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(thirdFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.Migrate(ctx, 1, exampleInput)
		assert.NoError(t, err)
		assert.True(t, migrationFuncCalled)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with failure executing creation query", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleInput := &types.TestUserCreationConfig{
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

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		// expect BuildMigrationFunc to be called
		mockQueryBuilder.On(
			"BuildMigrationFunc",
			mock.IsType(&sql.DB{})).
			Return(func() {})

		// expect TestUser to be queried for
		fakeTestUserExistenceQuery, fakeTestUserExistenceArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.UserSQLQueryBuilder.On(
			"BuildGetUserByUsernameQuery",
			testutils.ContextMatcher,
			exampleInput.Username,
		).Return(fakeTestUserExistenceQuery, fakeTestUserExistenceArgs)

		db.ExpectQuery(formatQueryForSQLMock(fakeTestUserExistenceQuery)).
			WithArgs(interfaceToDriverValue(fakeTestUserExistenceArgs)...).
			WillReturnError(sql.ErrNoRows)

		// expect TestUser to be created
		fakeTestUserCreationQuery, fakeTestUserCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.On(
			"BuildTestUserCreationQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeTestUserCreationQuery, fakeTestUserCreationArgs)

		c.sqlQueryBuilder = mockQueryBuilder

		// expect transaction begin
		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.Migrate(ctx, 1, exampleInput))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

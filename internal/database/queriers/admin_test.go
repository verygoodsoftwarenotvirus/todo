package queriers

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQuerier_UpdateUserReputation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := &types.UserReputationUpdateInput{
			TargetUserID:  exampleUser.ID,
			NewReputation: "new",
			Reason:        "because",
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.UserSQLQueryBuilder.On(
			"BuildSetUserStatusQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		assert.NoError(t, c.UpdateUserReputation(ctx, exampleUser.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := &types.UserReputationUpdateInput{
			TargetUserID:  exampleUser.ID,
			NewReputation: "new",
			Reason:        "because",
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.UserSQLQueryBuilder.On(
			"BuildSetUserStatusQuery",
			testutils.ContextMatcher,
			exampleInput,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserReputation(ctx, exampleUser.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_LogUserBanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleServiceAdmin := fakes.BuildFakeUser()
		exampleUser := fakes.BuildFakeUser()
		exampleReason := "smells bad"
		exampleAuditLogEntry := audit.BuildUserBanEventEntry(exampleServiceAdmin.ID, exampleUser.ID, exampleReason)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogUserBanEvent(ctx, exampleServiceAdmin.ID, exampleUser.ID, exampleReason)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_LogCycleCookieSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntry := audit.BuildCycleCookieSecretEvent(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogCycleCookieSecretEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_LogSuccessfulLoginEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntry := audit.BuildSuccessfulLoginEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogSuccessfulLoginEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_LogBannedUserLoginAttemptEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntry := audit.BuildBannedUserLoginAttemptEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogBannedUserLoginAttemptEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_LogUnsuccessfulLoginBadPasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntry := audit.BuildUnsuccessfulLoginBadPasswordEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogUnsuccessfulLoginBadPasswordEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_LogUnsuccessfulLoginBad2FATokenEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntry := audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogUnsuccessfulLoginBad2FATokenEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_LogLogoutEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntry := audit.BuildLogoutEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogLogoutEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

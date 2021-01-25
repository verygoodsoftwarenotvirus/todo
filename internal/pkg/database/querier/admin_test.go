package querier

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func TestClient_UpdateUserAccountStatus(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := types.AccountStatusUpdateInput{
			TargetAccountID: exampleUser.ID,
			NewStatus:       "new",
			Reason:          "because",
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.UserSQLQueryBuilder.
			On("BuildSetUserStatusQuery", exampleUser.ID, exampleInput).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleUser.ID))

		err := c.UpdateUserAccountStatus(ctx, exampleUser.ID, exampleInput)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_LogUserBanEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleSiteAdmin := fakes.BuildFakeUser()
		exampleUser := fakes.BuildFakeUser()
		exampleReason := "smells bad"
		exampleAuditLogEntry := audit.BuildUserBanEventEntry(exampleSiteAdmin.ID, exampleUser.ID, exampleReason)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogUserBanEvent(ctx, exampleSiteAdmin.ID, exampleUser.ID, exampleReason)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_LogAccountTerminationEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleSiteAdmin := fakes.BuildFakeUser()
		exampleUser := fakes.BuildFakeUser()
		exampleReason := "smells bad"
		exampleAuditLogEntry := audit.BuildAccountTerminationEventEntry(exampleSiteAdmin.ID, exampleUser.ID, exampleReason)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogAccountTerminationEvent(ctx, exampleSiteAdmin.ID, exampleUser.ID, exampleReason)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_LogCycleCookieSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

func TestClient_LogSuccessfulLoginEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

func TestClient_LogBannedUserLoginAttemptEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

func TestClient_LogUnsuccessfulLoginBadPasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

func TestClient_LogUnsuccessfulLoginBad2FATokenEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

func TestClient_LogLogoutEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

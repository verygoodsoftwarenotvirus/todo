package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
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

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		assert.NoError(t, c.UpdateUserReputation(ctx, exampleUser.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
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

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserReputation(ctx, exampleUser.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_LogUserBanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleServiceAdmin := fakes.BuildFakeUser()
		exampleUser := fakes.BuildFakeUser()
		exampleReason := "smells bad"
		//exampleAuditLogEntry := audit.BuildUserBanEventEntry(exampleServiceAdmin.ID, exampleUser.ID, exampleReason)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.LogUserBanEvent(ctx, exampleServiceAdmin.ID, exampleUser.ID, exampleReason)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_LogCycleCookieSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		//exampleAuditLogEntry := audit.BuildCycleCookieSecretEvent(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.LogCycleCookieSecretEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_LogSuccessfulLoginEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		//exampleAuditLogEntry := audit.BuildSuccessfulLoginEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.LogSuccessfulLoginEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_LogBannedUserLoginAttemptEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		//exampleAuditLogEntry := audit.BuildBannedUserLoginAttemptEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.LogBannedUserLoginAttemptEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_LogUnsuccessfulLoginBadPasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		//exampleAuditLogEntry := audit.BuildUnsuccessfulLoginBadPasswordEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.LogUnsuccessfulLoginBadPasswordEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_LogUnsuccessfulLoginBad2FATokenEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		//exampleAuditLogEntry := audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.LogUnsuccessfulLoginBad2FATokenEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_LogLogoutEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		//exampleAuditLogEntry := audit.BuildLogoutEventEntry(exampleUser.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.LogLogoutEvent(ctx, exampleUser.ID)

		mock.AssertExpectationsForObjects(t, db)
	})
}

package admin

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/alexedwards/scs/v2/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminService_UserAccountStatusChangeHandler_BanningAccounts(T *testing.T) {
	T.Parallel()

	T.Run("banning users", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleInput.NewReputation = types.BannedAccountStatus

		udb := &mocktypes.AdminUserDataManager{}
		udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), helper.exampleInput.TargetUserID, *helper.exampleInput).Return(nil)
		helper.service.userDB = udb

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, helper.exampleInput.TargetUserID, helper.exampleInput.Reason).Return()
		helper.service.auditLog = auditLog

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, udb, auditLog)
	})

	T.Run("terminating users", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleInput.NewReputation = types.TerminatedAccountStatus

		udb := &mocktypes.AdminUserDataManager{}
		udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), helper.exampleInput.TargetUserID, *helper.exampleInput).Return(nil)
		helper.service.userDB = udb

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogAccountTerminationEvent", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, helper.exampleInput.TargetUserID, helper.exampleInput.Reason).Return()
		helper.service.auditLog = auditLog

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, udb, auditLog)
	})

	T.Run("with no input attached to request", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		var err error
		helper.req, err = http.NewRequest(http.MethodGet, "/blah", nil)
		require.NoError(t, err)

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error fetching request context", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, errors.New("blah")
		}

		helper.exampleInput.NewReputation = types.BannedAccountStatus

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.neuterAdminUser()

		helper.exampleInput.NewReputation = types.BannedAccountStatus

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)
	})

	T.Run("with admin that has inadequate permissions", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.neuterAdminUser()

		helper.exampleInput.NewReputation = types.BannedAccountStatus

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)
	})

	T.Run("with no such user in database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleInput.NewReputation = types.BannedAccountStatus

		udb := &mocktypes.AdminUserDataManager{}
		udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), helper.exampleInput.TargetUserID, *helper.exampleInput).Return(sql.ErrNoRows)
		helper.service.userDB = udb

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error writing new reputation to database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleInput.NewReputation = types.BannedAccountStatus

		udb := &mocktypes.AdminUserDataManager{}
		udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), helper.exampleInput.TargetUserID, *helper.exampleInput).Return(errors.New("blah"))
		helper.service.userDB = udb

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error destroying session", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		ms := &mockstore.MockStore{}
		ms.ExpectDelete("", errors.New("blah"))
		helper.service.sessionManager.Store = ms

		helper.exampleInput.NewReputation = types.BannedAccountStatus

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, helper.exampleInput.TargetUserID, helper.exampleInput.Reason).Return()
		helper.service.auditLog = auditLog

		udb := &mocktypes.AdminUserDataManager{}
		udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), helper.exampleInput.TargetUserID, *helper.exampleInput).Return(nil)
		helper.service.userDB = udb

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})
}

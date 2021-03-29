package admin

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/alexedwards/scs/v2/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminServiceHTTPRoutesTestSuite(t *testing.T) {
	suite.Run(t, new(adminServiceHTTPRoutesTestHelper))
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_BanningAccounts() {
	t := helper.T()

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
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_TerminatingAccounts() {
	t := helper.T()

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
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_WithMissingInput() {
	t := helper.T()

	var err error
	helper.req, err = http.NewRequest(http.MethodGet, "/blah", nil)
	require.NoError(t, err)

	helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusBadRequest, helper.res.Code)
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_WithErrorFetchingSession() {
	t := helper.T()

	helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return nil, errors.New("blah")
	}

	helper.exampleInput.NewReputation = types.BannedAccountStatus

	helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_WithNonAdminUser() {
	t := helper.T()

	helper.neuterAdminUser()

	helper.exampleInput.NewReputation = types.BannedAccountStatus

	helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
	assert.Equal(t, http.StatusForbidden, helper.res.Code)
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_WithAdminThatHasInadequatePermissions() {
	t := helper.T()

	helper.neuterAdminUser()

	helper.exampleInput.NewReputation = types.BannedAccountStatus

	helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
	assert.Equal(t, http.StatusForbidden, helper.res.Code)
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_WithNonexistentUser() {
	t := helper.T()

	helper.exampleInput.NewReputation = types.BannedAccountStatus

	udb := &mocktypes.AdminUserDataManager{}
	udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), helper.exampleInput.TargetUserID, *helper.exampleInput).Return(sql.ErrNoRows)
	helper.service.userDB = udb

	helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
	assert.Equal(t, http.StatusNotFound, helper.res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_WithErrorPerformingReputationUpdate() {
	t := helper.T()

	helper.exampleInput.NewReputation = types.BannedAccountStatus

	udb := &mocktypes.AdminUserDataManager{}
	udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), helper.exampleInput.TargetUserID, *helper.exampleInput).Return(errors.New("blah"))
	helper.service.userDB = udb

	helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (helper *adminServiceHTTPRoutesTestHelper) TestAdminService_UserAccountStatusChangeHandler_WithErrorDestroyingSession() {
	t := helper.T()

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
}

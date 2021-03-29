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
	suite.Run(t, new(adminServiceHTTPRoutesTestSuite))
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_BanningAccounts() {
	t := s.T()

	s.exampleInput.NewReputation = types.BannedAccountStatus

	udb := &mocktypes.AdminUserDataManager{}
	udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), s.exampleInput.TargetUserID, *s.exampleInput).Return(nil)
	s.service.userDB = udb

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, s.exampleInput.TargetUserID, s.exampleInput.Reason).Return()
	s.service.auditLog = auditLog

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusAccepted, s.res.Code)

	mock.AssertExpectationsForObjects(t, udb, auditLog)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_TerminatingAccounts() {
	t := s.T()

	s.exampleInput.NewReputation = types.TerminatedAccountStatus

	udb := &mocktypes.AdminUserDataManager{}
	udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), s.exampleInput.TargetUserID, *s.exampleInput).Return(nil)
	s.service.userDB = udb

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogAccountTerminationEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, s.exampleInput.TargetUserID, s.exampleInput.Reason).Return()
	s.service.auditLog = auditLog

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusAccepted, s.res.Code)

	mock.AssertExpectationsForObjects(t, udb, auditLog)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithMissingInput() {
	t := s.T()

	var err error
	s.req, err = http.NewRequest(http.MethodGet, "/blah", nil)
	require.NoError(t, err)

	s.service.UserAccountStatusChangeHandler(s.res, s.req)

	assert.Equal(t, http.StatusBadRequest, s.res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithErrorFetchingSession() {
	t := s.T()

	s.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return nil, errors.New("blah")
	}

	s.exampleInput.NewReputation = types.BannedAccountStatus

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithNonAdminUser() {
	t := s.T()

	s.neuterAdminUser()

	s.exampleInput.NewReputation = types.BannedAccountStatus

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusForbidden, s.res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithAdminThatHasInadequatePermissions() {
	t := s.T()

	s.neuterAdminUser()

	s.exampleInput.NewReputation = types.BannedAccountStatus

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusForbidden, s.res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithNonexistentUser() {
	t := s.T()

	s.exampleInput.NewReputation = types.BannedAccountStatus

	udb := &mocktypes.AdminUserDataManager{}
	udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), s.exampleInput.TargetUserID, *s.exampleInput).Return(sql.ErrNoRows)
	s.service.userDB = udb

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusNotFound, s.res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithErrorPerformingReputationUpdate() {
	t := s.T()

	s.exampleInput.NewReputation = types.BannedAccountStatus

	udb := &mocktypes.AdminUserDataManager{}
	udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), s.exampleInput.TargetUserID, *s.exampleInput).Return(errors.New("blah"))
	s.service.userDB = udb

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithErrorDestroyingSession() {
	t := s.T()

	ms := &mockstore.MockStore{}
	ms.ExpectDelete("", errors.New("blah"))
	s.service.sessionManager.Store = ms

	s.exampleInput.NewReputation = types.BannedAccountStatus

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, s.exampleInput.TargetUserID, s.exampleInput.Reason).Return()
	s.service.auditLog = auditLog

	udb := &mocktypes.AdminUserDataManager{}
	udb.On("UpdateUserAccountStatus", mock.MatchedBy(testutil.ContextMatcher), s.exampleInput.TargetUserID, *s.exampleInput).Return(nil)
	s.service.userDB = udb

	s.service.UserAccountStatusChangeHandler(s.res, s.req)
	assert.Equal(t, http.StatusAccepted, s.res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

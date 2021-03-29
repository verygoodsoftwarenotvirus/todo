package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
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

type adminServiceHTTPRoutesTestSuite struct {
	suite.Suite

	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
}

var _ suite.SetupTestSuite = (*adminServiceHTTPRoutesTestSuite)(nil)

func (s *adminServiceHTTPRoutesTestSuite) SetupTest() {
	t := s.T()

	s.service = buildTestService(t)
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
	s.exampleAccount = fakes.BuildFakeAccount()

	var err error
	s.ctx, err = s.service.sessionManager.Load(context.Background(), "")
	require.NoError(t, err)

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(s.T(), err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	s.service.userIDFetcher = func(req *http.Request) uint64 {
		return s.exampleUser.ID
	}
}

var _ suite.WithStats = (*adminServiceHTTPRoutesTestSuite)(nil)

func (s *adminServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 9

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
}

func (s *adminServiceHTTPRoutesTestSuite) neuterAdminUser() {
	s.exampleUser.ServiceAdminPermissions = 0
	s.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
			s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
		})
	}
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_BanningAccounts() {
	t := s.T()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.BannedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	udb := &mocktypes.AdminUserDataManager{}
	udb.On(
		"UpdateUserAccountStatus",
		mock.Anything,
		exampleInput.TargetUserID,
		*exampleInput,
	).Return(nil)
	s.service.userDB = udb

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, exampleInput.TargetUserID, exampleInput.Reason).Return()
	s.service.auditLog = auditLog

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusAccepted, res.Code)

	mock.AssertExpectationsForObjects(t, udb, auditLog)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_TerminatingAccounts() {
	t := s.T()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.TerminatedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	udb := &mocktypes.AdminUserDataManager{}
	udb.On(
		"UpdateUserAccountStatus",
		mock.Anything,
		exampleInput.TargetUserID,
		*exampleInput,
	).Return(nil)
	s.service.userDB = udb

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogAccountTerminationEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, exampleInput.TargetUserID, exampleInput.Reason).Return()
	s.service.auditLog = auditLog

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusAccepted, res.Code)

	mock.AssertExpectationsForObjects(t, udb, auditLog)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithMissingInput() {
	t := s.T()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.UserAccountStatusChangeHandler(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithErrorFetchingSession() {
	t := s.T()

	s.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return nil, errors.New("blah")
	}

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.BannedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithNonAdminUser() {
	t := s.T()

	s.neuterAdminUser()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.BannedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusForbidden, res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithAdminThatHasInadequatePermissions() {
	t := s.T()

	s.neuterAdminUser()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.BannedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusForbidden, res.Code)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithNonexistentUser() {
	t := s.T()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.BannedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	udb := &mocktypes.AdminUserDataManager{}
	udb.On(
		"UpdateUserAccountStatus",
		mock.Anything,
		exampleInput.TargetUserID,
		*exampleInput,
	).Return(sql.ErrNoRows)
	s.service.userDB = udb

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusNotFound, res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithErrorPerformingReputationUpdate() {
	t := s.T()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.BannedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	udb := &mocktypes.AdminUserDataManager{}
	udb.On(
		"UpdateUserAccountStatus",
		mock.Anything,
		exampleInput.TargetUserID,
		*exampleInput,
	).Return(errors.New("blah"))
	s.service.userDB = udb

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *adminServiceHTTPRoutesTestSuite) TestAdminService_UserAccountStatusChangeHandler_WithErrorDestroyingSession() {
	t := s.T()

	ms := &mockstore.MockStore{}
	ms.ExpectDelete("", errors.New("blah"))
	s.service.sessionManager.Store = ms

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NotNil(t, req)
	require.NoError(t, err)

	exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
	exampleInput.NewReputation = types.BannedAccountStatus

	req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, exampleInput.TargetUserID, exampleInput.Reason).Return()
	s.service.auditLog = auditLog

	udb := &mocktypes.AdminUserDataManager{}
	udb.On(
		"UpdateUserAccountStatus",
		mock.Anything,
		exampleInput.TargetUserID,
		*exampleInput,
	).Return(nil)
	s.service.userDB = udb

	s.service.UserAccountStatusChangeHandler(res, req)
	assert.Equal(t, http.StatusAccepted, res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

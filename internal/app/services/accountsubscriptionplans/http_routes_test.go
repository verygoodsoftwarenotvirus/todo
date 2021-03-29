package accountsubscriptionplans

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountsSubscriptionPlansServiceHTTPRoutes(t *testing.T) {
	suite.Run(t, new(accountSubscriptionPlansServiceHTTPRoutesTestSuite))
}

type accountSubscriptionPlansServiceHTTPRoutesTestSuite struct {
	suite.Suite

	ctx                            context.Context
	req                            *http.Request
	res                            *httptest.ResponseRecorder
	service                        *service
	exampleUser                    *types.User
	exampleAccount                 *types.Account
	exampleAccountSubscriptionPlan *types.AccountSubscriptionPlan
	exampleInput                   *types.AccountSubscriptionPlanCreationInput
}

var _ suite.SetupTestSuite = (*accountSubscriptionPlansServiceHTTPRoutesTestSuite)(nil)

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) SetupTest() {
	t := s.T()

	s.ctx = context.Background()
	s.service = buildTestService()
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccountForUser(s.exampleUser)
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	s.exampleAccountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	s.exampleInput = fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(t, err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	s.service.accountSubscriptionPlanIDFetcher = func(req *http.Request) uint64 {
		return s.exampleAccountSubscriptionPlan.ID
	}

	s.res = httptest.NewRecorder()
	s.req, err = http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, s.req)
}

var _ suite.WithStats = (*accountSubscriptionPlansServiceHTTPRoutesTestSuite)(nil)

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 17

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ListHandler() {
	t := s.T()

	exampleAccountSubscriptionPlanList := fakes.BuildFakeAccountSubscriptionPlanList()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return(exampleAccountSubscriptionPlanList, nil)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlanList{}))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d ins.service.atus response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ListHandler_WithNoRowsReturned() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AccountSubscriptionPlanList)(nil), sql.ErrNoRows)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlanList{}))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected response status to be %d, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ListHandler_WithErrorFetchingAccountSubscriptionPlansFromDatabase() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AccountSubscriptionPlanList)(nil), errors.New("blah"))
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_CreateHandler() {
	t := s.T()

	exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("CreateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlanCreationInput{})).Return(s.exampleAccountSubscriptionPlan, nil)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	mc := &mockmetrics.UnitCounter{}
	mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
	s.service.planCounter = mc

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlan{}), http.StatusCreated)
	s.service.encoderDecoder = ed

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), createMiddlewareCtxKey, exampleInput))

	s.service.CreateHandler(s.res, s.req)

	assert.Equal(t, http.StatusCreated, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, mc, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_CreateHandler_WithoutInputAttached() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)

	assert.Equal(t, http.StatusBadRequest, s.res.Code)

	mock.AssertExpectationsForObjects(t, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_CreateHandler_WithErrorCreatingPlan() {
	t := s.T()

	exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("CreateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlanCreationInput{})).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), createMiddlewareCtxKey, exampleInput))

	s.service.CreateHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ReadHandler() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID).Return(s.exampleAccountSubscriptionPlan, nil)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlan{}))
	s.service.encoderDecoder = ed

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d ins.service.atus response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ReadHandler_WithNoSuchPlanInDatabase() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusNotFound, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ReadHandler_WithErrorFetchingAccountSubscriptionPlanFromDatabase() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_UpdateHandler() {
	t := s.T()

	exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(s.exampleAccountSubscriptionPlan)

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID).Return(s.exampleAccountSubscriptionPlan, nil)
	accountSubscriptionPlanDataManager.On("UpdateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlan{})).Return(nil)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlan{}))
	s.service.encoderDecoder = ed

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d ins.service.atus response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_UpdateHandler_WithoutUpdateInput() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusBadRequest, s.res.Code)

	mock.AssertExpectationsForObjects(t, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_UpdateHandler_WithNoResultsFromDatabase() {
	t := s.T()

	exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(s.exampleAccountSubscriptionPlan)

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusNotFound, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_UpdateHandler_WithErrorFetchingFromDatabase() {
	t := s.T()

	exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(s.exampleAccountSubscriptionPlan)

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_UpdateHandler_WithErrorPerformingUpdate() {
	t := s.T()

	exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(s.exampleAccountSubscriptionPlan)

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID).Return(s.exampleAccountSubscriptionPlan, nil)
	accountSubscriptionPlanDataManager.On("UpdateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlan{})).Return(errors.New("blah"))
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ArchiveHandler() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("ArchiveAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID, s.exampleUser.ID).Return(nil)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	mc := &mockmetrics.UnitCounter{}
	mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
	s.service.planCounter = mc

	s.service.ArchiveHandler(s.res, s.req)

	assert.Equal(t, http.StatusNoContent, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, mc)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ArchiveHandler_WithNoAccountSubscriptionPlanInDatabase() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("ArchiveAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID, s.exampleUser.ID).Return(sql.ErrNoRows)
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ArchiveHandler(s.res, s.req)

	assert.Equal(t, http.StatusNotFound, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) TestAccountSubscriptionPlansService_ArchiveHandler_WithErrorArchiving() {
	t := s.T()

	accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
	accountSubscriptionPlanDataManager.On("ArchiveAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccountSubscriptionPlan.ID, s.exampleUser.ID).Return(errors.New("blah"))
	s.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ArchiveHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
}

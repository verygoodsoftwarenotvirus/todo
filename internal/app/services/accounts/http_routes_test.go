package accounts

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	"github.com/stretchr/testify/suite"
)

func TestAccountsServiceHTTPRoutes(t *testing.T) {
	suite.Run(t, new(accountsServiceHTTPRoutesTestSuite))
}

type accountsServiceHTTPRoutesTestSuite struct {
	suite.Suite

	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
}

var _ suite.SetupTestSuite = (*accountsServiceHTTPRoutesTestSuite)(nil)

func (s *accountsServiceHTTPRoutesTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.service = buildTestService()
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccount()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(s.T(), err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	s.service.accountIDFetcher = func(req *http.Request) uint64 {
		return s.exampleAccount.ID
	}
}

var _ suite.WithStats = (*accountsServiceHTTPRoutesTestSuite)(nil)

func (s *accountsServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 17

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ListHandler() {
	t := s.T()

	exampleAccountList := fakes.BuildFakeAccountList()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return(exampleAccountList, nil)
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountList{}))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ListHandler(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ListHandler_WithNoRowsReturned() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return((*types.AccountList)(nil), sql.ErrNoRows)
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountList{}))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ListHandler(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ListHandler_WithErrorFetchingAccountsFromDatabase() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return((*types.AccountList)(nil), errors.New("blah"))
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ListHandler(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_CreateHandler() {
	t := s.T()

	exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{}), s.exampleUser.ID).Return(s.exampleAccount, nil)
	s.service.accountDataManager = accountDataManager

	mc := &mockmetrics.UnitCounter{}
	mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
	s.service.accountCounter = mc

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}), http.StatusCreated)
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

	s.service.CreateHandler(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, mc, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_CreateHandler_WithoutInputAttachedToRequest() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.CreateHandler(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)

	mock.AssertExpectationsForObjects(t, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_CreateHandler_WithErrorCreatingAccount() {
	t := s.T()

	exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{}), s.exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

	s.service.CreateHandler(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ReadHandler() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID).Return(s.exampleAccount, nil)
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ReadHandler(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ReadHandler_WithNoAccountInDatabase() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ReadHandler(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ReadHandler_WithErrorReadingFromDatabase() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ReadHandler(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_UpdateHandler() {
	t := s.T()

	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(s.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID).Return(s.exampleAccount, nil)
	accountDataManager.On("UpdateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Account{}), s.exampleUser.ID, mock.IsType([]*types.FieldChangeSummary{})).Return(nil)
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, s.service.accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_UpdateHandler_WithoutUpdateInputAttachedToRequest() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.UpdateHandler(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)

	mock.AssertExpectationsForObjects(t, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_UpdateHandler_WithNoRows() {
	t := s.T()

	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(s.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_UpdateHandler_WithErrorQueryingForAccount() {
	t := s.T()

	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(s.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_UpdateHandler_WithErrorUpdatingAccount() {
	t := s.T()

	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(s.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID).Return(s.exampleAccount, nil)
	accountDataManager.On("UpdateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Account{}), s.exampleUser.ID, mock.IsType([]*types.FieldChangeSummary{})).Return(errors.New("blah"))
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ArchiveHandler() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID, s.exampleUser.ID).Return(nil)
	s.service.accountDataManager = accountDataManager

	mc := &mockmetrics.UnitCounter{}
	mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
	s.service.accountCounter = mc

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ArchiveHandler(res, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, mc)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ArchiveHandler_WithNoAccountInDatabase() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID, s.exampleUser.ID).Return(sql.ErrNoRows)
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ArchiveHandler(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_ArchiveHandler_WithErrorWritingToDatabase() {
	t := s.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), s.exampleAccount.ID, s.exampleUser.ID, s.exampleUser.ID).Return(errors.New("blah"))
	s.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, req)
	require.NoError(t, err)

	s.service.ArchiveHandler(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

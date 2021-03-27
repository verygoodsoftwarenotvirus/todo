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

func TestAccountsService(t *testing.T) {
	suite.Run(t, new(accountsServiceTestSuite))
}

type accountsServiceTestSuite struct {
	suite.Suite

	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
}

var _ suite.SetupTestSuite = (*accountsServiceTestSuite)(nil)

func (s *accountsServiceTestSuite) SetupTest() {
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

func (s *accountsServiceTestSuite) TestService_ListHandler() {
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

func (s *accountsServiceTestSuite) TestListHandlerWithNoRowsReturned() {
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

func (s *accountsServiceTestSuite) TestListHandlerWithErrorFetchingAccountsFromDatabase() {
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

func (s *accountsServiceTestSuite) TestCreateHandler() {
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

func (s *accountsServiceTestSuite) TestCreateHandlerWithoutInputAttachedToRequest() {
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

func (s *accountsServiceTestSuite) TestCreateHandlerWithErrorCreatingAccount() {
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

func (s *accountsServiceTestSuite) TestReadHandler() {
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

func (s *accountsServiceTestSuite) TestReadHandlerWithNoAccountInDatabase() {
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

func (s *accountsServiceTestSuite) TestReadHandlerWithErrorReadingFromDatabase() {
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

func (s *accountsServiceTestSuite) TestUpdateHandler() {
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

func (s *accountsServiceTestSuite) TestUpdateHandlerWithoutUpdateInputAttachedToRequest() {
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

func (s *accountsServiceTestSuite) TestUpdateHandlerWithNoRows() {
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

func (s *accountsServiceTestSuite) TestUpdateHandlerWithErrorQueryingForAccount() {
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

func (s *accountsServiceTestSuite) TestUpdateHandlerWithErrorUpdatingAccount() {
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

func (s *accountsServiceTestSuite) TestArchiveHandler() {
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

func (s *accountsServiceTestSuite) TestArchiveHandlerWithNoAccountInDatabase() {
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

func (s *accountsServiceTestSuite) TestArchiveHandlerWithErrorWritingToDatabase() {
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

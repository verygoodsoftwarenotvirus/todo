package accounts

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountsServiceHTTPRoutes(t *testing.T) {
	suite.Run(t, new(accountsServiceHTTPRoutesTestSuite))
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

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

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

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

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

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

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

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), createMiddlewareCtxKey, exampleInput))

	s.service.CreateHandler(s.res, s.req)

	assert.Equal(t, http.StatusCreated, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, mc, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_CreateHandler_WithoutInputAttachedToRequest() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)

	assert.Equal(t, http.StatusBadRequest, s.res.Code)

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

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), createMiddlewareCtxKey, exampleInput))

	s.service.CreateHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

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

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

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

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusNotFound, s.res.Code)

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

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

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

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, s.service.accountDataManager, ed)
}

func (s *accountsServiceHTTPRoutesTestSuite) TestAccountsService_UpdateHandler_WithoutUpdateInputAttachedToRequest() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusBadRequest, s.res.Code)

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

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusNotFound, s.res.Code)

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

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

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

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), updateMiddlewareCtxKey, exampleInput))

	s.service.UpdateHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

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

	s.service.ArchiveHandler(s.res, s.req)

	assert.Equal(t, http.StatusNoContent, s.res.Code)

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

	s.service.ArchiveHandler(s.res, s.req)

	assert.Equal(t, http.StatusNotFound, s.res.Code)

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

	s.service.ArchiveHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

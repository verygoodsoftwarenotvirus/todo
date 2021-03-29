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
	suite.Run(t, new(accountsServiceHTTPRoutesTestHelper))
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ListHandler() {
	t := helper.T()

	exampleAccountList := fakes.BuildFakeAccountList()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return(exampleAccountList, nil)
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountList{}))
	helper.service.encoderDecoder = ed

	helper.service.ListHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ListHandler_WithNoRowsReturned() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return((*types.AccountList)(nil), sql.ErrNoRows)
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountList{}))
	helper.service.encoderDecoder = ed

	helper.service.ListHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ListHandler_WithErrorFetchingAccountsFromDatabase() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return((*types.AccountList)(nil), errors.New("blah"))
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ListHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_CreateHandler() {
	t := helper.T()

	exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(helper.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{}), helper.exampleUser.ID).Return(helper.exampleAccount, nil)
	helper.service.accountDataManager = accountDataManager

	mc := &mockmetrics.UnitCounter{}
	mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
	helper.service.accountCounter = mc

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}), http.StatusCreated)
	helper.service.encoderDecoder = ed

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

	helper.service.CreateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusCreated, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, mc, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_CreateHandler_WithoutInputAttachedToRequest() {
	t := helper.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.CreateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusBadRequest, helper.res.Code)

	mock.AssertExpectationsForObjects(t, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_CreateHandler_WithErrorCreatingAccount() {
	t := helper.T()

	exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(helper.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{}), helper.exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

	helper.service.CreateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ReadHandler() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID).Return(helper.exampleAccount, nil)
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}))
	helper.service.encoderDecoder = ed

	helper.service.ReadHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ReadHandler_WithNoAccountInDatabase() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ReadHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusNotFound, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ReadHandler_WithErrorReadingFromDatabase() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ReadHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_UpdateHandler() {
	t := helper.T()

	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID).Return(helper.exampleAccount, nil)
	accountDataManager.On("UpdateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Account{}), helper.exampleUser.ID, mock.IsType([]*types.FieldChangeSummary{})).Return(nil)
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}))
	helper.service.encoderDecoder = ed

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

	helper.service.UpdateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, helper.service.accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_UpdateHandler_WithoutUpdateInputAttachedToRequest() {
	t := helper.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.UpdateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusBadRequest, helper.res.Code)

	mock.AssertExpectationsForObjects(t, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_UpdateHandler_WithNoRows() {
	t := helper.T()

	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

	helper.service.UpdateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusNotFound, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_UpdateHandler_WithErrorQueryingForAccount() {
	t := helper.T()

	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

	helper.service.UpdateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_UpdateHandler_WithErrorUpdatingAccount() {
	t := helper.T()

	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID).Return(helper.exampleAccount, nil)
	accountDataManager.On("UpdateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Account{}), helper.exampleUser.ID, mock.IsType([]*types.FieldChangeSummary{})).Return(errors.New("blah"))
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

	helper.service.UpdateHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ArchiveHandler() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID, helper.exampleUser.ID).Return(nil)
	helper.service.accountDataManager = accountDataManager

	mc := &mockmetrics.UnitCounter{}
	mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
	helper.service.accountCounter = mc

	helper.service.ArchiveHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusNoContent, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, mc)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ArchiveHandler_WithNoAccountInDatabase() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID, helper.exampleUser.ID).Return(sql.ErrNoRows)
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ArchiveHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusNotFound, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

func (helper *accountsServiceHTTPRoutesTestHelper) TestAccountsService_ArchiveHandler_WithErrorWritingToDatabase() {
	t := helper.T()

	accountDataManager := &mocktypes.AccountDataManager{}
	accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, helper.exampleUser.ID, helper.exampleUser.ID).Return(errors.New("blah"))
	helper.service.accountDataManager = accountDataManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ArchiveHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

	mock.AssertExpectationsForObjects(t, accountDataManager, ed)
}

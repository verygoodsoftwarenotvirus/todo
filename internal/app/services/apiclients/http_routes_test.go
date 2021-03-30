package apiclients

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

func TestAPIClientsService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientList{}))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with no results returned from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), sql.ErrNoRows)
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientList{}))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error retrieving clients from the datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), errors.New("blah"))
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestAPIClientsService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return(helper.exampleAPIClient.ClientID, nil)
		sg.On("GenerateClientSecret").Return(helper.exampleAPIClient.ClientSecret, nil)
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB

		uc := &mockmetrics.UnitCounter{}
		uc.On("Increment", mock.MatchedBy(testutil.ContextMatcher)).Return()
		helper.service.apiClientCounter = uc

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientCreationResponse{}), http.StatusCreated)
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, uc, ed)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.req = testutil.BuildTestRequest(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error retrieving user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(false, nil)
		helper.service.authenticator = a

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})

	T.Run("with invalid password", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, errors.New("blah"))
		helper.service.authenticator = a

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})

	T.Run("with error generating client ID", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)

		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
	})

	T.Run("with error generating client secret", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return(helper.exampleAPIClient.ClientID, nil)
		sg.On("GenerateClientSecret").Return([]byte(nil), errors.New("blah"))
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
	})

	T.Run("with error creating API Client in data store", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return(helper.exampleAPIClient.ClientID, nil)
		sg.On("GenerateClientSecret").Return(helper.exampleAPIClient.ClientSecret, nil)
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return((*types.APIClient)(nil), errors.New("blah"))

		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
	})
}

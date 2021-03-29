package apiclients

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

func TestAPIClientsServiceHTTPRoutes(t *testing.T) {
	suite.Run(t, new(apiClientsServiceHTTPRoutesTestSuite))
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_ListHandler() {
	t := s.T()

	exampleAPIClientList := fakes.BuildFakeAPIClientList()

	mockDB := database.BuildMockDatabase()
	mockDB.APIClientDataManager.On(
		"GetAPIClients",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
		mock.IsType(&types.QueryFilter{}),
	).Return(exampleAPIClientList, nil)
	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientList{}))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)
	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_ListHandler_WithNoRowsReturned() {
	t := s.T()

	mockDB := database.BuildMockDatabase()
	mockDB.APIClientDataManager.On(
		"GetAPIClients",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
		mock.IsType(&types.QueryFilter{}),
	).Return((*types.APIClientList)(nil), sql.ErrNoRows)
	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientList{}))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)
	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_ListHandler_WithErrorRetrievingFromDatabase() {
	t := s.T()

	mockDB := database.BuildMockDatabase()
	mockDB.APIClientDataManager.On(
		"GetAPIClients",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
		mock.IsType(&types.QueryFilter{}),
	).Return((*types.APIClientList)(nil), errors.New("blah"))
	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler() {
	t := s.T()

	mockDB := database.BuildMockDatabase()

	mockDB.UserDataManager.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)

	a := &mockauth.Authenticator{}
	a.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = a

	sg := &mockSecretGenerator{}
	sg.On("GenerateClientID").Return(s.exampleAPIClient.ClientID, nil)
	sg.On("GenerateClientSecret").Return(s.exampleAPIClient.ClientSecret, nil)
	s.service.secretGenerator = sg

	mockDB.APIClientDataManager.On(
		"CreateAPIClient",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleInput,
		s.exampleUser.ID,
	).Return(s.exampleAPIClient, nil)

	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	uc := &mockmetrics.UnitCounter{}
	uc.On("Increment", mock.MatchedBy(testutil.ContextMatcher)).Return()
	s.service.apiClientCounter = uc

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientCreationResponse{}), http.StatusCreated)
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusCreated, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, a, sg, uc, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler_WithMissingInput() {
	t := s.T()

	s.req = testutil.BuildTestRequest(t)

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusBadRequest, s.res.Code)

	mock.AssertExpectationsForObjects(t, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler_WithErrorRetrievingUser() {
	t := s.T()

	mockDB := database.BuildMockDatabase()
	mockDB.UserDataManager.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return((*types.User)(nil), errors.New("blah"))
	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler_WithInvalidCredentials() {
	t := s.T()

	mockDB := database.BuildMockDatabase()
	mockDB.UserDataManager.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	mockDB.APIClientDataManager.On(
		"CreateAPIClient",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleInput,
		s.exampleUser.ID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	a := &mockauth.Authenticator{}
	a.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(false, nil)
	s.service.authenticator = a

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, a, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler_WithInvalidPassword() {
	t := s.T()

	mockDB := database.BuildMockDatabase()

	mockDB.UserDataManager.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	mockDB.APIClientDataManager.On(
		"CreateAPIClient",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleInput,
		s.exampleUser.ID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	a := &mockauth.Authenticator{}
	a.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, errors.New("blah"))
	s.service.authenticator = a

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, a, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler_WithErrorGeneratingClientID() {
	t := s.T()

	mockDB := database.BuildMockDatabase()

	mockDB.UserDataManager.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)

	a := &mockauth.Authenticator{}
	a.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = a

	sg := &mockSecretGenerator{}
	sg.On("GenerateClientID").Return("", errors.New("blah"))
	s.service.secretGenerator = sg

	mockDB.APIClientDataManager.On(
		"CreateAPIClient",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleInput,
		s.exampleUser.ID,
	).Return(s.exampleAPIClient, nil)

	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler_WithErrorGeneratingClientSecret() {
	t := s.T()

	mockDB := database.BuildMockDatabase()

	mockDB.UserDataManager.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)

	a := &mockauth.Authenticator{}
	a.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = a

	sg := &mockSecretGenerator{}
	sg.On("GenerateClientID").Return(s.exampleAPIClient.ClientID, nil)
	sg.On("GenerateClientSecret").Return([]byte(nil), errors.New("blah"))
	s.service.secretGenerator = sg

	mockDB.APIClientDataManager.On(
		"CreateAPIClient",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleInput,
		s.exampleUser.ID,
	).Return(s.exampleAPIClient, nil)

	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
}

func (s *apiClientsServiceHTTPRoutesTestSuite) TestAPIClientsService_CreateHandler_WithErrorCreatingAPIClientInDataStore() {
	t := s.T()

	mockDB := database.BuildMockDatabase()

	mockDB.UserDataManager.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)

	a := &mockauth.Authenticator{}
	a.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = a

	sg := &mockSecretGenerator{}
	sg.On("GenerateClientID").Return(s.exampleAPIClient.ClientID, nil)
	sg.On("GenerateClientSecret").Return(s.exampleAPIClient.ClientSecret, nil)
	s.service.secretGenerator = sg

	mockDB.APIClientDataManager.On(
		"CreateAPIClient",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleInput,
		s.exampleUser.ID,
	).Return((*types.APIClient)(nil), errors.New("blah"))

	s.service.apiClientDataManager = mockDB
	s.service.userDataManager = mockDB

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.CreateHandler(s.res, s.req)
	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
}

package apiclients

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	ctx := context.Background()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://verygoodsoftwarenotvirus.ru",
		nil,
	)

	require.NotNil(t, req)
	assert.NoError(t, err)
	return req
}

func TestAPIClientsServiceHTTPRoutes(t *testing.T) {
	suite.Run(t, new(apiClientsServiceHTTPRoutesTestSuite))
}

type apiClientsServiceHTTPRoutesTestSuite struct {
	suite.Suite

	ctx              context.Context
	req              *http.Request
	res              *httptest.ResponseRecorder
	service          *service
	exampleUser      *types.User
	exampleAccount   *types.Account
	exampleAPIClient *types.APIClient
	exampleInput     *types.APIClientCreationInput
}

var _ suite.SetupTestSuite = (*apiClientsServiceHTTPRoutesTestSuite)(nil)

func (s *apiClientsServiceHTTPRoutesTestSuite) SetupTest() {
	t := s.T()

	s.ctx = context.Background()
	s.service = buildTestService(t)
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	s.exampleAPIClient = fakes.BuildFakeAPIClient()
	s.exampleAPIClient.BelongsToUser = s.exampleUser.ID
	s.exampleInput = fakes.BuildFakeAPIClientCreationInputFromClient(s.exampleAPIClient)

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(s.T(), err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}

	req := buildRequest(t)

	s.req = req.WithContext(context.WithValue(req.Context(), types.RequestContextKey, reqCtx))
	s.req = s.req.WithContext(context.WithValue(s.req.Context(), creationMiddlewareCtxKey, s.exampleInput))

	s.res = httptest.NewRecorder()
}

var _ suite.WithStats = (*apiClientsServiceHTTPRoutesTestSuite)(nil)

func (s *apiClientsServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 11

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
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

	s.req = buildRequest(t)

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

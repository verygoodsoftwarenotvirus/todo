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

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
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

func Test_fetchUserID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		req := buildRequest(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

		// for the service.fetchUserID() call

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		s := buildTestService(t)

		actual := s.fetchUserID(req)
		assert.Equal(t, exampleUser.ID, actual)
	})

	T.Run("without context value present", func(t *testing.T) {
		t.Parallel()

		req := buildRequest(t)
		expected := uint64(0)
		s := buildTestService(t)

		actual := s.fetchUserID(req)
		assert.Equal(t, expected, actual)
	})
}

func TestService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientList{}))
		s.encoderDecoder = ed

		req := buildRequest(t)
		// for the service.fetchUserID() call

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.ListHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), sql.ErrNoRows)
		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientList{}))
		s.encoderDecoder = ed

		req := buildRequest(t)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.ListHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error fetching from apiClientDataManager", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), errors.New("blah"))
		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.ListHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToAccount = exampleUser.ID
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		s := buildTestService(t)
		mockDB := database.BuildMockDatabase()

		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return(exampleAPIClient.ClientID, nil)
		sg.On("GenerateClientSecret").Return(exampleAPIClient.ClientSecret, nil)
		s.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.Anything,
			exampleInput,
		).Return(exampleAPIClient, nil)

		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		uc := &mockmetrics.UnitCounter{}
		uc.On("Increment", mock.MatchedBy(testutil.ContextMatcher)).Return()
		s.apiClientCounter = uc

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.APIClientCreationResponse{}), http.StatusCreated)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, uc, ed)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error getting user", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToAccount = exampleUser.ID
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return((*types.User)(nil), errors.New("blah"))
		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToAccount = exampleUser.ID
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.Anything,
			exampleInput,
		).Return(exampleAPIClient, nil)
		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(false, nil)
		s.authenticator = a

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})

	T.Run("with error validating password", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToAccount = exampleUser.ID
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.Anything,
			exampleInput,
		).Return(exampleAPIClient, nil)
		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, errors.New("blah"))
		s.authenticator = a

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})

	T.Run("with error generating client ID", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToAccount = exampleUser.ID
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		s := buildTestService(t)
		mockDB := database.BuildMockDatabase()

		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return("", errors.New("blah"))
		s.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.Anything,
			exampleInput,
		).Return(exampleAPIClient, nil)

		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
	})

	T.Run("with error generating client secret", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToAccount = exampleUser.ID
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		s := buildTestService(t)
		mockDB := database.BuildMockDatabase()

		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return(exampleAPIClient.ClientID, nil)
		sg.On("GenerateClientSecret").Return([]byte(nil), errors.New("blah"))
		s.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.Anything,
			exampleInput,
		).Return(exampleAPIClient, nil)

		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
	})

	T.Run("with error creating API client", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToAccount = exampleUser.ID
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		s := buildTestService(t)
		mockDB := database.BuildMockDatabase()

		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = a

		sg := &mockSecretGenerator{}
		sg.On("GenerateClientID").Return(exampleAPIClient.ClientID, nil)
		sg.On("GenerateClientSecret").Return(exampleAPIClient.ClientSecret, nil)
		s.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			mock.Anything,
			exampleInput,
		).Return((*types.APIClient)(nil), errors.New("blah"))

		s.apiClientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		req = req.WithContext(
			context.WithValue(req.Context(), types.RequestContextKey, reqCtx),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, ed)
	})
}

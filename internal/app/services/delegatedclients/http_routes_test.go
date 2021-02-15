package delegatedclients

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

func Test_randString(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		actual := randString()
		assert.NotEmpty(t, actual)
	})
}

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
		exampleUser := fakes.BuildFakeUser()

		// for the service.fetchUserID() call
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
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

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleDelegatedClientList := fakes.BuildFakeDelegatedClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"GetDelegatedClients",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleDelegatedClientList, nil)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.DelegatedClientList{}))
		s.encoderDecoder = ed

		req := buildRequest(t)
		// for the service.fetchUserID() call
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.ListHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"GetDelegatedClients",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.DelegatedClientList)(nil), sql.ErrNoRows)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.DelegatedClientList{}))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.ListHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error fetching from clientDataManager", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"GetDelegatedClients",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.DelegatedClientList)(nil), errors.New("blah"))
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.ListHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.DelegatedClientDataManager.On(
			"CreateDelegatedClient",
			mock.Anything,
			exampleInput,
		).Return(exampleDelegatedClient, nil)
		s.clientDataManager = mockDB
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
		).Return(true, nil)
		s.authenticator = a

		uc := &mockmetrics.UnitCounter{}
		uc.On("Increment", mock.MatchedBy(testutil.ContextMatcher())).Return()
		s.delegatedClientCounter = uc

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.DelegatedClient{}), http.StatusCreated)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, uc, ed)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error getting user", func(t *testing.T) {
		t.Parallel()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return((*types.User)(nil), errors.New("blah"))
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.DelegatedClientDataManager.On(
			"CreateDelegatedClient",
			mock.Anything,
			exampleInput,
		).Return(exampleDelegatedClient, nil)
		s.clientDataManager = mockDB
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
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})

	T.Run("with error validating password", func(t *testing.T) {
		t.Parallel()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.DelegatedClientDataManager.On(
			"CreateDelegatedClient",
			mock.Anything,
			exampleInput,
		).Return(exampleDelegatedClient, nil)
		s.clientDataManager = mockDB
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
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})

	T.Run("with error creating delegated client", func(t *testing.T) {
		t.Parallel()

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.DelegatedClientDataManager.On(
			"CreateDelegatedClient",
			mock.Anything,
			exampleInput,
		).Return((*types.DelegatedClient)(nil), errors.New("blah"))
		s.clientDataManager = mockDB
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
		).Return(true, nil)
		s.authenticator = a

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})
}

package delegatedclients

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
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

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.DelegatedClientList{}))
		s.encoderDecoder = ed

		req := buildRequest(t)
		// for the service.fetchUserID() call
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
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

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.DelegatedClientList{}))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
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

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
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

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogDelegatedClientCreationEvent", mock.MatchedBy(testutil.ContextMatcher()), exampleDelegatedClient)
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.DelegatedClient{}), http.StatusCreated)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, uc, ed)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
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

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
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

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
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

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
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

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), creationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, ed)
	})
}

func TestService_ReadHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleDelegatedClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"GetDelegatedClient",
			mock.Anything,
			exampleDelegatedClient.ID,
			exampleDelegatedClient.BelongsToUser,
		).Return(exampleDelegatedClient, nil)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.DelegatedClient{}))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.ReadHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with no rows found", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleDelegatedClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"GetDelegatedClient",
			mock.Anything,
			exampleDelegatedClient.ID,
			exampleDelegatedClient.BelongsToUser,
		).Return(exampleDelegatedClient, sql.ErrNoRows)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.ReadHandler(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error fetching client from clientDataManager", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleDelegatedClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"GetDelegatedClient",
			mock.Anything,
			exampleDelegatedClient.ID,
			exampleDelegatedClient.BelongsToUser,
		).Return((*types.DelegatedClient)(nil), errors.New("blah"))
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.ReadHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleDelegatedClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"ArchiveDelegatedClient",
			mock.Anything,
			exampleDelegatedClient.ID,
			exampleDelegatedClient.BelongsToUser,
		).Return(nil)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		uc := &mockmetrics.UnitCounter{}
		uc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher())).Return()
		s.delegatedClientCounter = uc

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogDelegatedClientArchiveEvent", mock.MatchedBy(testutil.ContextMatcher()), exampleUser.ID, exampleDelegatedClient.ID)
		s.auditLog = auditLog

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.ArchiveHandler(res, req)
		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, uc)
	})

	T.Run("with no rows found", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleDelegatedClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"ArchiveDelegatedClient",
			mock.Anything,
			exampleDelegatedClient.ID,
			exampleDelegatedClient.BelongsToUser,
		).Return(sql.ErrNoRows)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.ArchiveHandler(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error deleting record", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClient.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleDelegatedClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.DelegatedClientDataManager.On(
			"ArchiveDelegatedClient",
			mock.Anything,
			exampleDelegatedClient.ID,
			exampleDelegatedClient.BelongsToUser,
		).Return(errors.New("blah"))
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)
		res := httptest.NewRecorder()

		s.ArchiveHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

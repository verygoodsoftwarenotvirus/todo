package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics/mock"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	oauth2models "gopkg.in/oauth2.v3/models"
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

func TestService_ExtractOAuth2ClientFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mh := &mockOAuth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: exampleOAuth2Client.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		req := buildRequest(t)
		req.URL.Path = fmt.Sprintf("/api/v1/%s", exampleOAuth2Client.Scopes[0])
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mh, mockDB)
	})

	T.Run("with error validating token", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		mh := &mockOAuth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return((*oauth2models.Token)(nil), errors.New("blah"))
		s.oauth2Handler = mh

		req := buildRequest(t)
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error fetching from clientDataManager", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mh := &mockOAuth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: exampleOAuth2Client.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return((*types.OAuth2Client)(nil), errors.New("blah"))
		s.clientDataManager = mockDB

		req := buildRequest(t)
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, mh, mockDB)
	})

	T.Run("with invalid scope", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mh := &mockOAuth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: exampleOAuth2Client.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		req := buildRequest(t)
		req.URL.Path = "/api/v1/stuff"
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, mh, mockDB)
	})
}

func TestService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientsForUser",
			mock.Anything,
			exampleUser.ID,
			mock.AnythingOfType("*types.QueryFilter"),
		).Return(exampleOAuth2ClientList, nil)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.OAuth2ClientList"))
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
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientsForUser",
			mock.Anything,
			exampleUser.ID,
			mock.AnythingOfType("*types.QueryFilter"),
		).Return((*types.OAuth2ClientList)(nil), sql.ErrNoRows)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.OAuth2ClientList"))
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
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientsForUser",
			mock.Anything,
			exampleUser.ID,
			mock.AnythingOfType("*types.QueryFilter"),
		).Return((*types.OAuth2ClientList)(nil), errors.New("blah"))
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return(exampleOAuth2Client, nil)
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
		uc.On("Increment", mock.Anything).Return()
		s.oauth2ClientCounter = uc

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogOAuth2ClientCreationEvent", mock.Anything, exampleOAuth2Client)
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponseWithStatus", mock.Anything, mock.AnythingOfType("*types.OAuth2Client"), http.StatusCreated)
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
		ed.On("EncodeNoInputResponse", mock.Anything)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		s.CreateHandler(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error getting user", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

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
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return(exampleOAuth2Client, nil)
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
		ed.On("EncodeUnauthorizedResponse", mock.Anything)
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

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return(exampleOAuth2Client, nil)
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
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

	T.Run("with error creating oauth2 client", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return((*types.OAuth2Client)(nil), errors.New("blah"))
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
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2Client.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Client",
			mock.Anything,
			exampleOAuth2Client.ID,
			exampleOAuth2Client.BelongsToUser,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.OAuth2Client"))
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
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2Client.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Client",
			mock.Anything,
			exampleOAuth2Client.ID,
			exampleOAuth2Client.BelongsToUser,
		).Return(exampleOAuth2Client, sql.ErrNoRows)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
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
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2Client.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Client",
			mock.Anything,
			exampleOAuth2Client.ID,
			exampleOAuth2Client.BelongsToUser,
		).Return((*types.OAuth2Client)(nil), errors.New("blah"))
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2Client.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"ArchiveOAuth2Client",
			mock.Anything,
			exampleOAuth2Client.ID,
			exampleOAuth2Client.BelongsToUser,
		).Return(nil)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		uc := &mockmetrics.UnitCounter{}
		uc.On("Decrement", mock.Anything).Return()
		s.oauth2ClientCounter = uc

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogOAuth2ClientArchiveEvent", mock.Anything, exampleUser.ID, exampleOAuth2Client.ID)
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
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2Client.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"ArchiveOAuth2Client",
			mock.Anything,
			exampleOAuth2Client.ID,
			exampleOAuth2Client.BelongsToUser,
		).Return(sql.ErrNoRows)
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
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
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleOAuth2Client.BelongsToUser = exampleUser.ID

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2Client.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"ArchiveOAuth2Client",
			mock.Anything,
			exampleOAuth2Client.ID,
			exampleOAuth2Client.BelongsToUser,
		).Return(errors.New("blah"))
		s.clientDataManager = mockDB
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

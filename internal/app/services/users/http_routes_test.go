package users

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics/mock"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestService_validateCredentialChangeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			examplePassword,
			exampleUser.TwoFactorSecret,
			exampleTOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = auth

		actual, sc := s.validateCredentialChangeRequest(
			ctx,
			exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Equal(t, exampleUser, actual)
		assert.Equal(t, http.StatusOK, sc)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with no rows found in database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return((*types.User)(nil), sql.ErrNoRows)
		s.userDataManager = mockDB

		actual, sc := s.validateCredentialChangeRequest(
			ctx,
			exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusNotFound, sc)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
		s.userDataManager = mockDB

		actual, sc := s.validateCredentialChangeRequest(
			ctx,
			exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusInternalServerError, sc)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			examplePassword,
			exampleUser.TwoFactorSecret,
			exampleTOTPToken,
			exampleUser.Salt,
		).Return(false, errors.New("blah"))
		s.authenticator = auth

		actual, sc := s.validateCredentialChangeRequest(
			ctx,
			exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusBadRequest, sc)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			examplePassword,
			exampleUser.TwoFactorSecret,
			exampleTOTPToken,
			exampleUser.Salt,
		).Return(false, nil)
		s.authenticator = auth

		actual, sc := s.validateCredentialChangeRequest(
			ctx,
			exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusUnauthorized, sc)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})
}

func TestService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, mock.Anything).Return(exampleUserList, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.UserList"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, mock.Anything).Return((*types.UserList)(nil), errors.New("blah"))
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ListHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_UsernameSearchHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUsername := fakes.BuildFakeUser().Username
		exampleUserList := fakes.BuildFakeUserList().Users

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("SearchForUsersByUsername", mock.Anything, exampleUsername).Return(exampleUserList, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("[]types.User"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		v := req.URL.Query()
		v.Set(types.SearchQueryKey, exampleUsername)
		req.URL.RawQuery = v.Encode()

		s.UsernameSearchHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleUsername := fakes.BuildFakeUser().Username

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("SearchForUsersByUsername", mock.Anything, exampleUsername).Return([]types.User{}, errors.New("blah"))
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		v := req.URL.Query()
		v.Set(types.SearchQueryKey, exampleUsername)
		req.URL.RawQuery = v.Encode()

		s.UsernameSearchHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("types.UserDatabaseCreationInput")).Return(exampleUser, nil)
		s.userDataManager = db

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.userCounter = mc

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUserCreationEvent", mock.Anything, exampleUser)
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponseWithStatus", mock.Anything, mock.AnythingOfType("*types.UserCreationResponse"), http.StatusCreated)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, mc, ed)
	})

	T.Run("with user creation disabled", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"user creation is disabled",
			http.StatusForbidden,
		)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.userCreationEnabled = false
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNoInputResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.userCreationEnabled = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, errors.New("blah"))
		s.authenticator = auth

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth, ed)
	})

	T.Run("with error generating two factor secret", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("types.UserDatabaseCreationInput")).Return(exampleUser, nil)
		s.userDataManager = db

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("", errors.New("blah"))
		s.secretGenerator = sg

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, sg, ed)
	})

	T.Run("with error generating salt", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("types.UserDatabaseCreationInput")).Return(exampleUser, nil)
		s.userDataManager = db

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("PRETENDTHISISASECRET", nil)
		sg.On("GenerateSalt").Return([]byte{}, errors.New("blah"))
		s.secretGenerator = sg

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, sg, ed)
	})

	T.Run("with error creating entry in database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("types.UserDatabaseCreationInput")).Return(exampleUser, errors.New("blah"))
		s.userDataManager = db

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, ed)
	})

	T.Run("with pre-existing entry in database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		s.encoderDecoder = encoding.ProvideResponseEncoder(s.logger)

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("types.UserDatabaseCreationInput")).Return(exampleUser, dbclient.ErrUserExists)
		s.userDataManager = db

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db)
	})
}

func TestService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.userIDFetcher = func(_ *http.Request) uint64 {
			return exampleUser.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.User"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with no rows found", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.userIDFetcher = func(_ *http.Request) uint64 {
			return exampleUser.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, sql.ErrNoRows)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.userIDFetcher = func(_ *http.Request) uint64 {
			return exampleUser.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, errors.New("blah"))
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_NewTOTPSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*types.User")).Return(nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = auth

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUserUpdateTwoFactorSecretEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponseWithStatus", mock.Anything, mock.AnythingOfType("*types.TOTPSecretRefreshResponse"), http.StatusAccepted)
		s.encoderDecoder = ed

		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNoInputResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with input attached but without user information", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeErrorResponse", mock.Anything, "invalid request", http.StatusUnauthorized)
		s.encoderDecoder = ed

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*types.User")).Return(nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(false, errors.New("blah"))
		s.authenticator = auth

		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error generating secret", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*types.User")).Return(nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = auth

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("", errors.New("blah"))
		s.secretGenerator = sg

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, sg, ed)
	})

	T.Run("with error updating user in database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*types.User")).Return(errors.New("blah"))
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = auth

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})
}

func TestService_TOTPSecretValidationHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.Anything, exampleUser.ID).Return(nil)
		s.userDataManager = mockDB

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUserVerifyTwoFactorSecretEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		s.TOTPSecretVerificationHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without valid input attached", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil

		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNoInputResponse", mock.Anything)
		s.encoderDecoder = ed

		s.TOTPSecretVerificationHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with secret already validated", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		og := exampleUser.TwoFactorSecretVerifiedOn
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		exampleUser.TwoFactorSecretVerifiedOn = og

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"TOTP secret already verified",
			http.StatusAlreadyReported,
		)
		s.encoderDecoder = ed

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler(res, req)

		assert.Equal(t, http.StatusAlreadyReported, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with invalid code", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)
		exampleInput.TOTPToken = "INVALID"

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error verifying two factor secret", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.Anything, exampleUser.ID).Return(errors.New("blah"))
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		s.TOTPSecretVerificationHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_UpdatePasswordHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		req = req.WithContext(
			context.WithValue(
				req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.Anything, exampleUser.ID, mock.AnythingOfType("string")).Return(nil)
		s.userDataManager = mockDB

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUserUpdatePasswordEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		auth.On("HashPassword", mock.Anything, exampleInput.NewPassword).Return("blah", nil)
		s.authenticator = auth

		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusSeeOther, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNoInputResponse", mock.Anything)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with input but without user info", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeErrorResponse", mock.Anything, "invalid request", http.StatusUnauthorized)
		s.encoderDecoder = ed

		exampleInput := fakes.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.Anything, exampleUser.ID, mock.AnythingOfType("string")).Return(nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(false, errors.New("blah"))
		s.authenticator = auth

		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.Anything, exampleUser.ID, mock.AnythingOfType("string")).Return(nil)
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		auth.On("HashPassword", mock.Anything, exampleInput.NewPassword).Return("blah", errors.New("blah"))
		s.authenticator = auth

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})

	T.Run("with error updating user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.Anything, exampleUser.ID, mock.AnythingOfType("string")).Return(errors.New("blah"))
		s.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		auth.On("HashPassword", mock.Anything, exampleInput.NewPassword).Return("blah", nil)
		s.authenticator = auth

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})
}

func TestService_Archive(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUser.ID
		}
		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleUser.ID).Return(nil)
		s.userDataManager = mockDB

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUserArchiveEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything)
		s.userCounter = mc

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, mc)
	})

	T.Run("with error updating database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUser.ID
		}
		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleUser.ID).Return(errors.New("blah"))
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_buildQRCode(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		actual := s.buildQRCode(ctx, exampleUser.Username, exampleUser.TwoFactorSecret)

		assert.NotEmpty(t, actual)
		assert.True(t, strings.HasPrefix(actual, base64ImagePrefix))
	})
}

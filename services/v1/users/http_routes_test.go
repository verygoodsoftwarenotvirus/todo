package users

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth/mock"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	mocknewsman "gitlab.com/verygoodsoftwarenotvirus/newsman/mock"
)

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequest(
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
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
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
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return((*models.User)(nil), sql.ErrNoRows)
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
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return((*models.User)(nil), errors.New("blah"))
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
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
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
		assert.Equal(t, http.StatusInternalServerError, sc)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with invalid login", func(t *testing.T) {
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
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
		s := buildTestService(t)

		exampleUserList := fakemodels.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, mock.Anything).Return(exampleUserList, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.UserList")).Return(nil)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ListHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, mock.Anything).Return((*models.UserList)(nil), errors.New("blah"))
		s.userDataManager = mockDB

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ListHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)

		exampleUserList := fakemodels.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, mock.Anything).Return(exampleUserList, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.UserList")).Return(errors.New("blah"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ListHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("models.UserDatabaseCreationInput")).Return(exampleUser, nil)
		s.userDataManager = db

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.userCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.AnythingOfType("newsman.Event")).Return()
		s.reporter = r

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.UserCreationResponse")).Return(nil)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				UserCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, mc, r, ed)
	})

	T.Run("with user creation disabled", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.userCreationEnabled = false
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.userCreationEnabled = true
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, errors.New("blah"))
		s.authenticator = auth

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				UserCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth)
	})

	T.Run("with error generating secret", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("models.UserDatabaseCreationInput")).Return(exampleUser, nil)
		s.userDataManager = db

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("", errors.New("blah"))
		s.secretGenerator = sg

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				UserCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, sg)
	})

	T.Run("with error creating entry in database", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("models.UserDatabaseCreationInput")).Return(exampleUser, errors.New("blah"))
		s.userDataManager = db

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				UserCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db)
	})

	T.Run("with pre-existing entry in database", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("models.UserDatabaseCreationInput")).Return(exampleUser, dbclient.ErrUserExists)
		s.userDataManager = db

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				UserCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.Anything, mock.AnythingOfType("models.UserDatabaseCreationInput")).Return(exampleUser, nil)
		s.userDataManager = db

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.userCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.AnythingOfType("newsman.Event")).Return()
		s.reporter = r

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.UserCreationResponse")).Return(errors.New("blah"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				UserCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.userCreationEnabled = true
		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, mc, r, ed)
	})
}

func TestService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		s.userIDFetcher = func(_ *http.Request) uint64 {
			return exampleUser.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with no rows found", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		s.userIDFetcher = func(_ *http.Request) uint64 {
			return exampleUser.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, sql.ErrNoRows)
		s.userDataManager = mockDB

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		s.userIDFetcher = func(_ *http.Request) uint64 {
			return exampleUser.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, errors.New("blah"))
		s.userDataManager = mockDB

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		s.userIDFetcher = func(_ *http.Request) uint64 {
			return exampleUser.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("blah"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_NewTOTPSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
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
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.TOTPSecretRefreshResponse")).Return(nil)
		s.encoderDecoder = ed

		s.NewTOTPSecretHandler()(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.NewTOTPSecretHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with input attached but without user information", func(t *testing.T) {
		s := buildTestService(t)

		exampleInput := fakemodels.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.NewTOTPSecretHandler()(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
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

		s.NewTOTPSecretHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error generating secret", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
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

		s.NewTOTPSecretHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error updating user in database", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("blah"))
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

		s.NewTOTPSecretHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeTOTPSecretRefreshInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
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
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.TOTPSecretRefreshResponse")).Return(errors.New("blah"))
		s.encoderDecoder = ed

		s.NewTOTPSecretHandler()(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})
}

func TestService_TOTPSecretValidationHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakemodels.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.Anything, exampleUser.ID).Return(nil)
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler()(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without valid input attached", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil

		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakemodels.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return((*models.User)(nil), errors.New("blah"))
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with secret already validated", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		og := exampleUser.TwoFactorSecretVerifiedOn
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakemodels.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		exampleUser.TwoFactorSecretVerifiedOn = og

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler()(res, req)

		assert.Equal(t, http.StatusAlreadyReported, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with invalid code", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakemodels.BuildFakeTOTPSecretValidationInputForUser(exampleUser)
		exampleInput.TOTPToken = "INVALID"

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error updating user", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakemodels.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.Anything, exampleUser.ID).Return(nil)
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler()(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error verifying two factor secret", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakemodels.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				TOTPSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.Anything, exampleUser.ID).Return(errors.New("blah"))
		s.userDataManager = mockDB

		s.TOTPSecretVerificationHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_UpdatePasswordHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakePasswordUpdateInput()

		req = req.WithContext(
			context.WithValue(
				req.Context(),
				PasswordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
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

		s.UpdatePasswordHandler()(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.UpdatePasswordHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with input but without user info", func(t *testing.T) {
		s := buildTestService(t)

		exampleInput := fakemodels.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				PasswordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.UpdatePasswordHandler()(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				PasswordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
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

		s.UpdatePasswordHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				PasswordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
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

		s.UpdatePasswordHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error updating user", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakePasswordUpdateInput()

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				PasswordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.UserIDKey,
				exampleUser.ID,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("blah"))
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

		s.UpdatePasswordHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})
}

func TestService_Archive(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUser.ID
		}
		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleUser.ID).Return(nil)
		s.userDataManager = mockDB

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.AnythingOfType("newsman.Event")).Return()
		s.reporter = r

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything)
		s.userCounter = mc

		s.ArchiveHandler()(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, r, mc)
	})

	T.Run("with error updating database", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUser.ID
		}
		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleUser.ID).Return(errors.New("blah"))
		s.userDataManager = mockDB

		s.ArchiveHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildQRCode(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		actual := s.buildQRCode(ctx, exampleUser.Username, exampleUser.TwoFactorSecret)

		assert.NotEmpty(t, actual)
		assert.True(t, strings.HasPrefix(actual, "data:image/jpeg;base64,"))
	})
}

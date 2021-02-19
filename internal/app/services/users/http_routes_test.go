package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	mockuploads "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/mock"

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
		examplePassword := "authentication"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
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
		examplePassword := "authentication"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return((*types.User)(nil), sql.ErrNoRows)
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
		examplePassword := "authentication"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
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
		examplePassword := "authentication"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
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
		examplePassword := "authentication"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
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
		mockDB.UserDataManager.On("GetUsers", mock.MatchedBy(testutil.ContextMatcher), mock.Anything).Return(exampleUserList, nil)
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.UserList{}))
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
		mockDB.UserDataManager.On("GetUsers", mock.MatchedBy(testutil.ContextMatcher), mock.Anything).Return((*types.UserList)(nil), errors.New("blah"))
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
		mockDB.UserDataManager.On("SearchForUsersByUsername", mock.MatchedBy(testutil.ContextMatcher), exampleUsername).Return(exampleUserList, nil)
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType([]*types.User{}))
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
		mockDB.UserDataManager.On("SearchForUsersByUsername", mock.MatchedBy(testutil.ContextMatcher), exampleUsername).Return([]*types.User{}, errors.New("blah"))
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(types.UserDataStoreCreationInput{})).Return(exampleUser, nil)
		s.userDataManager = db

		db.AccountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{})).Return(exampleAccount, nil)
		s.accountDataManager = db

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
		s.userCounter = mc

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.UserCreationResponse{}), http.StatusCreated)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.authSettings.EnableUserSignup = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, mc, ed)
	})

	T.Run("with user creation disabled", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			mock.Anything,
			"user creation is disabled",
			http.StatusForbidden,
		)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.authSettings.EnableUserSignup = false
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.authSettings.EnableUserSignup = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error authentication authentication", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(exampleUser.HashedPassword, errors.New("blah"))
		s.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.authSettings.EnableUserSignup = true
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
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(types.UserDataStoreCreationInput{})).Return(exampleUser, nil)
		s.userDataManager = db

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("", errors.New("blah"))
		s.secretGenerator = sg

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.authSettings.EnableUserSignup = true
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
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(types.UserDataStoreCreationInput{})).Return(exampleUser, nil)
		s.userDataManager = db

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("PRETENDTHISISASECRET", nil)
		sg.On("GenerateSalt").Return([]byte{}, errors.New("blah"))
		s.secretGenerator = sg

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.authSettings.EnableUserSignup = true
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
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(exampleUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(types.UserDataStoreCreationInput{})).Return(exampleUser, errors.New("blah"))
		s.userDataManager = db

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		s.authSettings.EnableUserSignup = true
		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, ed)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.User{}))
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, sql.ErrNoRows)
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, errors.New("blah"))
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
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

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.TOTPSecretRefreshResponse{}), http.StatusAccepted)
		s.encoderDecoder = ed

		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.NewTOTPSecretHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with input attached but without user information", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
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

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(errors.New("blah"))
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

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(nil)
		s.userDataManager = mockDB

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
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
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
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		exampleUser.TwoFactorSecretVerifiedOn = og

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			mock.Anything,
			"TOTP secret already verified",
			http.StatusAlreadyReported,
		)
		s.encoderDecoder = ed

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
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
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)
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
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
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
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(errors.New("blah"))
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, mock.AnythingOfType("string")).Return(nil)
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
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.NewPassword).Return("blah", nil)
		s.authenticator = auth

		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusSeeOther, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with input but without user info", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, mock.AnythingOfType("string")).Return(nil)
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

	T.Run("with error authentication authentication", func(t *testing.T) {
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, mock.AnythingOfType("string")).Return(nil)
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
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.NewPassword).Return("blah", errors.New("blah"))
		s.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, mock.AnythingOfType("string")).Return(errors.New("blah"))
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
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.NewPassword).Return("blah", nil)
		s.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		s.UpdatePasswordHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})
}

func TestService_AvatarUploadHandler(T *testing.T) {
	T.Parallel()

	// these aren't very good tests, because the major request work is handled by interfaces.

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := fakes.BuildFakeUser()

		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.AnythingOfType("*http.Request"), "avatar").Return(returnImage, nil)
		s.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On("SaveFile", mock.MatchedBy(testutil.ContextMatcher), fmt.Sprintf("avatar_%d", exampleUser.ID), returnImage.Data).Return(nil)
		s.uploadManager = um

		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
		s.userDataManager = mockDB

		s.AvatarUploadHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um)
	})

	T.Run("without session info", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.AvatarUploadHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := fakes.BuildFakeUser()

		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		s.encoderDecoder = ed

		s.AvatarUploadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error processing image", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := fakes.BuildFakeUser()

		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.AnythingOfType("*http.Request"), "avatar").Return((*images.Image)(nil), errors.New("blah"))
		s.imageUploadProcessor = ip

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		s.encoderDecoder = ed

		s.AvatarUploadHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, ed)
	})

	T.Run("with error saving file", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := fakes.BuildFakeUser()

		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		s.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.AnythingOfType("*http.Request"), "avatar").Return(returnImage, nil)
		s.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On("SaveFile", mock.MatchedBy(testutil.ContextMatcher), fmt.Sprintf("avatar_%d", exampleUser.ID), returnImage.Data).Return(errors.New("blah"))
		s.uploadManager = um

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		s.encoderDecoder = ed

		s.AvatarUploadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um, ed)
	})

	T.Run("with error updating user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := fakes.BuildFakeUser()

		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, types.SessionInfoFromUser(exampleUser)),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(errors.New("blah"))
		s.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.AnythingOfType("*http.Request"), "avatar").Return(returnImage, nil)
		s.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On("SaveFile", mock.MatchedBy(testutil.ContextMatcher), fmt.Sprintf("avatar_%d", exampleUser.ID), returnImage.Data).Return(nil)
		s.uploadManager = um

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		s.encoderDecoder = ed

		s.AvatarUploadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um, ed)
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
		mockDB.UserDataManager.On("ArchiveUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(nil)
		s.userDataManager = mockDB

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher))
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
		mockDB.UserDataManager.On("ArchiveUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(errors.New("blah"))
		s.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

package users

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	random "gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/images"
	mockuploads "gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_validateCredentialChangeRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			examplePassword,
			helper.exampleUser.TwoFactorSecret,
			exampleTOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		actual, sc := helper.service.validateCredentialChangeRequest(
			helper.ctx,
			helper.exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Equal(t, helper.exampleUser, actual)
		assert.Equal(t, http.StatusOK, sc)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with no rows found in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), sql.ErrNoRows)
		helper.service.userDataManager = mockDB

		actual, sc := helper.service.validateCredentialChangeRequest(
			helper.ctx,
			helper.exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusNotFound, sc)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		actual, sc := helper.service.validateCredentialChangeRequest(
			helper.ctx,
			helper.exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusInternalServerError, sc)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			examplePassword,
			helper.exampleUser.TwoFactorSecret,
			exampleTOTPToken,
		).Return(false, errors.New("blah"))
		helper.service.authenticator = auth

		actual, sc := helper.service.validateCredentialChangeRequest(
			helper.ctx,
			helper.exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusBadRequest, sc)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleTOTPToken := "123456"
		examplePassword := "password"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			examplePassword,
			helper.exampleUser.TwoFactorSecret,
			exampleTOTPToken,
		).Return(false, nil)
		helper.service.authenticator = auth

		actual, sc := helper.service.validateCredentialChangeRequest(
			helper.ctx,
			helper.exampleUser.ID,
			examplePassword,
			exampleTOTPToken,
		)

		assert.Nil(t, actual)
		assert.Equal(t, http.StatusUnauthorized, sc)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})
}

func TestService_UsernameSearchHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"SearchForUsersByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(exampleUserList.Users, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			mock.IsType([]*types.User{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		v := helper.req.URL.Query()
		v.Set(types.SearchQueryKey, helper.exampleUser.Username)
		helper.req.URL.RawQuery = v.Encode()

		helper.service.UsernameSearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"SearchForUsersByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return([]*types.User{}, errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		v := helper.req.URL.Query()
		v.Set(types.SearchQueryKey, helper.exampleUser.Username)
		helper.req.URL.RawQuery = v.Encode()

		helper.service.UsernameSearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutils.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			mock.IsType(&types.UserList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutils.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.UserList)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeUserRegistrationInputFromUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = helper.exampleUser.ID

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutils.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On(
			"CreateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.UserDataStoreCreationInput{}),
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutils.ContextMatcher).Return()
		helper.service.userCounter = unitCounter

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				types.UserRegistrationInputContextKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, unitCounter)
	})

	T.Run("with user creation disabled", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeUserRegistrationInputFromUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.authSettings.EnableUserSignup = false
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := &types.UserRegistrationInput{}
		exampleInput.Password = "a"
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = helper.exampleUser.ID

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error validating password", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeUserRegistrationInputFromUser(helper.exampleUser)
		exampleInput.Password = "a"
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = helper.exampleUser.ID

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				types.UserRegistrationInputContextKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeUserRegistrationInputFromUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutils.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, errors.New("blah"))
		helper.service.authenticator = auth

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				types.UserRegistrationInputContextKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	})

	T.Run("with error generating two factor secret", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeUserRegistrationInputFromUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutils.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On(
			"CreateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.UserDataStoreCreationInput{}),
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase32EncodedString",
			testutils.ContextMatcher,
			totpSecretSize,
		).Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				types.UserRegistrationInputContextKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, auth, db, sg)
	})

	T.Run("with error creating entry in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeUserRegistrationInputFromUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutils.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On(
			"CreateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.UserDataStoreCreationInput{}),
		).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = db

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				types.UserRegistrationInputContextKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, auth, db)
	})
}

func TestService_buildQRCode(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		actual := helper.service.buildQRCode(helper.ctx, helper.exampleUser.Username, helper.exampleUser.TwoFactorSecret)

		assert.NotEmpty(t, actual)
		assert.True(t, strings.HasPrefix(actual, base64ImagePrefix))
	})
}

func TestService_SelfHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			mock.IsType(&types.User{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SelfHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnauthorizedResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SelfHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no rows found", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, sql.ErrNoRows)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SelfHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SelfHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			mock.IsType(&types.User{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with no rows found", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, sql.ErrNoRows)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestService_TOTPSecretVerificationHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"MarkUserTwoFactorSecretAsVerified",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("without valid input attached", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := &types.TOTPSecretVerificationInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with invalid TOTP token", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)
		exampleInput.TOTPToken = "000000"
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"MarkUserTwoFactorSecretAsVerified",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.userDataManager = mockDB

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with secret already validated", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		og := helper.exampleUser.TwoFactorSecretVerifiedOn
		helper.exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		helper.exampleUser.TwoFactorSecretVerifiedOn = og

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAlreadyReported, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with invalid code", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)
		exampleInput.TOTPToken = "INVALID"
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error verifying two factor secret", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"MarkUserTwoFactorSecretAsVerified",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_NewTOTPSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := &types.TOTPSecretRefreshInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with input attached but without user information", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(false, errors.New("blah"))
		helper.service.authenticator = auth

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error generating secret", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase32EncodedString",
			testutils.ContextMatcher,
			totpSecretSize,
		).Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, sg)
	})

	T.Run("with error updating user in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})
}

func TestService_UpdatePasswordHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)

		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			testutils.ContextMatcher,
			exampleInput.NewPassword,
		).Return("blah", nil)
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := &types.PasswordUpdateInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with input but without user info", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(false, errors.New("blah"))
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with invalid password", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		exampleInput.NewPassword = "a"
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			testutils.ContextMatcher,
			exampleInput.NewPassword,
		).Return("blah", errors.New("blah"))
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error updating user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		auth := &authentication.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			testutils.ContextMatcher,
			exampleInput.NewPassword,
		).Return("blah", nil)
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})
}

func TestService_AvatarUploadHandler(T *testing.T) {
	T.Parallel()

	// these aren't very good tests, because the major request work is handled by interfaces.

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutils.ContextMatcher,
			testutils.HTTPRequestMatcher,
			"avatar",
		).Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On(
			"SaveFile",
			testutils.ContextMatcher,
			fmt.Sprintf("avatar_%s", helper.exampleUser.ID),
			returnImage.Data,
		).Return(nil)
		helper.service.uploadManager = um

		mockDB.UserDataManager.On(
			"UpdateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = mockDB

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um)
	})

	T.Run("without session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnauthorizedResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error processing image", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutils.ContextMatcher,
			testutils.HTTPRequestMatcher, "avatar").Return((*images.Image)(nil), errors.New("blah"))
		helper.service.imageUploadProcessor = ip

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, encoderDecoder)
	})

	T.Run("with error saving file", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutils.ContextMatcher,
			testutils.HTTPRequestMatcher,
			"avatar",
		).Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On(
			"SaveFile",
			testutils.ContextMatcher,
			fmt.Sprintf("avatar_%s", helper.exampleUser.ID),
			returnImage.Data,
		).Return(errors.New("blah"))
		helper.service.uploadManager = um

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um, encoderDecoder)
	})

	T.Run("with error updating user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutils.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutils.ContextMatcher,
			testutils.HTTPRequestMatcher,
			"avatar",
		).Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On(
			"SaveFile",
			testutils.ContextMatcher,
			fmt.Sprintf("avatar_%s", helper.exampleUser.ID),
			returnImage.Data,
		).Return(nil)
		helper.service.uploadManager = um

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um, encoderDecoder)
	})
}

func TestService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"ArchiveUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.userDataManager = mockDB

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutils.ContextMatcher).Return()
		helper.service.userCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, unitCounter)
	})

	T.Run("with no results in the database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"ArchiveUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.userDataManager = mockDB

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error updating database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"ArchiveUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

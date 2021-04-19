package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/passwords"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/random"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	mockuploads "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
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
			testutil.ContextMatcher,
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
			testutil.ContextMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.Username,
		).Return(exampleUserList.Users, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.Username,
		).Return([]*types.User{}, errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.UserList)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = helper.exampleUser.ID

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On(
			"CreateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.UserDataStoreCreationInput{}),
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.userCounter = unitCounter

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.UserCreationResponse{}), http.StatusCreated)
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auth, db, unitCounter, encoderDecoder)
	})

	T.Run("with user creation disabled", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"user creation is disabled",
			http.StatusForbidden,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.authSettings.EnableUserSignup = false
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error validating password", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)
		exampleInput.Password = "a"

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = helper.exampleUser.ID

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"password too weak",
			http.StatusBadRequest,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, errors.New("blah"))
		helper.service.authenticator = auth

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auth, encoderDecoder)
	})

	T.Run("with error generating two factor secret", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On(
			"CreateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.UserDataStoreCreationInput{}),
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase32EncodedString",
			testutil.ContextMatcher,
			totpSecretSize,
		).Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auth, db, sg, encoderDecoder)
	})

	T.Run("with error generating salt", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On(
			"CreateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.UserDataStoreCreationInput{}),
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase32EncodedString",
			testutil.ContextMatcher,
			totpSecretSize,
		).Return("PRETENDTHISISASECRET", nil)
		sg.On(
			"GenerateRawBytes",
			testutil.ContextMatcher,
			saltSize,
		).Return([]byte{}, errors.New("blah"))
		helper.service.secretGenerator = sg

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auth, db, sg, encoderDecoder)
	})

	T.Run("with error creating entry in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.Password,
		).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On(
			"CreateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.UserDataStoreCreationInput{}),
		).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = db

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				userCreationMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auth, db, encoderDecoder)
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnauthorizedResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, sql.ErrNoRows)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, sql.ErrNoRows)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestService_TOTPSecretValidationHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"VerifyUserTwoFactorSecret",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without valid input attached", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with secret already validated", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		og := helper.exampleUser.TwoFactorSecretVerifiedOn
		helper.exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.exampleUser.TwoFactorSecretVerifiedOn = og

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"TOTP secret already verified",
			http.StatusAlreadyReported,
		)
		helper.service.encoderDecoder = encoderDecoder

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutil.ContextMatcher,
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

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)
		exampleInput.TOTPToken = "INVALID"

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutil.ContextMatcher,
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

		helper.exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretVerificationMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserWithUnverifiedTwoFactorSecret",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"VerifyUserTwoFactorSecret",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestService_NewTOTPSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.TOTPSecretRefreshResponse{}), http.StatusAccepted)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, auth, encoderDecoder)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with input attached but without user information", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnauthorizedResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
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

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		helper.req = helper.req.WithContext(context.WithValue(
			helper.req.Context(),
			totpSecretRefreshMiddlewareCtxKey,
			exampleInput,
		))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase32EncodedString",
			testutil.ContextMatcher,
			totpSecretSize,
		).Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, auth, sg, encoderDecoder)
	})

	T.Run("with error updating user in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = auth

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, auth, encoderDecoder)
	})
}

func TestService_UpdatePasswordHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakePasswordUpdateInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.NewPassword,
		).Return("blah", nil)
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusSeeOther, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with input but without user info", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnauthorizedResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		exampleInput := fakes.BuildFakePasswordUpdateInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakePasswordUpdateInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
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

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		exampleInput.NewPassword = "a"

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
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

		exampleInput := fakes.BuildFakePasswordUpdateInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.NewPassword,
		).Return("blah", errors.New("blah"))
		helper.service.authenticator = auth

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, auth, encoderDecoder)
	})

	T.Run("with error updating user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakePasswordUpdateInput()

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				passwordChangeMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUserPassword",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType("string"),
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		auth := &passwords.MockAuthenticator{}
		auth.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			testutil.ContextMatcher,
			exampleInput.NewPassword,
		).Return("blah", nil)
		helper.service.authenticator = auth

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, auth, encoderDecoder)
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutil.ContextMatcher,
			testutil.RequestMatcher, "avatar").Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On(
			"SaveFile",
			testutil.ContextMatcher,
			fmt.Sprintf("avatar_%d", helper.exampleUser.ID), returnImage.Data,
		).Return(nil)
		helper.service.uploadManager = um

		mockDB.UserDataManager.On(
			"UpdateUser",
			testutil.ContextMatcher,
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

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnauthorizedResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutil.ContextMatcher,
			testutil.RequestMatcher, "avatar").Return((*images.Image)(nil), errors.New("blah"))
		helper.service.imageUploadProcessor = ip

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutil.ContextMatcher,
			testutil.RequestMatcher, "avatar").Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On(
			"SaveFile",
			testutil.ContextMatcher,
			fmt.Sprintf("avatar_%d", helper.exampleUser.ID), returnImage.Data,
		).Return(errors.New("blah"))
		helper.service.uploadManager = um

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On(
			"UpdateUser",
			testutil.ContextMatcher,
			mock.IsType(&types.User{}),
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On(
			"Process",
			testutil.ContextMatcher,
			testutil.RequestMatcher, "avatar").Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On(
			"SaveFile",
			testutil.ContextMatcher,
			fmt.Sprintf("avatar_%d", helper.exampleUser.ID), returnImage.Data,
		).Return(nil)
		helper.service.uploadManager = um

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.userDataManager = mockDB

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.userCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, unitCounter)
	})

	T.Run("with now results in the database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"ArchiveUser",
			testutil.ContextMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestService_AuditEntryHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetAuditLogEntriesForUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.userDataManager = userDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType([]*types.AuditLogEntry{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager, encoderDecoder)
	})

	T.Run("with sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetAuditLogEntriesForUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.userDataManager = userDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetAuditLogEntriesForUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.userDataManager = userDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager, encoderDecoder)
	})
}

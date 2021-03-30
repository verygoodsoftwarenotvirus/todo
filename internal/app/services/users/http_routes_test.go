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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			examplePassword,
			helper.exampleUser.TwoFactorSecret,
			exampleTOTPToken,
			helper.exampleUser.Salt,
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return((*types.User)(nil), sql.ErrNoRows)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			examplePassword,
			helper.exampleUser.TwoFactorSecret,
			exampleTOTPToken,
			helper.exampleUser.Salt,
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			examplePassword,
			helper.exampleUser.TwoFactorSecret,
			exampleTOTPToken,
			helper.exampleUser.Salt,
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

func TestService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUsers", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return(exampleUserList, nil)
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.UserList{}))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUsers", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.UserList)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})
}

func TestService_UsernameSearchHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("SearchForUsersByUsername", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.Username).Return(exampleUserList.Users, nil)
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType([]*types.User{}))
		helper.service.encoderDecoder = ed

		v := helper.req.URL.Query()
		v.Set(types.SearchQueryKey, helper.exampleUser.Username)
		helper.req.URL.RawQuery = v.Encode()

		helper.service.UsernameSearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("SearchForUsersByUsername", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.Username).Return([]*types.User{}, errors.New("blah"))
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		v := helper.req.URL.Query()
		v.Set(types.SearchQueryKey, helper.exampleUser.Username)
		helper.req.URL.RawQuery = v.Encode()

		helper.service.UsernameSearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
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

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.UserDataStoreCreationInput{})).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		db.AccountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{})).Return(exampleAccount, nil)
		helper.service.accountDataManager = db

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
		helper.service.userCounter = mc

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.UserCreationResponse{}), http.StatusCreated)
		helper.service.encoderDecoder = ed

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

		mock.AssertExpectationsForObjects(t, auth, db, mc, ed)
	})

	T.Run("with user creation disabled", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			"user creation is disabled",
			http.StatusForbidden,
		)
		helper.service.encoderDecoder = ed

		helper.service.authSettings.EnableUserSignup = false
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.authSettings.EnableUserSignup = true
		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error authentication authentication", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(helper.exampleUser.HashedPassword, errors.New("blah"))
		helper.service.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

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

		mock.AssertExpectationsForObjects(t, auth, ed)
	})

	T.Run("with error generating two factor secret", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.UserDataStoreCreationInput{})).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

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

		mock.AssertExpectationsForObjects(t, auth, db, sg, ed)
	})

	T.Run("with error generating salt", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.UserDataStoreCreationInput{})).Return(helper.exampleUser, nil)
		helper.service.userDataManager = db

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("PRETENDTHISISASECRET", nil)
		sg.On("GenerateSalt").Return([]byte{}, errors.New("blah"))
		helper.service.secretGenerator = sg

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

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

		mock.AssertExpectationsForObjects(t, auth, db, sg, ed)
	})

	T.Run("with error creating entry in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(helper.exampleUser)

		auth := &mockauth.Authenticator{}
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.Password).Return(helper.exampleUser.HashedPassword, nil)
		helper.service.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.On("CreateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.UserDataStoreCreationInput{})).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = db

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

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

		mock.AssertExpectationsForObjects(t, auth, db, ed)
	})
}

func TestService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.User{}))
		helper.service.encoderDecoder = ed

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with no rows found", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, sql.ErrNoRows)
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, errors.New("blah"))
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.TOTPSecretRefreshResponse{}), http.StatusAccepted)
		helper.service.encoderDecoder = ed

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with input attached but without user information", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, errors.New("blah")
		}

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

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

		mock.AssertExpectationsForObjects(t, ed)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
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

		helper.req = helper.req.WithContext(
			context.WithValue(
				helper.req.Context(),
				totpSecretRefreshMiddlewareCtxKey,
				exampleInput,
			),
		)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = auth

		sg := &mockSecretGenerator{}
		sg.On("GenerateTwoFactorSecret").Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, sg, ed)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.NewTOTPSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
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
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(nil)
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
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
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

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
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

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			"TOTP secret already verified",
			http.StatusAlreadyReported,
		)
		helper.service.encoderDecoder = ed

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
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
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
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
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.TOTPSecretVerificationHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, mock.IsType("string")).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.NewPassword).Return("blah", nil)
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusSeeOther, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with input but without user info", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, errors.New("blah")
		}

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

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

		mock.AssertExpectationsForObjects(t, ed)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, mock.IsType("string")).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(false, errors.New("blah"))
		helper.service.authenticator = auth

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth)
	})

	T.Run("with error authentication authentication", func(t *testing.T) {
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, mock.IsType("string")).Return(nil)
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.NewPassword).Return("blah", errors.New("blah"))
		helper.service.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
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
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUserPassword", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID, mock.IsType("string")).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		auth := &mockauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			helper.exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		auth.On("HashPassword", mock.MatchedBy(testutil.ContextMatcher), exampleInput.NewPassword).Return("blah", nil)
		helper.service.authenticator = auth

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdatePasswordHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, auth, ed)
	})
}

func TestService_AvatarUploadHandler(T *testing.T) {
	T.Parallel()

	// these aren't very good tests, because the major request work is handled by interfaces.

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&http.Request{}), "avatar").Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On("SaveFile", mock.MatchedBy(testutil.ContextMatcher), fmt.Sprintf("avatar_%d", helper.exampleUser.ID), returnImage.Data).Return(nil)
		helper.service.uploadManager = um

		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(nil)
		helper.service.userDataManager = mockDB

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um)
	})

	T.Run("without request context", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, errors.New("blah")
		}

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnauthorizedResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		helper.service.encoderDecoder = ed

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		helper.service.encoderDecoder = ed

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("with error processing image", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&http.Request{}), "avatar").Return((*images.Image)(nil), errors.New("blah"))
		helper.service.imageUploadProcessor = ip

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		helper.service.encoderDecoder = ed

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, ed)
	})

	T.Run("with error saving file", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&http.Request{}), "avatar").Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On("SaveFile", mock.MatchedBy(testutil.ContextMatcher), fmt.Sprintf("avatar_%d", helper.exampleUser.ID), returnImage.Data).Return(errors.New("blah"))
		helper.service.uploadManager = um

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		helper.service.encoderDecoder = ed

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um, ed)
	})

	T.Run("with error updating user", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.User{})).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		returnImage := &images.Image{}
		ip := &images.MockImageUploadProcessor{}
		ip.On("Process", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&http.Request{}), "avatar").Return(returnImage, nil)
		helper.service.imageUploadProcessor = ip

		um := &mockuploads.UploadManager{}
		um.On("SaveFile", mock.MatchedBy(testutil.ContextMatcher), fmt.Sprintf("avatar_%d", helper.exampleUser.ID), returnImage.Data).Return(nil)
		helper.service.uploadManager = um

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher())).Return()
		helper.service.encoderDecoder = ed

		helper.service.AvatarUploadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ip, um, ed)
	})
}

func TestService_Archive(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("ArchiveUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(nil)
		helper.service.userDataManager = mockDB

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher))
		helper.service.userCounter = mc

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, mc)
	})

	T.Run("with error updating database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("ArchiveUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(errors.New("blah"))
		helper.service.userDataManager = mockDB

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
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

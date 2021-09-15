package authentication

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/gorilla/securecookie"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationService_issueSessionManagedCookie(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		expectedToken, err := random.GenerateBase64EncodedString(helper.ctx, 32)
		require.NoError(t, err)

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(nil)
		sm.On("Put", testutils.ContextMatcher, userIDContextKey, helper.exampleUser.ID)
		sm.On("Put", testutils.ContextMatcher, accountIDContextKey, helper.exampleAccount.ID)
		sm.On("Commit", testutils.ContextMatcher).Return(expectedToken, time.Now().Add(24*time.Hour), nil)
		helper.service.sessionManager = sm

		cookie, err := helper.service.issueSessionManagedCookie(helper.ctx, helper.exampleAccount.ID, helper.exampleUser.ID)
		require.NotNil(t, cookie)
		assert.NoError(t, err)

		var actualToken string
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, cookie.Value, &actualToken))

		assert.Equal(t, expectedToken, actualToken)

		mock.AssertExpectationsForObjects(t, sm)
	})

	T.Run("with error loading from session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, errors.New("blah"))
		helper.service.sessionManager = sm

		cookie, err := helper.service.issueSessionManagedCookie(helper.ctx, helper.exampleAccount.ID, helper.exampleUser.ID)
		require.Nil(t, cookie)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, sm)
	})

	T.Run("with error renewing token", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(errors.New("blah"))
		helper.service.sessionManager = sm

		cookie, err := helper.service.issueSessionManagedCookie(helper.ctx, helper.exampleAccount.ID, helper.exampleUser.ID)
		require.Nil(t, cookie)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, sm)
	})

	T.Run("with error committing", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		expectedToken, err := random.GenerateBase64EncodedString(helper.ctx, 32)
		require.NoError(t, err)

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(nil)
		sm.On("Put", testutils.ContextMatcher, userIDContextKey, helper.exampleUser.ID)
		sm.On("Put", testutils.ContextMatcher, accountIDContextKey, helper.exampleAccount.ID)
		sm.On("Commit", testutils.ContextMatcher).Return(expectedToken, time.Now(), errors.New("blah"))
		helper.service.sessionManager = sm

		cookie, err := helper.service.issueSessionManagedCookie(helper.ctx, helper.exampleAccount.ID, helper.exampleUser.ID)
		require.Nil(t, cookie)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, sm)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		expectedToken, err := random.GenerateBase64EncodedString(helper.ctx, 32)
		require.NoError(t, err)

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(nil)
		sm.On("Put", testutils.ContextMatcher, userIDContextKey, helper.exampleUser.ID)
		sm.On("Put", testutils.ContextMatcher, accountIDContextKey, helper.exampleAccount.ID)
		sm.On("Commit", testutils.ContextMatcher).Return(expectedToken, time.Now().Add(24*time.Hour), nil)
		helper.service.sessionManager = sm

		helper.service.cookieManager = securecookie.New(
			securecookie.GenerateRandomKey(0),
			[]byte(""),
		)

		cookie, err := helper.service.issueSessionManagedCookie(helper.ctx, helper.exampleAccount.ID, helper.exampleUser.ID)
		require.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func TestAuthenticationService_LoginHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetDefaultAccountIDForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, nil)
		helper.service.accountMembershipManager = membershipDB

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		assert.NotEmpty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator, membershipDB)
	})

	T.Run("with missing login data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, &types.UserLoginInput{})

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))
	})

	T.Run("with no results in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return((*types.User)(nil), sql.ErrNoRows)
		helper.service.userDataManager = userDataManager

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with error retrieving user from datastore", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = userDataManager

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.ServiceAccountStatus = types.BannedUserAccountStatus
		helper.exampleUser.ReputationExplanation = "bad behavior"
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(false, nil)
		helper.service.authenticator = authenticator

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, errors.New("blah"))
		helper.service.authenticator = authenticator

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
	})

	T.Run("with invalid two factor code error returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(false, authentication.ErrInvalidTOTPToken)
		helper.service.authenticator = authenticator

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
	})

	T.Run("with non-matching password error returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(false, authentication.ErrPasswordDoesNotMatch)
		helper.service.authenticator = authenticator

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
	})

	T.Run("with error fetching default account", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetDefaultAccountIDForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return("", errors.New("blah"))
		helper.service.accountMembershipManager = membershipDB

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator, membershipDB)
	})

	T.Run("with error loading from session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetDefaultAccountIDForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, nil)
		helper.service.accountMembershipManager = membershipDB

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator, membershipDB, sm)
	})

	T.Run("with error renewing token in session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetDefaultAccountIDForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, nil)
		helper.service.accountMembershipManager = membershipDB

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator, membershipDB, sm)
	})

	T.Run("with error committing to session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetDefaultAccountIDForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, nil)
		helper.service.accountMembershipManager = membershipDB

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(nil)
		sm.On("Put", testutils.ContextMatcher, userIDContextKey, helper.exampleUser.ID)
		sm.On("Put", testutils.ContextMatcher, accountIDContextKey, helper.exampleAccount.ID)
		sm.On("Commit", testutils.ContextMatcher).Return("", time.Now(), errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, userDataManager, authenticator, membershipDB, sm)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",

			helper.service.config.Cookies.Name,
			mock.IsType("string"),
		).Return("", errors.New("blah"))
		helper.service.cookieManager = cb

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetDefaultAccountIDForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, nil)
		helper.service.accountMembershipManager = membershipDB

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, userDataManager, authenticator, membershipDB)
	})

	T.Run("with error building cookie and error encoding cookie response", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleLoginInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			helper.service.config.Cookies.Name,
			mock.IsType("string"),
		).Return("", errors.New("blah"))
		helper.service.cookieManager = cb

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		authenticator := &authentication.MockAuthenticator{}
		authenticator.On(
			"ValidateLogin",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetDefaultAccountIDForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, nil)
		helper.service.accountMembershipManager = membershipDB

		helper.service.BeginSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, userDataManager, authenticator, membershipDB)
	})
}

func TestAuthenticationService_ChangeActiveAccountHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(true, nil)
		helper.service.accountMembershipManager = accountMembershipManager

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		assert.NotEmpty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := &types.ChangeActiveAccountInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))
	})

	T.Run("with error checking user account membership", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(false, errors.New("blah"))
		helper.service.accountMembershipManager = accountMembershipManager

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager)
	})

	T.Run("without account authorization", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(false, nil)
		helper.service.accountMembershipManager = accountMembershipManager

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager)
	})

	T.Run("with error loading from session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(true, nil)
		helper.service.accountMembershipManager = accountMembershipManager

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager, sm)
	})

	T.Run("with error renewing token in session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(true, nil)
		helper.service.accountMembershipManager = accountMembershipManager

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager, sm)
	})

	T.Run("with error committing to session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(true, nil)
		helper.service.accountMembershipManager = accountMembershipManager

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(nil)
		sm.On("Put", testutils.ContextMatcher, userIDContextKey, helper.exampleUser.ID)
		sm.On("Put", testutils.ContextMatcher, accountIDContextKey, exampleInput.AccountID)
		sm.On("Commit", testutils.ContextMatcher).Return("", time.Now(), errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager, sm)
	})

	T.Run("with error renewing token in session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(true, nil)
		helper.service.accountMembershipManager = accountMembershipManager

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("RenewToken", testutils.ContextMatcher).Return(errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager, sm)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		accountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipManager.On(
			"UserIsMemberOfAccount",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
			exampleInput.AccountID,
		).Return(true, nil)
		helper.service.accountMembershipManager = accountMembershipManager

		cookieManager := &mockCookieEncoderDecoder{}
		cookieManager.On(
			"Encode",
			helper.service.config.Cookies.Name,
			mock.IsType("string"),
		).Return("", errors.New("blah"))
		helper.service.cookieManager = cookieManager

		helper.service.ChangeActiveAccountHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, accountMembershipManager)
	})
}

func TestAuthenticationService_LogoutHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.ctx, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		helper.service.EndSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusSeeOther, helper.res.Code)
		actualCookie := helper.res.Header().Get("Set-Cookie")
		assert.Contains(t, actualCookie, "Max-Age=0")
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.EndSessionHandler(helper.res, helper.req)

		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))
	})

	T.Run("with error loading from session manager", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.ctx, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(context.Background(), errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.EndSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		actualCookie := helper.res.Header().Get("Set-Cookie")
		assert.Empty(t, actualCookie)

		mock.AssertExpectationsForObjects(t, sm)
	})

	T.Run("with error deleting from session store", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sm := &mockSessionManager{}
		sm.On("Load", testutils.ContextMatcher, "").Return(helper.ctx, nil)
		sm.On("Destroy", testutils.ContextMatcher).Return(errors.New("blah"))
		helper.service.sessionManager = sm

		helper.service.EndSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		actualCookie := helper.res.Header().Get("Set-Cookie")
		assert.Empty(t, actualCookie)

		mock.AssertExpectationsForObjects(t, sm)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.ctx, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
		helper.service.cookieManager = securecookie.New(
			securecookie.GenerateRandomKey(0),
			[]byte(""),
		)

		helper.service.EndSessionHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	})
}

func TestAuthenticationService_StatusHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.StatusHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("with problem fetching session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.StatusHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})
}

func TestAuthenticationService_CycleSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.ServiceRoles = []string{authorization.ServiceAdminRole.String()}
		helper.setContextFetcher(t)

		helper.ctx, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
		c := helper.req.Cookies()[0]

		var token string
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))

		helper.service.CycleCookieSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, helper.res.Code)
		assert.Error(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))
	})

	T.Run("with error getting session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.ctx, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
		c := helper.req.Cookies()[0]

		var token string
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))

		helper.service.CycleCookieSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, helper.res.Code)
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))
	})

	T.Run("with invalid permissions", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.ctx, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
		c := helper.req.Cookies()[0]

		var token string
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))

		helper.service.CycleCookieSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, helper.res.Code)
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))
	})
}

func TestAuthenticationService_PASETOHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		helper.service.config.PASETO.Lifetime = time.Minute

		exampleInput := &types.PASETOCreationInput{
			AccountID:   helper.exampleAccount.ID,
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		expected := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = apiClientDataManager

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(expected, nil)
		helper.service.accountMembershipManager = membershipDB

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		// validate results

		var result *types.PASETOResponse
		require.NoError(t, json.NewDecoder(helper.res.Body).Decode(&result))

		assert.NotEmpty(t, result.Token)

		var targetPayload paseto.JSONToken
		require.NoError(t, paseto.NewV2().Decrypt(result.Token, helper.service.config.PASETO.LocalModeKey, &targetPayload, nil))

		assert.True(t, targetPayload.Expiration.After(time.Now().UTC()))

		payload := targetPayload.Get(pasetoDataKey)

		gobEncoding, err := base64.RawURLEncoding.DecodeString(payload)
		require.NoError(t, err)

		var actual *types.SessionContextData
		require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&actual))

		assert.NotNil(t, actual)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, userDataManager, membershipDB)
	})

	T.Run("does not issue token with longer lifetime than package maximum", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		helper.service.config.PASETO.Lifetime = 24 * time.Hour * 365 // one year

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		expected := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = apiClientDataManager

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(expected, nil)
		helper.service.accountMembershipManager = membershipDB

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		// validate results

		var result *types.PASETOResponse
		require.NoError(t, json.NewDecoder(helper.res.Body).Decode(&result))

		assert.NotEmpty(t, result.Token)

		var targetPayload paseto.JSONToken
		require.NoError(t, paseto.NewV2().Decrypt(result.Token, helper.service.config.PASETO.LocalModeKey, &targetPayload, nil))

		assert.True(t, targetPayload.Expiration.Before(time.Now().UTC().Add(maxPASETOLifetime)))

		payload := targetPayload.Get(pasetoDataKey)

		gobEncoding, err := base64.RawURLEncoding.DecodeString(payload)
		require.NoError(t, err)

		var actual *types.SessionContextData
		require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&actual))

		assert.NotNil(t, actual)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, userDataManager, membershipDB)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		helper.service.config.PASETO.Lifetime = time.Minute

		exampleInput := &types.PASETOCreationInput{}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid request time", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: 1,
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error decoding signature header", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(jsonBytes)
		require.NoError(t, macWriteErr)

		sigHeader := base32.HexEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error fetching API client", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		helper.service.apiClientManager = apiClientDataManager

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = apiClientDataManager

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = userDataManager

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, userDataManager)
	})

	T.Run("with error fetching account memberships", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = apiClientDataManager

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.SessionContextData)(nil), errors.New("blah"))
		helper.service.accountMembershipManager = membershipDB

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, userDataManager, membershipDB)
	})

	T.Run("with invalid checksum", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = apiClientDataManager

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write([]byte("lol"))
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager)
	})

	T.Run("with inadequate account permissions", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		helper.service.config.PASETO.Lifetime = time.Minute

		exampleInput := &types.PASETOCreationInput{
			AccountID:   helper.exampleAccount.ID,
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = apiClientDataManager

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		delete(helper.sessionCtxData.AccountPermissions, helper.exampleAccount.ID)

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.sessionCtxData, nil)
		helper.service.accountMembershipManager = membershipDB

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with token encryption error", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = nil

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByClientID",
			testutils.ContextMatcher,
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = apiClientDataManager

		userDataManager := &mocktypes.UserDataManager{}
		userDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = userDataManager

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.sessionCtxData, nil)
		helper.service.accountMembershipManager = membershipDB

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, userDataManager, membershipDB)
	})
}

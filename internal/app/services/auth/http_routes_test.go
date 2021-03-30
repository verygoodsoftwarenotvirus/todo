package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/bcrypt"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/gorilla/securecookie"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_DecodeCookieFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		_, userID, err := helper.service.getUserIDFromCookie(helper.ctx, helper.req)
		assert.NoError(t, err)
		assert.Equal(t, helper.exampleUser.ID, userID)
	})

	T.Run("with invalid cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		// begin building bad cookie.
		// NOTE: any code here is duplicated from service.buildAuthCookie
		// any changes made there might need to be reflected here.
		c := &http.Cookie{
			Name:     helper.service.config.Cookies.Name,
			Value:    "blah blah blah this is not a real cookie",
			Path:     "/",
			HttpOnly: true,
		}
		// end building bad cookie.
		helper.req.AddCookie(c)

		_, userID, err := helper.service.getUserIDFromCookie(helper.req.Context(), helper.req)
		assert.Error(t, err)
		assert.Zero(t, userID)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		_, userID, err := helper.service.getUserIDFromCookie(helper.req.Context(), helper.req)
		assert.Error(t, err)
		assert.Equal(t, err, http.ErrNoCookie)
		assert.Zero(t, userID)
	})
}

func TestAuthService_determineUserFromRequestCookie(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		actualUser, err := helper.service.determineUserFromRequestCookie(helper.ctx, helper.req)
		assert.Equal(t, helper.exampleUser, actualUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		actualUser, err := helper.service.determineUserFromRequestCookie(helper.req.Context(), helper.req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})

	T.Run("with error retrieving user from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		expectedError := errors.New("blah")
		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return((*types.User)(nil), expectedError)
		helper.service.userDataManager = udb

		actualUser, err := helper.service.determineUserFromRequestCookie(helper.req.Context(), helper.req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, udb)
	})
}

func TestAuthService_LoginHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx = context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, helper.exampleLoginInput)
		helper.req = helper.req.WithContext(helper.ctx)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = authr

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, helper.examplePerms, nil)
		helper.service.accountMembershipManager = membershipDB

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On(
			"LogSuccessfulLoginEvent",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		)
		helper.service.auditLog = auditLog

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		assert.NotEmpty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr, auditLog, membershipDB)
	})

	T.Run("with missing login data", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))
	})

	T.Run("with error retrieving user from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx = context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, helper.exampleLoginInput)
		helper.req = helper.req.WithContext(helper.ctx)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.Username,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = udb

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleUser.Reputation = types.BannedAccountStatus
		helper.exampleUser.ReputationExplanation = "bad behavior"

		helper.ctx = context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, helper.exampleLoginInput)
		helper.req = helper.req.WithContext(helper.ctx)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogBannedUserLoginAttemptEvent", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID)
		helper.service.auditLog = auditLog

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, auditLog)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx = context.WithValue(helper.ctx, userLoginInputMiddlewareCtxKey, helper.exampleLoginInput)
		helper.req = helper.req.WithContext(helper.ctx)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(false, nil)
		helper.service.authenticator = authr

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogUnsuccessfulLoginBadPasswordEvent", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID)
		helper.service.auditLog = auditLog

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr, auditLog)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx = context.WithValue(helper.ctx, userLoginInputMiddlewareCtxKey, helper.exampleLoginInput)
		helper.req = helper.req.WithContext(helper.ctx)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, errors.New("blah"))
		helper.service.authenticator = authr

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx = context.WithValue(helper.ctx, userLoginInputMiddlewareCtxKey, helper.exampleLoginInput)
		helper.req = helper.req.WithContext(helper.ctx)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",

			helper.service.config.Cookies.Name,
			mock.IsType("string"),
		).Return("", errors.New("blah"))
		helper.service.cookieManager = cb

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = authr

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, helper.examplePerms, nil)
		helper.service.accountMembershipManager = membershipDB

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, udb, authr, membershipDB)
	})

	T.Run("with error building cookie and error encoding cookie response", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx = context.WithValue(helper.ctx, userLoginInputMiddlewareCtxKey, helper.exampleLoginInput)
		helper.req = helper.req.WithContext(helper.ctx)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			helper.service.config.Cookies.Name,
			mock.IsType("string"),
		).Return("", errors.New("blah"))
		helper.service.cookieManager = cb

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.Username,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = authr

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, helper.examplePerms, nil)
		helper.service.accountMembershipManager = membershipDB

		helper.service.LoginHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		assert.Empty(t, helper.res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, udb, authr, membershipDB)
	})
}

func TestAuthService_LogoutHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogLogoutEvent", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID)
		helper.service.auditLog = auditLog

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		helper.service.LogoutHandler(helper.res, helper.req)

		actualCookie := helper.res.Header().Get("Set-Cookie")
		assert.Contains(t, actualCookie, "Max-Age=0")

		mock.AssertExpectationsForObjects(t, auditLog)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		var err error
		helper.ctx, err = helper.service.sessionManager.Load(helper.ctx, "")
		require.NoError(t, err)
		require.NoError(t, helper.service.sessionManager.RenewToken(helper.ctx))

		// Then make the privilege-level change.
		helper.service.sessionManager.Put(helper.ctx, userIDContextKey, helper.exampleUser.ID)
		helper.service.sessionManager.Put(helper.ctx, accountIDContextKey, helper.exampleAccount.ID)

		helper.service.LogoutHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
		helper.service.cookieManager = securecookie.New(
			securecookie.GenerateRandomKey(0),
			[]byte(""),
		)

		helper.service.LogoutHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	})
}

func TestAuthService_validateLogin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, nil)
		helper.service.authenticator = authr

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})

	T.Run("with too weak of a password hash", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, authentication.ErrPasswordHashTooWeak)
		helper.service.authenticator = authr

		authr.On(
			"HashPassword",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleLoginInput.Password,
		).Return("blah", nil)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"UpdateUser",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.User{}),
		).Return(nil)
		helper.service.userDataManager = udb

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authr, udb)
	})

	T.Run("with error re-hashing too-weak password hash", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		expectedErr := errors.New("arbitrary")

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, authentication.ErrPasswordHashTooWeak)

		authr.On(
			"HashPassword",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleLoginInput.Password,
		).Return("", expectedErr)
		helper.service.authenticator = authr

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})

	T.Run("with too weak a password hash and error updating user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		expectedErr := errors.New("arbitrary")

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(true, bcrypt.ErrCostTooLow)

		authr.On(
			"HashPassword",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleLoginInput.Password,
		).Return("blah", nil)
		helper.service.authenticator = authr

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"UpdateUser",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.User{}),
		).Return(expectedErr)
		helper.service.userDataManager = udb

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authr, udb)
	})

	T.Run("with error returned from validator", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		expectedErr := errors.New("arbitrary")

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(false, expectedErr)
		helper.service.authenticator = authr

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
			helper.exampleUser.Salt,
		).Return(false, nil)
		helper.service.authenticator = authr

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})
}

func TestAuthService_StatusHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		helper.service.StatusHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error retrieving user from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = udb

		helper.service.StatusHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})
}

func TestAuthService_CycleSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
		helper.setContextFetcher(t)

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogCycleCookieSecretEvent", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID)
		helper.service.auditLog = auditLog

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
		c := helper.req.Cookies()[0]

		var token string
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))

		helper.service.CycleCookieSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, helper.res.Code)
		assert.Error(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))

		mock.AssertExpectationsForObjects(t, auditLog)
	})

	T.Run("with error getting request context", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, errors.New("blah")
		}

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
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

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)
		c := helper.req.Cookies()[0]

		var token string
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))

		helper.service.CycleCookieSecretHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, helper.res.Code)
		assert.NoError(t, helper.service.cookieManager.Decode(helper.service.config.Cookies.Name, c.Value, &token))
	})
}

func TestAuthService_buildCookie(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		cookie, err := helper.service.buildCookie("example", time.Now().Add(helper.service.config.Cookies.Lifetime))
		assert.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	T.Run("with invalid cookie builder", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.cookieManager = securecookie.New(
			securecookie.GenerateRandomKey(0),
			[]byte(""),
		)

		cookie, err := helper.service.buildCookie("example", time.Now().Add(helper.service.config.Cookies.Lifetime))
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func TestAuthService_PASETOHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		helper.service.config.PASETO.Lifetime = time.Minute

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		expectedOutput := &types.RequestContext{
			User: types.UserRequestContext{
				ID:                      helper.exampleUser.ID,
				Status:                  helper.exampleUser.Reputation,
				ServiceAdminPermissions: helper.exampleUser.ServiceAdminPermissions,
			},
			ActiveAccountID:       helper.exampleAccount.ID,
			AccountPermissionsMap: helper.examplePerms,
		}

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, helper.examplePerms, nil)
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

		var si *types.RequestContext
		require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&si))

		assert.Equal(t, expectedOutput, si)

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
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

		expectedOutput := &types.RequestContext{
			User: types.UserRequestContext{
				ID:                      helper.exampleUser.ID,
				Status:                  helper.exampleUser.Reputation,
				ServiceAdminPermissions: helper.exampleUser.ServiceAdminPermissions,
			},
			ActiveAccountID:       helper.exampleAccount.ID,
			AccountPermissionsMap: helper.examplePerms,
		}

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, helper.examplePerms, nil)
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

		var si *types.RequestContext
		require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&si))

		assert.Equal(t, expectedOutput, si)

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with invalid request time", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: 1,
		}

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

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

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
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

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAPIClient.ClientID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		helper.service.apiClientManager = dcm

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

		mock.AssertExpectationsForObjects(t, dcm)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.userDataManager = udb

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

		mock.AssertExpectationsForObjects(t, dcm, udb)
	})

	T.Run("with error fetching account memberships", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(uint64(0), map[uint64]permissions.ServiceUserPermissions(nil), errors.New("blah"))
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

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})

	T.Run("with invalid checksum", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = dcm

		// set HMAC signature
		mac := hmac.New(sha256.New, helper.exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write([]byte("lol"))
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		helper.req.Header.Set(signatureHeaderKey, sigHeader)

		helper.service.PASETOHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, dcm)
	})

	T.Run("with token encryption error", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.config.PASETO.LocalModeKey = nil

		exampleInput := &types.PASETOCreationInput{
			ClientID:    helper.exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		helper.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
		helper.req = helper.req.WithContext(helper.ctx)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAPIClient.ClientID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleAccount.ID, helper.examplePerms, nil)
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

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})
}

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
	"net/http/httptest"
	"testing"
	"time"

	"github.com/o1egl/paseto"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/bcrypt"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testURL = "http://todo.verygoodsoftwarenotvirus.ru/testing"
)

func attachCookieToRequestForTest(t *testing.T, s *service, req *http.Request, user *types.User) (context.Context, *http.Request) {
	t.Helper()

	exampleAccount := fakes.BuildFakeAccount()
	examplePerms := map[uint64]bitmask.ServiceUserPermissions{
		exampleAccount.ID: testutil.BuildMaxUserPerms(),
	}

	ctx, sessionErr := s.sessionManager.Load(req.Context(), "")
	require.NoError(t, sessionErr)
	require.NoError(t, s.sessionManager.RenewToken(ctx))

	// Then make the privilege-level change.
	s.sessionManager.Put(ctx, sessionInfoKey, types.RequestContextFromUser(user, exampleAccount.ID, examplePerms))

	token, _, err := s.sessionManager.Commit(ctx)
	assert.NotEmpty(t, token)
	assert.NoError(t, err)

	c, err := s.buildCookie(token, time.Now().Add(s.config.Cookies.Lifetime))
	require.NoError(t, err)
	req.AddCookie(c)

	return ctx, req.WithContext(ctx)
}

func TestService_DecodeCookieFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		ctx, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		cookie, err := s.DecodeCookieFromRequest(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, cookie)
	})

	T.Run("with invalid cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		// begin building bad cookie.
		// NOTE: any code here is duplicated from service.buildAuthCookie
		// any changes made there might need to be reflected here.
		c := &http.Cookie{
			Name:     s.config.Cookies.Name,
			Value:    "blah blah blah this is not a real cookie",
			Path:     "/",
			HttpOnly: true,
		}
		// end building bad cookie.
		req.AddCookie(c)

		cookie, err := s.DecodeCookieFromRequest(req.Context(), req)
		assert.Error(t, err)
		assert.Nil(t, cookie)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		cookie, err := s.DecodeCookieFromRequest(req.Context(), req)
		assert.Error(t, err)
		assert.Equal(t, err, http.ErrNoCookie)
		assert.Nil(t, cookie)
	})
}

func TestService_fetchUserFromCookie(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		ctx, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		actualUser, err := s.fetchUserFromCookie(ctx, req)
		assert.Equal(t, exampleUser, actualUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actualUser, err := s.fetchUserFromCookie(req.Context(), req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		expectedError := errors.New("blah")
		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return((*types.User)(nil), expectedError)
		s.userDB = udb

		actualUser, err := s.fetchUserFromCookie(req.Context(), req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, udb)
	})
}

func TestService_LoginHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx := context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.Username,
		).Return(exampleUser, nil)
		s.userDB = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = authr

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = membershipDB

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On(
			"LogSuccessfulLoginEvent",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		)
		s.auditLog = auditLog

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)
		assert.NotEmpty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr, auditLog, membershipDB)
	})

	T.Run("with error fetching login data from request", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error retrieving user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx := context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return((*types.User)(nil), errors.New("blah"))
		s.userDB = udb

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleUser := fakes.BuildFakeUser()
		exampleUser.AccountStatus = types.BannedAccountStatus
		exampleUser.AccountStatusExplanation = "bad behavior"
		exampleAccount := fakes.BuildFakeAccount()

		s.sessionInfoFetcher = func(*http.Request) (*types.RequestContext, error) {
			reqCtx := types.RequestContextFromUser(
				exampleUser,
				exampleAccount.ID,
				map[uint64]bitmask.ServiceUserPermissions{},
			)

			return reqCtx, nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx := context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)
		s.userDB = udb

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogBannedUserLoginAttemptEvent", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID)
		s.auditLog = auditLog

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, auditLog)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)
		s.userDB = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(false, nil)
		s.authenticator = authr

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogUnsuccessfulLoginBadPasswordEvent", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID)
		s.auditLog = auditLog

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)
		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr, auditLog)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)
		s.userDB = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, errors.New("blah"))
		s.authenticator = authr

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)
		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",

			s.config.Cookies.Name,
			mock.AnythingOfType("string"),
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)
		s.userDB = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = authr

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.Anything,
			exampleUser.ID,
		).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = membershipDB

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)
		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, udb, authr, membershipDB)
	})

	T.Run("with error building cookie and error encoding cookie response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			s.config.Cookies.Name,
			mock.AnythingOfType("string"),
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)
		s.userDB = udb

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = authr

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.Anything,
			exampleUser.ID,
		).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = membershipDB

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)
		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, udb, authr, membershipDB)
	})
}

func TestService_LogoutHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogLogoutEvent", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID)
		s.auditLog = auditLog

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		res := httptest.NewRecorder()

		s.LogoutHandler(res, req)

		actualCookie := res.Header().Get("Set-Cookie")
		assert.Contains(t, actualCookie, "Max-Age=0")

		mock.AssertExpectationsForObjects(t, auditLog)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		ctx, sessionErr := s.sessionManager.Load(ctx, "")
		require.NoError(t, sessionErr)
		require.NoError(t, s.sessionManager.RenewToken(ctx))

		// Then make the privilege-level change.
		s.sessionManager.Put(ctx, sessionInfoKey, types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		s.LogoutHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)
		s.cookieManager = securecookie.New(
			securecookie.GenerateRandomKey(0),
			[]byte(""),
		)

		res := httptest.NewRecorder()

		s.LogoutHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}

func TestService_validateLogin(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}
		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = authr

		actual, err := s.validateLogin(ctx, exampleUser, exampleLoginData)
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})

	T.Run("with too weak a authentication hash", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}
		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, authentication.ErrPasswordHashTooWeak)
		s.authenticator = authr

		authr.On(
			"HashPassword",
			mock.Anything,
			exampleLoginData.Password,
		).Return("blah", nil)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"UpdateUser",
			mock.Anything,
			mock.IsType(&types.User{}),
		).Return(nil)
		s.userDB = udb

		actual, err := s.validateLogin(ctx, exampleUser, exampleLoginData)
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authr, udb)
	})

	T.Run("with too weak a authentication hash and error authentication the authentication", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		expectedErr := errors.New("arbitrary")

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}
		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, authentication.ErrPasswordHashTooWeak)

		authr.On(
			"HashPassword",
			mock.Anything,
			exampleLoginData.Password,
		).Return("", expectedErr)
		s.authenticator = authr

		actual, err := s.validateLogin(ctx, exampleUser, exampleLoginData)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})

	T.Run("with too weak a authentication hash and error updating user", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		expectedErr := errors.New("arbitrary")

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}
		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(true, bcrypt.ErrCostTooLow)

		authr.On(
			"HashPassword",
			mock.Anything,
			exampleLoginData.Password,
		).Return("blah", nil)
		s.authenticator = authr

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"UpdateUser",
			mock.Anything,
			mock.IsType(&types.User{}),
		).Return(expectedErr)
		s.userDB = udb

		actual, err := s.validateLogin(ctx, exampleUser, exampleLoginData)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authr, udb)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		expectedErr := errors.New("arbitrary")

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}
		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(false, expectedErr)
		s.authenticator = authr

		actual, err := s.validateLogin(ctx, exampleUser, exampleLoginData)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}
		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		authr := &mockauth.Authenticator{}
		authr.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleLoginData.Password,
			exampleUser.TwoFactorSecret,
			exampleLoginData.TOTPToken,
			exampleUser.Salt,
		).Return(false, nil)
		s.authenticator = authr

		actual, err := s.validateLogin(ctx, exampleUser, exampleLoginData)
		assert.False(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authr)
	})
}

func TestService_StatusHandler(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		s.StatusHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		s.userDB = udb

		s.StatusHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})
}

func TestService_CycleSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogCycleCookieSecretEvent", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID)
		s.auditLog = auditLog

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)
		c := req.Cookies()[0]

		var token string
		assert.NoError(t, s.cookieManager.Decode(s.config.Cookies.Name, c.Value, &token))

		s.CycleCookieSecretHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, res.Code)
		assert.Error(t, s.cookieManager.Decode(s.config.Cookies.Name, c.Value, &token))

		mock.AssertExpectationsForObjects(t, auditLog)
	})

	T.Run("with error retrieving session info", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)
		c := req.Cookies()[0]

		var token string
		assert.NoError(t, s.cookieManager.Decode(s.config.Cookies.Name, c.Value, &token))

		s.CycleCookieSecretHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, res.Code)
		assert.NoError(t, s.cookieManager.Decode(s.config.Cookies.Name, c.Value, &token))
	})

	T.Run("with invalid permissions", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)
		c := req.Cookies()[0]

		var token string
		assert.NoError(t, s.cookieManager.Decode(s.config.Cookies.Name, c.Value, &token))

		s.CycleCookieSecretHandler(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, res.Code)
		assert.NoError(t, s.cookieManager.Decode(s.config.Cookies.Name, c.Value, &token))
	})
}

func TestService_buildCookie(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		cookie, err := s.buildCookie("example", time.Now().Add(s.config.Cookies.Lifetime))
		assert.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	T.Run("with erroneous cookie building setup", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)
		s.cookieManager = securecookie.New(
			securecookie.GenerateRandomKey(0),
			[]byte(""),
		)

		cookie, err := s.buildCookie("example", time.Now().Add(s.config.Cookies.Lifetime))
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func TestService_PASETOHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		s.config.PASETO.Lifetime = time.Minute

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		expectedOutput := &types.RequestContext{
			User: types.UserRequestContext{
				Username:                exampleUser.Username,
				ID:                      exampleUser.ID,
				ActiveAccountID:         exampleAccount.ID,
				UserAccountStatus:       exampleUser.AccountStatus,
				AccountPermissionsMap:   examplePerms,
				ServiceAdminPermissions: exampleUser.ServiceAdminPermissions,
			},
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleAPIClient.ClientID,
		).Return(exampleAPIClient, nil)
		s.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = membershipDB

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		// validate results

		var result *types.PASETOResponse
		require.NoError(t, json.NewDecoder(res.Body).Decode(&result))

		assert.NotEmpty(t, result.Token)

		var targetPayload paseto.JSONToken
		require.NoError(t, paseto.NewV2().Decrypt(result.Token, s.config.PASETO.LocalModeKey, &targetPayload, nil))

		assert.True(t, targetPayload.Expiration.After(time.Now().UTC()))

		payload := targetPayload.Get(pasetoDataKey)

		gobEncoding, err := base64.RawURLEncoding.DecodeString(payload)
		require.NoError(t, err)

		var si *types.RequestContext
		require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&si))

		assert.Equal(t, expectedOutput, si)

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})

	T.Run("never issues token that has a longer lifetime than package maximum", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
		s.config.PASETO.Lifetime = 24 * time.Hour * 365 // one year

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		expectedOutput := &types.RequestContext{
			User: types.UserRequestContext{
				Username:                exampleUser.Username,
				ID:                      exampleUser.ID,
				ActiveAccountID:         exampleAccount.ID,
				UserAccountStatus:       exampleUser.AccountStatus,
				AccountPermissionsMap:   examplePerms,
				ServiceAdminPermissions: exampleUser.ServiceAdminPermissions,
			},
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleAPIClient.ClientID,
		).Return(exampleAPIClient, nil)
		s.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = membershipDB

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)

		// validate results

		var result *types.PASETOResponse
		require.NoError(t, json.NewDecoder(res.Body).Decode(&result))

		assert.NotEmpty(t, result.Token)

		var targetPayload paseto.JSONToken
		require.NoError(t, paseto.NewV2().Decrypt(result.Token, s.config.PASETO.LocalModeKey, &targetPayload, nil))

		assert.True(t, targetPayload.Expiration.Before(time.Now().UTC().Add(maxPASETOLifetime)))

		payload := targetPayload.Get(pasetoDataKey)

		gobEncoding, err := base64.RawURLEncoding.DecodeString(payload)
		require.NoError(t, err)

		var si *types.RequestContext
		require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&si))

		assert.Equal(t, expectedOutput, si)

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})

	T.Run("missing input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		ctx := context.Background()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with invalid request time", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: 1,
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("error decoding signature header", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base32.HexEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("error fetching API client", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleAPIClient.ClientID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		s.apiClientManager = dcm

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, dcm)
	})

	T.Run("error fetching user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleAPIClient.ClientID,
		).Return(exampleAPIClient, nil)
		s.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		s.userDB = udb

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, dcm, udb)
	})

	T.Run("error fetching account memberships", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleAPIClient.ClientID,
		).Return(exampleAPIClient, nil)
		s.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(uint64(0), map[uint64]bitmask.ServiceUserPermissions(nil), errors.New("blah"))
		s.accountMembershipManager = membershipDB

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})

	T.Run("invalid checksum", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleAPIClient.ClientID,
		).Return(exampleAPIClient, nil)
		s.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = membershipDB

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write([]byte("lol"))
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})

	T.Run("token encryption error", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.config.PASETO.LocalModeKey = nil

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.BelongsToUser = exampleUser.ID

		s.sessionInfoFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			return types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms), nil
		}

		exampleInput := &types.PASETOCreationInput{
			ClientID:    exampleAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		ctx := context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)

		dcm := &mocktypes.APIClientDataManager{}
		dcm.On(
			"GetAPIClientByClientID",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleAPIClient.ClientID,
		).Return(exampleAPIClient, nil)
		s.apiClientManager = dcm

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		membershipDB := &mocktypes.AccountUserMembershipDataManager{}
		membershipDB.On(
			"GetMembershipsForUser",
			mock.MatchedBy(testutil.ContextMatcher),
			exampleUser.ID,
		).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = membershipDB

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		var bodyBytes bytes.Buffer
		marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
		require.NoError(t, marshalErr)

		// set HMAC signature
		mac := hmac.New(sha256.New, exampleAPIClient.ClientSecret)
		_, macWriteErr := mac.Write(bodyBytes.Bytes())
		require.NoError(t, macWriteErr)

		sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Set(signatureHeaderKey, sigHeader)

		s.PASETOHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
	})
}

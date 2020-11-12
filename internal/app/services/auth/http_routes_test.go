package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func attachCookieToRequestForTest(t *testing.T, s *Service, req *http.Request, user *types.User) (context.Context, *http.Request) {
	t.Helper()

	ctx, sessionErr := s.sessionManager.Load(req.Context(), "")
	require.NoError(t, sessionErr)
	require.NoError(t, s.sessionManager.RenewToken(ctx))

	// Then make the privilege-level change.
	s.sessionManager.Put(ctx, sessionInfoKey, user.ToSessionInfo())

	token, _, err := s.sessionManager.Commit(ctx)
	assert.NotEmpty(t, token)
	assert.NoError(t, err)

	c, err := s.buildCookie(token, time.Now().Add(s.config.CookieLifetime))
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

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/something", nil)
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/something", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		// begin building bad cookie.
		// NOTE: any code here is duplicated from service.buildAuthCookie
		// any changes made there might need to be reflected here.
		c := &http.Cookie{
			Name:     CookieName,
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/something", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		cookie, err := s.DecodeCookieFromRequest(req.Context(), req)
		assert.Error(t, err)
		assert.Equal(t, err, http.ErrNoCookie)
		assert.Nil(t, cookie)
	})
}

func TestService_WebsocketAuthFunction(T *testing.T) {
	T.Parallel()

	T.Run("with valid oauth2 client", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		oacv := &mockOAuth2ClientValidator{}
		oacv.On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = oacv

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)

		mock.AssertExpectationsForObjects(t, oacv)
	})

	T.Run("with valid cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		oacv := &mockOAuth2ClientValidator{}
		oacv.On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return(exampleOAuth2Client, errors.New("blah"))
		s.oauth2ClientsService = oacv

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)

		mock.AssertExpectationsForObjects(t, oacv)
	})

	T.Run("with nothing", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		oacv := &mockOAuth2ClientValidator{}
		oacv.On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return(exampleOAuth2Client, errors.New("blah"))
		s.oauth2ClientsService = oacv

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual := s.WebsocketAuthFunction(req)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, oacv)
	})
}

func TestService_fetchUserFromCookie(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		ctx, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		udb := &mockmodels.UserDataManager{}
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
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

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		expectedError := errors.New("blah")
		udb := &mockmodels.UserDataManager{}
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
		exampleUser := fakes.BuildFakeUser()

		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx := context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mockmodels.UserDataManager{}
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

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogSuccessfulLoginEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)
		assert.NotEmpty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr)
	})

	T.Run("with error fetching login data from request", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mockmodels.UserDataManager{}
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

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUnsuccessfulLoginBadPasswordEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)
		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, udb, authr)
	})

	T.Run("with error validating login", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		udb := &mockmodels.UserDataManager{}
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
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

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			CookieName,
			mock.AnythingOfType("string"),
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		udb := &mockmodels.UserDataManager{}
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)
		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, udb, authr)
	})

	T.Run("with error building cookie and error encoding cookie response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		exampleLoginData := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		ctx = context.WithValue(ctx, userLoginInputMiddlewareCtxKey, exampleLoginData)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			CookieName,
			mock.AnythingOfType("string"),
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		udb := &mockmodels.UserDataManager{}
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)
		res := httptest.NewRecorder()

		s.LoginHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))

		mock.AssertExpectationsForObjects(t, cb, udb, authr)
	})
}

func TestService_LogoutHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogLogoutEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		res := httptest.NewRecorder()

		s.LogoutHandler(res, req)

		actualCookie := res.Header().Get("Set-Cookie")
		assert.Contains(t, actualCookie, "Max-Age=0")
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		s.LogoutHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error building cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
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

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
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

	T.Run("with too weak a password hash", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
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
		).Return(true, auth.ErrPasswordHashTooWeak)
		s.authenticator = authr

		authr.On(
			"HashPassword",
			mock.Anything,
			exampleLoginData.Password,
		).Return("blah", nil)

		udb := &mockmodels.UserDataManager{}
		udb.On(
			"UpdateUser",
			mock.Anything,
			mock.AnythingOfType("*types.User"),
		).Return(nil)
		s.userDB = udb

		actual, err := s.validateLogin(ctx, exampleUser, exampleLoginData)
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authr, udb)
	})

	T.Run("with too weak a password hash and error hashing the password", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		expectedErr := errors.New("arbitrary")

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
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
		).Return(true, auth.ErrPasswordHashTooWeak)

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

	T.Run("with too weak a password hash and error updating user", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		expectedErr := errors.New("arbitrary")
		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
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
		).Return(true, auth.ErrCostTooLow)

		authr.On(
			"HashPassword",
			mock.Anything,
			exampleLoginData.Password,
		).Return("blah", nil)
		s.authenticator = authr

		udb := &mockmodels.UserDataManager{}
		udb.On(
			"UpdateUser",
			mock.Anything,
			mock.AnythingOfType("*types.User"),
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
		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
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

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
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

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		udb := &mockmodels.UserDataManager{}
		udb.On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return(exampleUser, nil)
		s.userDB = udb

		s.StatusHandler(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		udb := &mockmodels.UserDataManager{}
		udb.On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		s.userDB = udb

		s.StatusHandler(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})
}

func TestService_CycleSecretHandler(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: false}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogCycleCookieSecretEvent", mock.Anything, exampleUser.ID)
		s.auditLog = auditLog

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)
		c := req.Cookies()[0]

		var token string
		assert.NoError(t, s.cookieManager.Decode(CookieName, c.Value, &token))

		s.CycleSecretHandler(res, req)

		assert.Error(t, s.cookieManager.Decode(CookieName, c.Value, &token))
	})
}

func TestService_buildCookie(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		cookie, err := s.buildCookie("example", time.Now().Add(s.config.CookieLifetime))
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

		cookie, err := s.buildCookie("example", time.Now().Add(s.config.CookieLifetime))
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})
}

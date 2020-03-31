package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth/mock"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_DecodeCookieFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/something", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(exampleUser)
		require.NoError(t, err)
		req.AddCookie(c)

		cookie, err := s.DecodeCookieFromRequest(req.Context(), req)
		assert.NoError(t, err)
		assert.NotNil(t, cookie)
	})

	T.Run("with invalid cookie", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/something", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		// begin building bad cookie
		// NOTE: any code here is duplicated from service.buildAuthCookie
		// any changes made there might need to be reflected here
		c := &http.Cookie{
			Name:     CookieName,
			Value:    "blah blah blah this is not a real cookie",
			Path:     "/",
			HttpOnly: true,
		}
		// end building bad cookie
		req.AddCookie(c)

		cookie, err := s.DecodeCookieFromRequest(req.Context(), req)
		assert.Error(t, err)
		assert.Nil(t, cookie)
	})

	T.Run("without cookie", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/something", nil)
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
		s := buildTestService(t)

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		s.oauth2ClientsService.(*mockOAuth2ClientValidator).On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
			mock.Anything,
		).Return(exampleOAuth2Client, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)
	})

	T.Run("with valid cookie", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()
		s.oauth2ClientsService.(*mockOAuth2ClientValidator).On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
			mock.Anything,
		).Return(exampleOAuth2Client, errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(exampleUser)
		require.NoError(t, err)
		req.AddCookie(c)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)
	})

	T.Run("with nothing", func(t *testing.T) {
		s := buildTestService(t)

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()
		s.oauth2ClientsService.(*mockOAuth2ClientValidator).On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
			mock.Anything,
		).Return(exampleOAuth2Client, errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual := s.WebsocketAuthFunction(req)
		assert.False(t, actual)
	})
}

func TestService_FetchUserFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(exampleUser)
		require.NoError(t, err)
		req.AddCookie(c)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return(exampleUser, nil)

		actualUser, err := s.FetchUserFromRequest(req.Context(), req)
		assert.Equal(t, exampleUser, actualUser)
		assert.NoError(t, err)
	})

	T.Run("without cookie", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return(exampleUser, nil)

		actualUser, err := s.FetchUserFromRequest(req.Context(), req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(exampleUser)
		require.NoError(t, err)
		req.AddCookie(c)

		expectedError := errors.New("blah")
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUser",
			mock.Anything,
			exampleUser.ID,
		).Return((*models.User)(nil), expectedError)

		actualUser, err := s.FetchUserFromRequest(req.Context(), req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})
}

func TestService_LoginHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)
		assert.NotEmpty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error fetching login data from request", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, errors.New("arbitrary"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error encoding error fetching login data", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"EncodeResponse",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("blah"))
		s.encoderDecoder = ed

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, errors.New("arbitrary"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with invalid login", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(false, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error building cookie", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			mock.Anything,
			mock.Anything,
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error building cookie and error encoding cookie response", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			mock.Anything,
			mock.Anything,
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"EncodeResponse",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("blah"))
		s.encoderDecoder = ed

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})
}

func TestService_Logout(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(exampleUser)
		require.NoError(t, err)

		req.AddCookie(c)
		res := httptest.NewRecorder()

		s.LogoutHandler()(res, req)

		actualCookie := res.Header().Get("Set-Cookie")
		assert.Contains(t, actualCookie, "Max-Age=0")
	})

	T.Run("without cookie", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		s.LogoutHandler()(res, req)
	})
}

func TestService_fetchLoginDataFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return(exampleUser, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))
		loginData, err := s.fetchLoginDataFromRequest(req)

		require.NotNil(t, loginData)
		assert.Equal(t, loginData.user, exampleUser)
		assert.Nil(t, err)
	})

	T.Run("without login data attached to request", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		_, err = s.fetchLoginDataFromRequest(req)
		assert.Error(t, err)
	})

	T.Run("with DB error fetching user", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return((*models.User)(nil), sql.ErrNoRows)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))
		_, err = s.fetchLoginDataFromRequest(req)
		assert.Error(t, err)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			exampleUser.Username,
		).Return((*models.User)(nil), errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))
		_, err = s.fetchLoginDataFromRequest(req)
		assert.Error(t, err)
	})
}

func TestService_validateLogin(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)
		exampleInput := loginData{
			loginInput: exampleLoginData,
			user:       exampleUser,
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.True(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with too weak a password hash", func(t *testing.T) {
		s := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)
		exampleInput := loginData{
			loginInput: exampleLoginData,
			user:       exampleUser,
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, auth.ErrPasswordHashTooWeak)

		s.authenticator.(*mockauth.Authenticator).On(
			"HashPassword",
			mock.Anything,
			mock.Anything,
		).Return("blah", nil)

		s.userDB.(*mockmodels.UserDataManager).On(
			"UpdateUser",
			mock.Anything,
			mock.Anything,
		).Return(nil)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.True(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with too weak a password hash and error hashing the password", func(t *testing.T) {
		s := buildTestService(t)
		ctx := context.Background()
		expectedErr := errors.New("arbitrary")

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)
		exampleInput := loginData{
			loginInput: exampleLoginData,
			user:       exampleUser,
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, auth.ErrPasswordHashTooWeak)

		s.authenticator.(*mockauth.Authenticator).On(
			"HashPassword",
			mock.Anything,
			mock.Anything,
		).Return("", expectedErr)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with too weak a password hash and error updating user", func(t *testing.T) {
		s := buildTestService(t)
		expectedErr := errors.New("arbitrary")
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)
		exampleInput := loginData{
			loginInput: exampleLoginData,
			user:       exampleUser,
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, auth.ErrPasswordHashTooWeak)

		s.authenticator.(*mockauth.Authenticator).On(
			"HashPassword",
			mock.Anything,
			mock.Anything,
		).Return("blah", nil)

		s.userDB.(*mockmodels.UserDataManager).On(
			"UpdateUser",
			mock.Anything,
			mock.Anything,
		).Return(expectedErr)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)
		ctx := context.Background()
		expectedErr := errors.New("arbitrary")

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)
		exampleInput := loginData{
			loginInput: exampleLoginData,
			user:       exampleUser,
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(false, expectedErr)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid login", func(t *testing.T) {
		s := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleLoginData := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)
		exampleInput := loginData{
			loginInput: exampleLoginData,
			user:       exampleUser,
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(false, nil)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.False(t, actual)
		assert.NoError(t, err)
	})
}

func TestService_buildCookie(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		cookie, err := s.buildAuthCookie(exampleUser)
		assert.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	T.Run("with error encoding", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			mock.Anything,
			mock.Anything,
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		cookie, err := s.buildAuthCookie(exampleUser)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func TestService_CycleSecret(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		c, err := s.buildAuthCookie(exampleUser)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		var ca models.CookieAuth
		decodeErr := s.cookieManager.Decode(CookieName, c.Value, &ca)
		assert.NoError(t, decodeErr)

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.CycleSecretHandler()(res, req)

		decodeErr2 := s.cookieManager.Decode(CookieName, c.Value, &ca)
		assert.Error(t, decodeErr2)
	})
}

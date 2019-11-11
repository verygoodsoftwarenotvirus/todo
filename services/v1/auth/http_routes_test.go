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
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_DecodeCookieFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/something", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(&models.User{ID: 1, Username: "username"})
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

		expected := &models.OAuth2Client{}
		s.oauth2ClientsService.(*mockOAuth2ClientValidator).On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
		).Return(expected, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)
	})

	T.Run("with valid cookie", func(t *testing.T) {
		s := buildTestService(t)

		oac := &models.OAuth2Client{}
		s.oauth2ClientsService.(*mockOAuth2ClientValidator).On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
		).Return(oac, errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(&models.User{ID: 1, Username: "username"})
		require.NoError(t, err)
		req.AddCookie(c)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)
	})

	T.Run("with nothing", func(t *testing.T) {
		s := buildTestService(t)

		oac := &models.OAuth2Client{}
		s.oauth2ClientsService.(*mockOAuth2ClientValidator).On(
			"ExtractOAuth2ClientFromRequest",
			mock.Anything,
		).Return(oac, errors.New("blah"))

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

		userID := uint64(1)
		expectedUser := &models.User{ID: userID, Username: "username"}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(expectedUser)
		require.NoError(t, err)
		req.AddCookie(c)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUser",
			mock.Anything,
			userID,
		).Return(expectedUser, nil)

		actualUser, err := s.FetchUserFromRequest(req.Context(), req)
		assert.Equal(t, expectedUser, actualUser)
		assert.NoError(t, err)
	})

	T.Run("without cookie", func(t *testing.T) {
		s := buildTestService(t)

		userID := uint64(1)
		expectedUser := &models.User{ID: userID, Username: "username"}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUser",
			mock.Anything,
			userID,
		).Return(expectedUser, nil)

		actualUser, err := s.FetchUserFromRequest(req.Context(), req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		s := buildTestService(t)

		userID := uint64(1)
		expectedUser := &models.User{ID: userID, Username: "username"}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(expectedUser)
		require.NoError(t, err)
		req.AddCookie(c)

		expectedError := errors.New("blah")
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUser",
			mock.Anything,
			userID,
		).Return((*models.User)(nil), expectedError)

		actualUser, err := s.FetchUserFromRequest(req.Context(), req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})
}

func TestService_Login(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusNoContent)
		assert.NotEmpty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error fetching login data from request", func(t *testing.T) {
		s := buildTestService(t)

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, errors.New("arbitrary"))

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusUnauthorized)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error encoding error fetching login data", func(t *testing.T) {
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"EncodeResponse",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("blah"))
		s.encoderDecoder = ed

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, errors.New("arbitrary"))

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusUnauthorized)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with invalid login", func(t *testing.T) {
		s := buildTestService(t)

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(false, nil)

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusUnauthorized)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, errors.New("blah"))

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusUnauthorized)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error building cookie", func(t *testing.T) {
		s := buildTestService(t)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			mock.Anything,
			mock.Anything,
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error building cookie and error encoding cookie response", func(t *testing.T) {
		s := buildTestService(t)

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

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, nil)

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))

		s.LoginHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})
}

func TestService_Logout(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildAuthCookie(&models.User{ID: 1, Username: "username"})
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

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return(expectedUser, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))
		loginData, err := s.fetchLoginDataFromRequest(req)

		require.NotNil(t, loginData)
		assert.Equal(t, loginData.user, expectedUser)
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

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return((*models.User)(nil), sql.ErrNoRows)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))
		_, err = s.fetchLoginDataFromRequest(req)
		assert.Error(t, err)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		s := buildTestService(t)

		expectedUser := &models.User{Username: "username"}
		s.userDB.(*mockmodels.UserDataManager).On(
			"GetUserByUsername",
			mock.Anything,
			expectedUser.Username,
		).Return((*models.User)(nil), errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleLoginData := &models.UserLoginInput{
			Username:  expectedUser.Username,
			Password:  "password",
			TOTPToken: "123456",
		}

		req = req.WithContext(context.WithValue(req.Context(), UserLoginInputMiddlewareCtxKey, exampleLoginData))
		_, err = s.fetchLoginDataFromRequest(req)
		assert.Error(t, err)
	})
}

func TestService_validateLogin(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		expected := true
		ctx := context.Background()
		exampleInput := loginData{
			loginInput: &models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "012345",
			},
			user: &models.User{},
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(expected, nil)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})

	T.Run("with too weak a password hash", func(t *testing.T) {
		s := buildTestService(t)

		expected := true
		ctx := context.Background()
		exampleInput := loginData{
			loginInput: &models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "012345",
			},
			user: &models.User{},
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(expected, auth.ErrPasswordHashTooWeak)

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
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})

	T.Run("with too weak a password hash and error hashing the password", func(t *testing.T) {
		s := buildTestService(t)

		expected := false
		expectedErr := errors.New("arbitrary")
		ctx := context.Background()
		exampleInput := loginData{
			loginInput: &models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "012345",
			},
			user: &models.User{},
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
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})

	T.Run("with too weak a password hash and error updating user", func(t *testing.T) {
		s := buildTestService(t)

		expected := false
		expectedErr := errors.New("arbitrary")
		ctx := context.Background()
		exampleInput := loginData{
			loginInput: &models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "012345",
			},
			user: &models.User{},
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
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)

		ctx := context.Background()
		expected := false
		expectedErr := errors.New("arbitrary")
		exampleInput := loginData{
			loginInput: &models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "012345",
			},
			user: &models.User{},
		}

		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(expected, expectedErr)

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid login", func(t *testing.T) {
		s := buildTestService(t)

		expected := false
		s.authenticator.(*mockauth.Authenticator).On(
			"ValidateLogin",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(expected, nil)

		ctx := context.Background()
		exampleInput := loginData{
			loginInput: &models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "012345",
			},
			user: &models.User{},
		}

		actual, err := s.validateLogin(ctx, exampleInput)
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})
}

func TestService_buildCookie(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleInput := &models.User{Username: "username"}
		cookie, err := s.buildAuthCookie(exampleInput)
		assert.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	T.Run("with error encoding", func(t *testing.T) {
		s := buildTestService(t)

		cb := &mockCookieEncoderDecoder{}
		cb.On(
			"Encode",
			mock.Anything,
			mock.Anything,
		).Return("", errors.New("blah"))
		s.cookieManager = cb

		exampleInput := &models.User{Username: "username"}
		cookie, err := s.buildAuthCookie(exampleInput)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func TestService_CycleSecret(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := &models.User{ID: 123, Username: "username"}
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

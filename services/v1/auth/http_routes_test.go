package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	mauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	logger := noop.ProvideNoopLogger()
	cfg := &config.ServerConfig{
		Auth: config.AuthSettings{
			CookieSecret: "BLAHBLAHBLAHPRETENDTHISISSECRET!",
		},
	}
	auth := &mauth.Authenticator{}
	userDB := &mmodels.UserDataManager{}
	oauth := &mockOAuth2ClientValidator{}
	userIDFetcher := func(*http.Request) uint64 {
		return 1
	}
	ed := encoding.ProvideResponseEncoder()

	service := ProvideAuthService(
		logger,
		cfg,
		auth,
		userDB,
		oauth,
		userIDFetcher,
		ed,
	)

	return service
}

func TestService_DecodeCookieFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/items", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildCookie(&models.User{ID: 1, Username: "username"})
		require.NoError(t, err)
		req.AddCookie(c)

		cookie, err := s.DecodeCookieFromRequest(req)
		assert.NoError(t, err)
		assert.NotNil(t, cookie)
	})

	T.Run("with invalid cookie", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/items", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		// begin building bad cookie
		// NOTE: any code here is duplicated from service.buildCookie
		// any changes made there might need to be reflected here
		c := &http.Cookie{
			Name:     CookieName,
			Value:    "blah blah blah this is not a real cookie",
			Path:     "/",
			HttpOnly: true,
		}
		// end building bad cookie

		req.AddCookie(c)
		cookie, err := s.DecodeCookieFromRequest(req)
		assert.Error(t, err)
		assert.Nil(t, cookie)
	})

	T.Run("without cookie", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/api/v1/items", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		cookie, err := s.DecodeCookieFromRequest(req)
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

		s.oauth2ClientsService.(*mockOAuth2ClientValidator).
			On("RequestIsAuthenticated", mock.Anything).
			Return(expected, nil)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)
	})

	T.Run("with valid cookie", func(t *testing.T) {
		s := buildTestService(t)
		oac := &models.OAuth2Client{}
		s.oauth2ClientsService.(*mockOAuth2ClientValidator).
			On("RequestIsAuthenticated", mock.Anything).
			Return(oac, errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildCookie(&models.User{ID: 1, Username: "username"})
		require.NoError(t, err)
		req.AddCookie(c)

		actual := s.WebsocketAuthFunction(req)
		assert.True(t, actual)
	})

	T.Run("with nothing", func(t *testing.T) {
		s := buildTestService(t)
		oac := &models.OAuth2Client{}
		s.oauth2ClientsService.(*mockOAuth2ClientValidator).
			On("RequestIsAuthenticated", mock.Anything).
			Return(oac, errors.New("blah"))

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual := s.WebsocketAuthFunction(req)
		assert.False(t, actual)
	})

}

func TestService_FetchUserFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)

		expectedUser := &models.User{
			ID:       userID,
			Username: "username",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildCookie(expectedUser)
		require.NoError(t, err)
		req.AddCookie(c)

		s.database.(*mmodels.UserDataManager).
			On("GetUser", mock.Anything, userID).
			Return(expectedUser, nil)

		actualUser, err := s.FetchUserFromRequest(req)
		assert.Equal(t, expectedUser, actualUser)
		assert.NoError(t, err)
	})

	T.Run("without cookie", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)

		expectedUser := &models.User{
			ID:       userID,
			Username: "username",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.database.(*mmodels.UserDataManager).
			On("GetUser", mock.Anything, userID).
			Return(expectedUser, nil)

		actualUser, err := s.FetchUserFromRequest(req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)

		expectedUser := &models.User{
			ID:       userID,
			Username: "username",
		}

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildCookie(expectedUser)
		require.NoError(t, err)
		req.AddCookie(c)

		expectedError := errors.New("blah")

		s.database.(*mmodels.UserDataManager).
			On("GetUser", mock.Anything, userID).
			Return((*models.User)(nil), expectedError)

		actualUser, err := s.FetchUserFromRequest(req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})
}

//func TestService_Login(T *testing.T) {
//	T.Parallel()
//
//	T.Run("normal operation", func(t *testing.T) {
//		s := buildTestService(t)
//		userID := uint64(1)
//
//		expectedUser := &models.User{
//			ID:       userID,
//			Username: "username",
//		}
//
//		s.database.(*mmodels.UserDataManager).
//			On("GetUser", mock.Anything, userID).
//			Return(expectedUser, nil)
//	})
//}

func TestService_Logout(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c, err := s.buildCookie(&models.User{ID: 1, Username: "username"})
		require.NoError(t, err)
		req.AddCookie(c)

		res := httptest.NewRecorder()
		s.Logout(res, req)

		actualCookie := res.Header().Get("Set-Cookie")
		assert.Contains(t, actualCookie, "Max-Age=0")
	})

	T.Run("without cookie", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru/testing", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		s.Logout(res, req)
	})
}

func TestService_fetchLoginDataFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)

		expectedUser := &models.User{
			Username: "username",
		}

		s.database.(*mmodels.UserDataManager).
			On("GetUserByUsername", mock.Anything, expectedUser.Username).
			Return(expectedUser, nil)

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

		expectedUser := &models.User{
			Username: "username",
		}

		s.database.(*mmodels.UserDataManager).
			On("GetUserByUsername", mock.Anything, expectedUser.Username).
			Return((*models.User)(nil), sql.ErrNoRows)

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

		expectedUser := &models.User{
			Username: "username",
		}

		s.database.(*mmodels.UserDataManager).
			On("GetUserByUsername", mock.Anything, expectedUser.Username).
			Return((*models.User)(nil), errors.New("blah"))

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

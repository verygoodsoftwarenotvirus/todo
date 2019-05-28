package auth

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
)

func TestService_CookieAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)
		exampleUser := &models.User{
			Username: "username",
		}

		md := &mmodels.UserDataManager{}
		md.On("GetUser", mock.Anything, mock.Anything).Return(exampleUser, nil)
		s.userDB = md

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		cookie, err := s.buildCookie(exampleUser)
		require.NotNil(t, cookie)
		require.NoError(t, err)
		req.AddCookie(cookie)

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)
	})

	T.Run("with nil user", func(t *testing.T) {
		s := buildTestService(t)
		exampleUser := &models.User{
			Username: "username",
		}

		md := &mmodels.UserDataManager{}
		md.On("GetUser", mock.Anything, mock.Anything).
			Return((*models.User)(nil), nil)
		s.userDB = md

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		cookie, err := s.buildCookie(exampleUser)
		require.NotNil(t, cookie)
		require.NoError(t, err)
		req.AddCookie(cookie)

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusUnauthorized)
	})

	T.Run("without user attached", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)
	})
}

func TestService_AuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService(t)

		exampleClient := &models.OAuth2Client{
			ClientID:     "PRETEND_THIS_IS_A_REAL_CLIENT_ID",
			ClientSecret: "PRETEND_THIS_IS_A_REAL_CLIENT_SECRET",
		}

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("RequestIsAuthenticated", mock.Anything).
			Return(exampleClient, nil)
		s.oauth2ClientsService = ocv

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		s.AuthenticationMiddleware(true)(h).ServeHTTP(res, req)
	})

	T.Run("with error fetching client but able to use cookie", func(t *testing.T) {
		s := buildTestService(t)

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("RequestIsAuthenticated", mock.Anything).
			Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.oauth2ClientsService = ocv

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		c, err := s.buildCookie(&models.User{ID: 1, Username: "username"})
		require.NoError(t, err)
		req.AddCookie(c)

		s.AuthenticationMiddleware(true)(h).ServeHTTP(res, req)
	})

	T.Run("with error fetching client but able to use cookie but unable to decode cookie", func(t *testing.T) {
		s := buildTestService(t)

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("RequestIsAuthenticated", mock.Anything).
			Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.oauth2ClientsService = ocv

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		c, err := s.buildCookie(&models.User{ID: 1, Username: "username"})
		require.NoError(t, err)
		req.AddCookie(c)

		cb := &mockCookieEncoderDecoder{}
		cb.On("Decode", CookieName, mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.cookieBuilder = cb

		s.AuthenticationMiddleware(true)(h).ServeHTTP(res, req)
	})

	T.Run("with invalid authentication", func(t *testing.T) {
		s := buildTestService(t)

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("RequestIsAuthenticated", mock.Anything).
			Return((*models.OAuth2Client)(nil), nil)
		s.oauth2ClientsService = ocv

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		s.AuthenticationMiddleware(false)(h).ServeHTTP(res, req)
		assert.Equal(t, res.Code, http.StatusUnauthorized)
	})
}

func Test_parseLoginInputFromForm(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		expected := &models.UserLoginInput{
			Username:  "username",
			Password:  "password",
			TOTPToken: "123456",
		}

		req.Form = map[string][]string{
			UsernameFormKey:  {expected.Username},
			PasswordFormKey:  {expected.Password},
			TOTPTokenFormKey: {expected.TOTPToken},
		}

		actual := parseLoginInputFromForm(req)
		assert.NotNil(t, actual)
		assert.Equal(t, expected, actual)
	})

	T.Run("returns nil with error parsing form", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.URL.RawQuery = "%gh&%ij"
		req.Form = nil

		actual := parseLoginInputFromForm(req)
		assert.Nil(t, actual)
	})
}

func TestService_UserLoginInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		exampleInput := &models.UserLoginInput{
			Username:  "username",
			Password:  "password",
			TOTPToken: "1233456",
		}

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleInput))

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		s := buildTestService(t)
		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)

	})

	T.Run("with error decoding request", func(t *testing.T) {
		exampleInput := &models.UserLoginInput{
			Username:  "username",
			Password:  "password",
			TOTPToken: "1233456",
		}

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleInput))

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		s := buildTestService(t)
		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)
	})

	T.Run("with error decoding request but valid value attached to form", func(t *testing.T) {
		exampleInput := &models.UserLoginInput{
			Username:  "username",
			Password:  "password",
			TOTPToken: "1233456",
		}
		form := url.Values{
			UsernameFormKey:  {exampleInput.Username},
			PasswordFormKey:  {exampleInput.Password},
			TOTPTokenFormKey: {exampleInput.TOTPToken},
		}

		req, err := http.NewRequest(
			http.MethodPost,
			"http://todo.verygoodsoftwarenotvirus.ru",
			strings.NewReader(form.Encode()),
		)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()
		req.Header.Set("Content-type", "application/x-www-form-urlencoded")

		s := buildTestService(t)
		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)
	})
}

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CookieAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()

		md := &mockmodels.UserDataManager{}
		md.On("GetUser", mock.Anything, mock.Anything).Return(exampleUser, nil)
		s.userDB = md

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, md, ms)
	})

	T.Run("with nil user", func(t *testing.T) {
		s := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()

		md := &mockmodels.UserDataManager{}
		md.On("GetUser", mock.Anything, mock.Anything).Return((*models.User)(nil), nil)
		s.userDB = md

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		ms := &MockHTTPHandler{}
		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, md, ms)
	})

	T.Run("without user attached", func(t *testing.T) {
		s := buildTestService(t)

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		ms := &MockHTTPHandler{}
		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ms)
	})
}

func TestService_AuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = ocv

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleOAuth2Client.BelongsToUser).Return(exampleUser, nil)
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		s.AuthenticationMiddleware(true)(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, mockDB, h)
	})

	T.Run("happy path without allowing cookies", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = ocv

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleOAuth2Client.BelongsToUser).Return(exampleUser, nil)
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		s.AuthenticationMiddleware(false)(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, mockDB, h)
	})

	T.Run("with error fetching client but able to use cookie", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		s.AuthenticationMiddleware(true)(h).ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, mockDB, h)
	})

	T.Run("able to use cookies but error fetching user info", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).Return((*models.User)(nil), errors.New("blah"))
		s.userDB = mockDB

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		h := &MockHTTPHandler{}
		s.AuthenticationMiddleware(true)(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, h)
	})

	T.Run("no cookies allowed, with error fetching user info", func(t *testing.T) {
		s := buildTestService(t)

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = ocv

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleOAuth2Client.BelongsToUser).Return((*models.User)(nil), errors.New("blah"))
		s.userDB = mockDB

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		h := &MockHTTPHandler{}
		s.AuthenticationMiddleware(false)(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, mockDB, h)
	})

	T.Run("with error fetching client but able to use cookie but unable to decode cookie", func(t *testing.T) {
		s := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.oauth2ClientsService = ocv

		cb := &mockCookieEncoderDecoder{}
		cb.On("Decode", CookieName, mock.Anything, mock.Anything).Return(errors.New("blah"))
		cb.On("Encode", CookieName, mock.Anything).Return("", nil)
		s.cookieManager = cb

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		h := &MockHTTPHandler{}
		s.AuthenticationMiddleware(true)(h).ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ocv, cb, h)
	})

	T.Run("with invalid authentication", func(t *testing.T) {
		s := buildTestService(t)

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return((*models.OAuth2Client)(nil), nil)
		s.oauth2ClientsService = ocv

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		h := &MockHTTPHandler{}
		s.AuthenticationMiddleware(false)(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, h)
	})

	T.Run("nightmare path", func(t *testing.T) {
		s := buildTestService(t)

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = ocv

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleOAuth2Client.BelongsToUser).Return((*models.User)(nil), nil)
		s.userDB = mockDB

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		h := &MockHTTPHandler{}
		s.AuthenticationMiddleware(false)(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, mockDB, h)
	})
}

func Test_parseLoginInputFromForm(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		exampleUser := fakemodels.BuildFakeUser()
		expected := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		req.Form = map[string][]string{
			usernameFormKey:  {expected.Username},
			passwordFormKey:  {expected.Password},
			totpTokenFormKey: {expected.TOTPToken},
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

	T.Run("happy path", func(t *testing.T) {
		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleInput))

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleInput))

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)

		s := buildTestService(t)
		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ed, ms)
	})

	T.Run("with error decoding request but valid value attached to form", func(t *testing.T) {
		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserLoginInputFromUser(exampleUser)

		form := url.Values{
			usernameFormKey:  {exampleInput.Username},
			passwordFormKey:  {exampleInput.Password},
			totpTokenFormKey: {exampleInput.TOTPToken},
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
		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ed, ms)
	})
}

func TestService_AdminMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.IsAdmin = true

		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.SessionInfoKey,
				exampleUser.ToSessionInfo(),
			),
		)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("without user attached", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.IsAdmin = false

		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				models.SessionInfoKey,
				exampleUser.ToSessionInfo(),
			),
		)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})
}

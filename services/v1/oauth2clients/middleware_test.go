package oauth2clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	oauth2models "gopkg.in/oauth2.v3/models"
)

var _ http.Handler = (*mockHTTPHandler)(nil)

type mockHTTPHandler struct {
	mock.Mock
}

func (m *mockHTTPHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

func TestService_CreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"DecodeRequest",
			mock.AnythingOfType("*http.Request"),
			mock.Anything,
		).Return(nil)
		s.encoderDecoder = ed

		mh := &mockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.Anything,
			mock.Anything,
		)

		h := s.CreationInputMiddleware(mh)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		expected := models.OAuth2ClientCreationInput{
			RedirectURI: "https://blah.com",
		}
		bs, err := json.Marshal(expected)
		require.NoError(t, err)
		req.Body = ioutil.NopCloser(bytes.NewReader(bs))

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"DecodeRequest",
			mock.AnythingOfType("*http.Request"),
			mock.Anything,
		).Return(errors.New("blah"))
		s.encoderDecoder = ed

		mh := &mockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.Anything,
			mock.Anything,
		)

		h := s.CreationInputMiddleware(mh)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
}

func TestService_RequestIsAuthenticated(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.OAuth2Client{
			ClientID: "THIS IS A FAKE CLIENT ID",
			Scopes:   []string{"things"},
		}

		mh := &mockOauth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: expected.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			expected.ClientID,
		).Return(expected, nil)
		s.database = mockDB

		req := buildRequest(t)
		req.URL.Path = "/api/v1/things"
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("with error validating token", func(t *testing.T) {
		s := buildTestService(t)

		mh := &mockOauth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return((*oauth2models.Token)(nil), errors.New("blah"))
		s.oauth2Handler = mh

		req := buildRequest(t)
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.OAuth2Client{
			ClientID: "THIS IS A FAKE CLIENT ID",
		}

		mh := &mockOauth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: expected.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			expected.ClientID,
		).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.database = mockDB

		req := buildRequest(t)
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid scope", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.OAuth2Client{
			ClientID: "THIS IS A FAKE CLIENT ID",
			Scopes:   []string{"things"},
		}

		mh := &mockOauth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: expected.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			expected.ClientID,
		).Return(expected, nil)
		s.database = mockDB

		req := buildRequest(t)
		req.URL.Path = "/api/v1/stuff"
		actual, err := s.ExtractOAuth2ClientFromRequest(req.Context(), req)

		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestService_OAuth2TokenAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	// These tests have a lot of overlap to those of ExtractOAuth2ClientFromRequest, which is deliberate

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.OAuth2Client{
			ClientID: "THIS IS A FAKE CLIENT ID",
			Scopes:   []string{"things"},
		}

		mh := &mockOauth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: expected.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			expected.ClientID,
		).Return(expected, nil)
		s.database = mockDB

		req := buildRequest(t)
		req.URL.Path = "/api/v1/things"
		res := httptest.NewRecorder()

		mhh := &mockHTTPHandler{}
		mhh.On(
			"ServeHTTP",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return()

		s.OAuth2TokenAuthenticationMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error authenticating request", func(t *testing.T) {
		s := buildTestService(t)

		mh := &mockOauth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return((*oauth2models.Token)(nil), errors.New("blah"))
		s.oauth2Handler = mh

		res := httptest.NewRecorder()
		req := buildRequest(t)

		mhh := &mockHTTPHandler{}
		mhh.On(
			"ServeHTTP",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return()

		s.OAuth2TokenAuthenticationMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

func TestService_OAuth2ClientInfoMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := "blah"

		mhh := &mockHTTPHandler{}
		mhh.On(
			"ServeHTTP",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return()

		res, req := httptest.NewRecorder(), buildRequest(t)
		q := url.Values{}
		q.Set(oauth2ClientIDURIParamKey, expected)
		req.URL.RawQuery = q.Encode()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			expected,
		).Return(&models.OAuth2Client{}, nil)
		s.database = mockDB

		s.OAuth2ClientInfoMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)
		expected := "blah"

		mhh := &mockHTTPHandler{}
		mhh.On(
			"ServeHTTP",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return()

		res, req := httptest.NewRecorder(), buildRequest(t)
		q := url.Values{}
		q.Set(oauth2ClientIDURIParamKey, expected)
		req.URL.RawQuery = q.Encode()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			expected,
		).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.database = mockDB

		s.OAuth2ClientInfoMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

func TestService_fetchOAuth2ClientFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.OAuth2Client{
			ClientID: "THIS IS A FAKE CLIENT ID",
		}

		req := buildRequest(t).WithContext(
			context.WithValue(
				context.Background(),
				models.OAuth2ClientKey,
				expected,
			),
		)

		actual := s.fetchOAuth2ClientFromRequest(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		s := buildTestService(t)
		assert.Nil(t, s.fetchOAuth2ClientFromRequest(buildRequest(t)))
	})
}

func TestService_fetchOAuth2ClientIDFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.OAuth2Client{
			ClientID: "THIS IS A FAKE CLIENT ID",
		}

		req := buildRequest(t).WithContext(
			context.WithValue(
				context.Background(),
				clientIDKey,
				expected.ClientID,
			),
		)

		actual := s.fetchOAuth2ClientIDFromRequest(req)
		assert.Equal(t, expected.ClientID, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		s := buildTestService(t)
		assert.Empty(t, s.fetchOAuth2ClientIDFromRequest(buildRequest(t)))
	})
}

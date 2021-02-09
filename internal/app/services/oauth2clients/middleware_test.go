package oauth2clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	oauth2models "github.com/go-oauth2/oauth2/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"DecodeRequest",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
			mock.Anything,
		).Return(nil)
		s.encoderDecoder = ed

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.Anything,
			mock.Anything,
		)

		h := s.CreationInputMiddleware(mh)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		expected := types.OAuth2ClientCreationInput{
			RedirectURI: "https://blah.com",
		}
		bs, err := json.Marshal(expected)
		require.NoError(t, err)
		req.Body = ioutil.NopCloser(bytes.NewReader(bs))

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"DecodeRequest", mock.Anything,
			mock.AnythingOfType("*http.Request"),
			mock.Anything,
		).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse", mock.Anything,
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		mh := &testutil.MockHTTPHandler{}
		h := s.CreationInputMiddleware(mh)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

func TestService_OAuth2TokenAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	// These tests have a lot of overlap to those of ExtractOAuth2ClientFromRequest, which is deliberate.

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mh := &mockOAuth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return(&oauth2models.Token{ClientID: exampleOAuth2Client.ClientID}, nil)
		s.oauth2Handler = mh

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		req := buildRequest(t)
		req.URL.Path = fmt.Sprintf("/api/v1/%s", exampleOAuth2Client.Scopes[0])
		res := httptest.NewRecorder()

		mhh := &testutil.MockHTTPHandler{}
		mhh.On(
			"ServeHTTP",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
		).Return()

		s.OAuth2TokenAuthenticationMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh, mhh, mockDB)
	})

	T.Run("with error authenticating request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		mh := &mockOAuth2Handler{}
		mh.On(
			"ValidationBearerToken",
			mock.AnythingOfType("*http.Request"),
		).Return((*oauth2models.Token)(nil), errors.New("blah"))
		s.oauth2Handler = mh

		res := httptest.NewRecorder()
		req := buildRequest(t)

		mhh := &testutil.MockHTTPHandler{}
		s.OAuth2TokenAuthenticationMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, mh, mhh)
	})
}

func TestService_OAuth2ClientInfoMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		expected := "blah"

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mhh := &testutil.MockHTTPHandler{}
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
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		s.OAuth2ClientInfoMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mhh, mockDB)
	})

	T.Run("with error reading from clientDataManager", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		expected := "blah"
		res, req := httptest.NewRecorder(), buildRequest(t)

		q := url.Values{}
		q.Set(oauth2ClientIDURIParamKey, expected)
		req.URL.RawQuery = q.Encode()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			expected,
		).Return((*types.OAuth2Client)(nil), errors.New("blah"))
		s.clientDataManager = mockDB

		mhh := &testutil.MockHTTPHandler{}
		s.OAuth2ClientInfoMiddleware(mhh).ServeHTTP(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, mhh, mockDB)
	})
}

func TestService_fetchOAuth2ClientFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		req := buildRequest(t).WithContext(
			context.WithValue(
				ctx,
				types.OAuth2ClientKey,
				exampleOAuth2Client,
			),
		)

		actual := s.fetchOAuth2ClientFromRequest(req)
		assert.Equal(t, exampleOAuth2Client, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		assert.Nil(t, s.fetchOAuth2ClientFromRequest(buildRequest(t)))
	})
}

func TestService_fetchOAuth2ClientIDFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		req := buildRequest(t).WithContext(
			context.WithValue(
				ctx,
				clientIDKey,
				exampleOAuth2Client.ClientID,
			),
		)

		actual := s.fetchOAuth2ClientIDFromRequest(req)
		assert.Equal(t, exampleOAuth2Client.ClientID, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		assert.Empty(t, s.fetchOAuth2ClientIDFromRequest(buildRequest(t)))
	})
}

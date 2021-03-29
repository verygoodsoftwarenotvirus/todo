package apiclients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"DecodeRequest",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&http.Request{}),
			mock.IsType(&types.APIClientCreationInput{}),
		).Return(nil)
		s.encoderDecoder = ed

		mh := &testutil.MockHTTPHandler{}
		mh.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{}))

		h := s.CreationInputMiddleware(mh)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		expected := fakes.BuildFakeAPIClientCreationInput()
		bs, err := json.Marshal(expected)
		require.NoError(t, err)
		req.Body = ioutil.NopCloser(bytes.NewReader(bs))

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"DecodeRequest",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&http.Request{}),
			mock.IsType(&types.APIClientCreationInput{}),
		).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
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

func TestService_fetchAPIClientFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		req := buildRequest(t).WithContext(
			context.WithValue(
				ctx,
				types.APIClientKey,
				exampleAPIClient,
			),
		)

		actual := s.fetchAPIClientFromRequest(req)
		assert.Equal(t, exampleAPIClient, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		assert.Nil(t, s.fetchAPIClientFromRequest(buildRequest(t)))
	})
}

func TestService_fetchAPIClientIDFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)
		exampleAPIClient := fakes.BuildFakeAPIClient()

		req := buildRequest(t).WithContext(
			context.WithValue(
				ctx,
				clientIDKey,
				exampleAPIClient.ClientID,
			),
		)

		actual := s.fetchAPIClientIDFromRequest(req)
		assert.Equal(t, exampleAPIClient.ClientID, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		assert.Empty(t, s.fetchAPIClientIDFromRequest(buildRequest(t)))
	})
}

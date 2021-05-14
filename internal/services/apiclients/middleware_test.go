package apiclients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.APIClientCreationInput{}),
		).Return(nil)
		s.encoderDecoder = encoderDecoder

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			mock.IsType(&http.Request{}))

		h := s.CreationInputMiddleware(mh)
		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()

		expected := fakes.BuildFakeAPIClientCreationInput()
		bs, err := json.Marshal(expected)
		require.NoError(t, err)
		req.Body = io.NopCloser(bytes.NewReader(bs))

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.APIClientCreationInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = encoderDecoder

		mh := &testutil.MockHTTPHandler{}
		h := s.CreationInputMiddleware(mh)
		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})
}

func TestService_fetchAPIClientFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		req := testutil.BuildTestRequest(t).WithContext(
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

		assert.Nil(t, s.fetchAPIClientFromRequest(testutil.BuildTestRequest(t)))
	})
}

func TestService_fetchAPIClientIDFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)
		exampleAPIClient := fakes.BuildFakeAPIClient()

		req := testutil.BuildTestRequest(t).WithContext(
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

		assert.Empty(t, s.fetchAPIClientIDFromRequest(testutil.BuildTestRequest(t)))
	})
}

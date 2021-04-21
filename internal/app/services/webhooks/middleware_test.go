package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookCreationInput()
		jsonBytes, err := json.Marshal(&exampleUpdateInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("bad input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := &types.WebhookCreationInput{}
		jsonBytes, err := json.Marshal(&exampleUpdateInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &testutil.MockHTTPHandler{}
		actual := s.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.WebhookCreationInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = encoderDecoder

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &testutil.MockHTTPHandler{}
		actual := s.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})
}

func TestService_UpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookUpdateInputFromWebhook(fakes.BuildFakeWebhook())
		jsonBytes, err := json.Marshal(&exampleUpdateInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.WebhookUpdateInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = encoderDecoder

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &testutil.MockHTTPHandler{}
		actual := s.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookUpdateInputFromWebhook(fakes.BuildFakeWebhook())
		exampleUpdateInput.Name = ""

		jsonBytes, err := json.Marshal(&exampleUpdateInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

package users

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_UserCreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakeUserCreationInput())

		actual := s.UserCreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.UserCreationInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = encoderDecoder

		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.UserCreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}

		exampleInput := fakes.BuildFakeUserCreationInput()
		exampleInput.Username = ""
		exampleInput.Password = ""

		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()
		req.Body = testutil.CreateBodyFromStruct(t, exampleInput)

		actual := s.UserCreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestService_PasswordUpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		req := testutil.BuildTestRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakePasswordUpdateInput())
		res := httptest.NewRecorder()

		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.PasswordUpdateInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = encoderDecoder

		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}

		exampleInput := fakes.BuildFakePasswordUpdateInput()
		exampleInput.NewPassword = ""

		req := testutil.BuildTestRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, exampleInput)
		res := httptest.NewRecorder()

		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestService_TOTPSecretVerificationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		req := testutil.BuildTestRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakeTOTPSecretVerificationInput())
		res := httptest.NewRecorder()

		actual := s.TOTPSecretVerificationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"invalid request content",
			http.StatusBadRequest,
		)
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.TOTPSecretVerificationInput{}),
		).Return(errors.New("blah"))
		s.encoderDecoder = encoderDecoder

		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.TOTPSecretVerificationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInput()
		exampleInput.TOTPToken = ""

		req := testutil.BuildTestRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, exampleInput)
		res := httptest.NewRecorder()

		actual := s.TOTPSecretVerificationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestService_TOTPSecretRefreshInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		req := testutil.BuildTestRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakeTOTPSecretRefreshInput())
		res := httptest.NewRecorder()

		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.TOTPSecretRefreshInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = encoderDecoder

		req := testutil.BuildTestRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		mh := &testutil.MockHTTPHandler{}

		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		exampleInput.TOTPToken = ""

		req := testutil.BuildTestRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, exampleInput)
		res := httptest.NewRecorder()

		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

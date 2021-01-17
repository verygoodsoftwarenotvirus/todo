package users

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_UserCreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		s.encoderDecoder = encoding.ProvideEncoderDecoder(noop.NewLogger())

		mh := &testutil.MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakeUserCreationInput())

		actual := s.UserCreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.RequestMatcher()), mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.UserCreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

func TestService_PasswordUpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		s.encoderDecoder = encoding.ProvideEncoderDecoder(noop.NewLogger())

		mh := &testutil.MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakePasswordUpdateInput())
		res := httptest.NewRecorder()

		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserCount", mock.MatchedBy(testutil.ContextMatcher()), mock.Anything).Return(uint64(123), nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.RequestMatcher()), mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

func TestService_TOTPSecretVerificationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		s.encoderDecoder = encoding.ProvideEncoderDecoder(noop.NewLogger())

		mh := &testutil.MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakeTOTPSecretVerificationInput())
		res := httptest.NewRecorder()

		actual := s.TOTPSecretVerificationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.RequestMatcher()), mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.TOTPSecretVerificationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

func TestService_TOTPSecretRefreshInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		s.encoderDecoder = encoding.ProvideEncoderDecoder(noop.NewLogger())

		mh := &testutil.MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		req.Body = testutil.CreateBodyFromStruct(t, fakes.BuildFakeTOTPSecretRefreshInput())
		res := httptest.NewRecorder()

		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.RequestMatcher()), mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &testutil.MockHTTPHandler{}
		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

package users

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

var _ http.Handler = (*MockHTTPHandler)(nil)

type MockHTTPHandler struct {
	mock.Mock
}

func (m *MockHTTPHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

func TestService_UserInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := &Service{
			logger: noop.NewLogger(),
		}

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.UserInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := &Service{
			logger: noop.NewLogger(),
		}

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &MockHTTPHandler{}
		actual := s.UserInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

func TestService_PasswordUpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := &Service{
			logger: noop.NewLogger(),
		}

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := &Service{
			logger: noop.NewLogger(),
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserCount", mock.Anything, mock.Anything).Return(uint64(123), nil)
		s.userDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &MockHTTPHandler{}
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
		s := &Service{
			logger: noop.NewLogger(),
		}

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.TOTPSecretVerificationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := &Service{
			logger: noop.NewLogger(),
		}

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &MockHTTPHandler{}
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
		s := &Service{
			logger: noop.NewLogger(),
		}

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := &Service{
			logger: noop.NewLogger(),
		}

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()

		mh := &MockHTTPHandler{}
		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

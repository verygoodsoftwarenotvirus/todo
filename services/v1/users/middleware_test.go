package users

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		s := &Service{
			logger: noop.ProvideNoopLogger(),
		}

		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).
			Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.UserInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		s := &Service{
			logger: noop.ProvideNoopLogger(),
		}

		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).
			Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.UserInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusBadRequest)
	})
}

func TestService_PasswordUpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := &Service{
			logger: noop.ProvideNoopLogger(),
		}

		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).
			Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		s := &Service{logger: noop.ProvideNoopLogger()}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserCount", mock.Anything, mock.Anything).
			Return(uint64(123), nil)
		s.database = mockDB

		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).
			Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.PasswordUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusBadRequest)
	})
}

func TestService_TOTPSecretRefreshInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := &Service{
			logger: noop.ProvideNoopLogger(),
		}

		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).
			Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		s := &Service{
			logger: noop.ProvideNoopLogger(),
		}

		ed := &mencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		mh := &MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).
			Return()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual := s.TOTPSecretRefreshInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusBadRequest)
	})
}

package items

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
		s := buildTestService()

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		mh := &mockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		res := httptest.NewRecorder()
		actual := s.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		s := buildTestService()

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		mh := &mockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		res := httptest.NewRecorder()
		actual := s.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
}

func TestService_UpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		mh := &mockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		res := httptest.NewRecorder()
		actual := s.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		s := buildTestService()

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		mh := &mockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		res := httptest.NewRecorder()
		actual := s.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
}

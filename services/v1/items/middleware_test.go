package items

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

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
		t.Parallel()
		s := buildTestService()
		s.encoderDecoder = &encoding.ServerEncoderDecoder{}

		exampleCreationInput := fakemodels.BuildFakeItemCreationInput()
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		mh := &mockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", strings.NewReader(string(jsonBytes)))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("bad input", func(t *testing.T) {
		t.Parallel()
		s := buildTestService()
		s.encoderDecoder = &encoding.ServerEncoderDecoder{}

		exampleCreationInput := &models.ItemCreationInput{}
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", strings.NewReader(string(jsonBytes)))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.CreationInputMiddleware(&mockHTTPHandler{})
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService()

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &mockHTTPHandler{}
		actual := s.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

func TestService_UpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService()
		s.encoderDecoder = &encoding.ServerEncoderDecoder{}

		exampleCreationInput := fakemodels.BuildFakeItemUpdateInputFromItem(fakemodels.BuildFakeItem())
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		mh := &mockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService()

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &mockHTTPHandler{}
		actual := s.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

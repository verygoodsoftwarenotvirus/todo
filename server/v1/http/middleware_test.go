package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequest(
		http.MethodGet,
		"https://verygoodsoftwarenotvirus.ru",
		nil,
	)

	require.NotNil(t, req)
	assert.NoError(t, err)

	return req
}

func Test_formatSpanNameForRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		req := buildRequest(t)
		req.Method = http.MethodPatch
		req.URL.Path = "/blah"

		expected := "PATCH /blah"
		actual := formatSpanNameForRequest(req)

		assert.Equal(t, expected, actual)
	})
}

func TestServer_loggingMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestServer()

		mh := &mockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res, req := httptest.NewRecorder(), buildRequest(t)
		s.loggingMiddleware(mh).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

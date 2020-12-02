package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(),
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
		t.Parallel()

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
		t.Parallel()
		s := buildTestServer()

		res, req := httptest.NewRecorder(), buildRequest(t)
		buildLoggingMiddleware(s.logger)(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {})).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

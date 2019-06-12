package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestService_StaticDir(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := &Service{logger: noop.ProvideNoopLogger()}
		exampleDir := "frontend/v1/public/"

		require.NoError(t, os.Chdir("/home/jeffrey/src/gitlab.com/verygoodsoftwarenotvirus/todo/"))

		hf, err := s.StaticDir(exampleDir)
		assert.NoError(t, err)
		assert.NotNil(t, hf)

		req, res := buildRequest(t), httptest.NewRecorder()
		req.URL.Path = "/http_routes_test.go"

		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with frontend routing path", func(t *testing.T) {
		s := &Service{logger: noop.ProvideNoopLogger()}
		exampleDir := "."

		hf, err := s.StaticDir(exampleDir)
		assert.NoError(t, err)
		assert.NotNil(t, hf)

		req, res := buildRequest(t), httptest.NewRecorder()
		req.URL.Path = "/login"

		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with frontend items routing path", func(t *testing.T) {
		s := &Service{logger: noop.ProvideNoopLogger()}
		exampleDir := "."

		hf, err := s.StaticDir(exampleDir)
		assert.NoError(t, err)
		assert.NotNil(t, hf)

		req, res := buildRequest(t), httptest.NewRecorder()
		req.URL.Path = "/items/9"

		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

}

func TestService_Routes(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		assert.NotNil(t, (&Service{}).Routes())
	})
}

package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
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

		cwd, err := os.Getwd()
		require.NoError(t, err)

		hf, err := s.StaticDir(cwd)
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

func TestService_buildStaticFileServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s := &Service{
			config: config.FrontendSettings{
				CacheStaticFiles: true,
			},
		}
		cwd, err := os.Getwd()
		require.NoError(t, err)

		actual, err := s.buildStaticFileServer(cwd)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

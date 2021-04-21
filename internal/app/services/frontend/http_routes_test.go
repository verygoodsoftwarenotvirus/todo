package frontend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_shouldRedirectToHome(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		for route := range validRoutes {
			assert.True(t, shouldRedirectToHome(route))
		}
	})

	T.Run("false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, shouldRedirectToHome("          blah blah          "))
	})
}

func Test_service_cacheFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildService(nil, Config{CacheStaticFiles: true})
		fs := afero.NewMemMapFs()

		assert.NoError(t, s.cacheFile(ctx, fs, "http_routes_test.go"))
	})

	T.Run("with error reading file", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildService(nil, Config{CacheStaticFiles: true})

		fs := afero.NewOsFs()

		assert.Error(t, s.cacheFile(ctx, fs, "this/file/does/not/exist/http_routes_test.go"))
	})
}

func Test_service_buildStaticFileServer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildService(nil, Config{CacheStaticFiles: true})

		cwd, err := os.Getwd()
		require.NoError(t, err)

		actual, err := s.buildStaticFileServer(ctx, cwd)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func Test_service_StaticDir(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		s := buildService(nil, Config{CacheStaticFiles: true, LogStaticFiles: true})

		ctx := context.Background()
		cwd, err := os.Getwd()
		require.NoError(t, err)

		hf, err := s.StaticDir(ctx, cwd)
		assert.NoError(t, err)
		assert.NotNil(t, hf)

		req, res := testutil.BuildTestRequest(t), httptest.NewRecorder()
		req.URL.Path = "/http_routes_test.go"
		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
	})

	T.Run("with frontend routing path", func(t *testing.T) {
		t.Parallel()
		s := buildService(nil, Config{CacheStaticFiles: true})

		ctx := context.Background()
		exampleDir := "."

		hf, err := s.StaticDir(ctx, exampleDir)
		assert.NoError(t, err)
		assert.NotNil(t, hf)

		req, res := testutil.BuildTestRequest(t), httptest.NewRecorder()
		req.URL.Path = "/auth/login"
		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
	})
}

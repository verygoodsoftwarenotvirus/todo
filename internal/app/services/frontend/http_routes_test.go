package frontend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_StaticDir(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		s := buildService(nil, Config{CacheStaticFiles: true})

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

func TestService_buildStaticFileServer(T *testing.T) {
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

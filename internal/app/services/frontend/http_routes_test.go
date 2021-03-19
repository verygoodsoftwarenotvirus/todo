package frontend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	ctx := context.Background()
	req, err := http.NewRequestWithContext(
		ctx,
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

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		s := &service{
			logger: logging.NewNonOperationalLogger(),
			tracer: tracing.NewTracer("test"),
		}

		cwd, err := os.Getwd()
		require.NoError(t, err)

		hf, err := s.StaticDir(cwd)
		assert.NoError(t, err)
		assert.NotNil(t, hf)

		req, res := buildRequest(t), httptest.NewRecorder()
		req.URL.Path = "/http_routes_test.go"
		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
	})

	T.Run("with frontend routing path", func(t *testing.T) {
		t.Parallel()
		s := &service{
			logger: logging.NewNonOperationalLogger(),
			tracer: tracing.NewTracer("test"),
		}

		exampleDir := "."

		hf, err := s.StaticDir(exampleDir)
		assert.NoError(t, err)
		assert.NotNil(t, hf)

		req, res := buildRequest(t), httptest.NewRecorder()
		req.URL.Path = "/auth/login"
		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
	})
}

func TestService_buildStaticFileServer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		s := &service{
			config: Config{
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

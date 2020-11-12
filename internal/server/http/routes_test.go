package httpserver

import (
	"testing"

	"github.com/go-chi/chi"
)

func TestServer_logRoutes(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		s := buildTestServer()

		s.router = chi.NewMux()

		s.logRoutes()
	})
}

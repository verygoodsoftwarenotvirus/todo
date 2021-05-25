package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, EnsureLogger(NewNoopLogger()))
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, EnsureLogger(nil))
	})
}

func TestBuildLoggingMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		middleware := BuildLoggingMiddleware(NewNoopLogger())

		assert.NotNil(t, middleware)

		hf := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {})

		req, res := httptest.NewRequest(http.MethodPost, "/nil", nil), httptest.NewRecorder()

		middleware(hf).ServeHTTP(res, req)
	})

	T.Run("with non-logged route", func(t *testing.T) {
		t.Parallel()

		middleware := BuildLoggingMiddleware(NewNoopLogger())

		assert.NotNil(t, middleware)

		hf := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {})

		if len(doNotLog) == 0 {
			t.SkipNow()
		}

		var route string
		for k := range doNotLog {
			route = k
			break
		}

		req, res := httptest.NewRequest(http.MethodPost, route, nil), httptest.NewRecorder()

		middleware(hf).ServeHTTP(res, req)
	})
}

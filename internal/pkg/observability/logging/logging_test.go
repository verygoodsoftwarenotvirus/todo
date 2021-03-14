package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BuildLoggingMiddleware(t *testing.T) {
	t.Run("normal operation", func(t *testing.T) {
		middleware := BuildLoggingMiddleware(NewNonOperationalLogger())

		assert.NotNil(t, middleware)

		hf := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {})

		req, res := httptest.NewRequest(http.MethodPost, "/nil", nil), httptest.NewRecorder()

		middleware(hf).ServeHTTP(res, req)
	})
}

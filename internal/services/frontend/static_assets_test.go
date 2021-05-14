package frontend

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_favicon(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/favicon.svg", nil)

		s.favicon(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

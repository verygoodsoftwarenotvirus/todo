package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildDefaultTransport(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		actual := buildDefaultTransport(0)
		assert.NotNil(t, actual)
	})
}

func Test_defaultRoundTripper_RoundTrip(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					res.WriteHeader(http.StatusOK)
				},
			),
		)
		ts.EnableHTTP2 = true

		transport := newDefaultRoundTripper(0)
		assert.NotNil(t, transport)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
		assert.NotNil(t, req)
		assert.NoError(t, err)

		_, err = transport.RoundTrip(req)
		assert.NoError(t, err)
	})
}

func Test_newDefaultRoundTripper(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		rt := newDefaultRoundTripper(0)
		assert.NotNil(t, rt)
	})
}

package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildDefaultTransport(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = buildDefaultTransport(0)
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
			))

		transport := newDefaultRoundTripper(0)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)

		require.NotNil(t, req)
		assert.NoError(t, err)

		_, err = transport.RoundTrip(req)
		assert.NoError(t, err)
	})
}

func Test_newDefaultRoundTripper(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = newDefaultRoundTripper(0)
	})
}

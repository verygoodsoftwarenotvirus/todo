package client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildDefaultTransport(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = buildDefaultTransport()
	})
}

func Test_defaultRoundTripper_RoundTrip(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		transport := newDefaultRoundTripper()

		req, err := http.NewRequest(http.MethodGet, "https://verygoodsoftwarenotvirus.ru", nil)

		require.NotNil(t, req)
		assert.NoError(t, err)

		_, err = transport.RoundTrip(req)
		assert.NoError(t, err)
	})
}

func Test_newDefaultRoundTripper(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_ = newDefaultRoundTripper()
	})
}

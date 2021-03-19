package httpserver

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"https://verygoodsoftwarenotvirus.ru",
		nil,
	)

	require.NotNil(t, req)
	assert.NoError(t, err)

	return req
}

func Test_formatSpanNameForRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		req := buildRequest(t)
		req.Method = http.MethodPatch
		req.URL.Path = "/blah"

		exampleOperation := "fart"

		expected := "PATCH /blah: fart"
		actual := formatSpanNameForRequest(exampleOperation, req)

		assert.Equal(t, expected, actual)
	})
}

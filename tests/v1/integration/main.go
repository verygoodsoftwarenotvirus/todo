package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	http2 "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
)

const (
	debug = true

	nonexistentID = 999999999
)

var (
	urlToUse   string
	todoClient *http2.V1Client
)

func checkValueAndError(t *testing.T, i interface{}, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, i)
}

func buildHTTPClient() *http.Client {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	return httpc
}

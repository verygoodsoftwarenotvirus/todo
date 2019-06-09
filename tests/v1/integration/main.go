package integration

import (
	"net/http"
	"testing"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"

	"github.com/stretchr/testify/require"
)

const (
	debug = true

	nonexistentID = 999999999
)

var (
	urlToUse   string
	todoClient *client.V1Client
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

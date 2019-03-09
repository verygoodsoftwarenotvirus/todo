package integration

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/go"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/stretchr/testify/require"
)

const (
	debug = true

	nonexistentID          = 999999999
	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

var (
	urlToUse   string
	logger     logging.Logger
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

func initializeClient(clientID, clientSecret string) *client.V1Client {
	uri, _ := url.Parse(urlToUse)
	c, err := client.NewClient(
		clientID,
		clientSecret,
		uri,
		zerolog.ProvideLogger(),
		buildHTTPClient(),
		tracing.ProvideNoopTracer(),
		debug,
	)
	if err != nil {
		panic(err)
	}
	return c
}

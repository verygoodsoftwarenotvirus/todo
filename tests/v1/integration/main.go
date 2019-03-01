package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/go"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/opentracing/opentracing-go"
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
	todoClient *client.V1Client
)

func buildSpanContext(operationName string) context.Context {
	tspan := opentracing.GlobalTracer().StartSpan(fmt.Sprintf("integration-tests-%s", operationName))
	return opentracing.ContextWithSpan(context.Background(), tspan)
}

func checkValueAndError(t *testing.T, i interface{}, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, i)
}

func initializeTracer() {
	tracer := tracing.ProvideTracer("integration-tests-client")
	opentracing.SetGlobalTracer(tracer)
}

func buildHTTPClient() *http.Client {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		// Timeout:   5 * time.Second,
	}

	return httpc
}

func initializeClient(clientID, clientSecret string) *client.V1Client {
	httpc := buildHTTPClient()

	u, _ := url.Parse(urlToUse)
	c, err := client.NewClient(
		clientID,
		clientSecret,
		u,
		zerolog.ProvideLogger(zerolog.ProvideZerologger()),
		httpc,
		opentracing.GlobalTracer(),
		debug,
	)
	if err != nil {
		panic(err)
	}
	return c
}

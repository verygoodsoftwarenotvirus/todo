package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/go"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
)

const (
	debug = true

	nonexistentID          = 999999999
	localTestInstanceURL   = "https://localhost"
	defaultTestInstanceURL = "https://todo-server"

	dockerPostgresAddress = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"
	localPostgresAddress  = "postgres://todo:hunter2@localhost:2345/todo?sslmode=disable"

	defaultSecret                   = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientID     = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientSecret = defaultSecret
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
	tracer, err := tracing.ProvideTracer("integration-tests-client")
	if err != nil {
		log.Fatal(err)
	}
	opentracing.SetGlobalTracer(tracer)
}

func buildHTTPClient() *http.Client {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
	httpc.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return httpc
}

func initializeClient(clientID, clientSecret string) {
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
	todoClient = c
}

func isUp(baseURI string) bool {
	uri := fmt.Sprintf("%s/_meta_/health", baseURI)

	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	httpc := buildHTTPClient()

	res, err := httpc.Do(req)
	if err != nil {
		return false
	}

	return res.StatusCode == http.StatusOK
}

func ensureServerIsUp(address string) {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !isUp(address) {
			log.Printf("waiting half a second before pinging again")
			time.Sleep(500 * time.Millisecond)
			numberOfAttempts++
			if numberOfAttempts >= maxAttempts {
				log.Fatalf("Maximum number of attempts made, something's gone awry")
			}
		} else {
			isDown = false
		}
	}
}

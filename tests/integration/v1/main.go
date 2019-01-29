package integration

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	config "github.com/uber/jaeger-client-go/config"
	jexpvar "github.com/uber/jaeger-lib/metrics/expvar"
)

const (
	debug                  = false
	nonexistentID          = 999999999
	localTestInstanceURL   = "https://localhost"
	defaultTestInstanceURL = "https://demo-server"

	defaultSecret                   = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientID     = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientSecret = defaultSecret
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

// provideJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func provideJaeger() (tracer opentracing.Tracer) {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal("cannot parse Jaeger env vars", err)
	}
	cfg.ServiceName = "integration-tests-client"
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	// TODO(ys) a quick hack to ensure random generators get different seeds, which are based on current time.
	metricsFactory := jexpvar.NewFactory(10).Namespace(cfg.ServiceName, nil)
	if tracer, _, err = cfg.NewTracer(
		// config.Logger(jaegerLogger),
		config.Metrics(metricsFactory),
		// config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	); err != nil {
		log.Fatalf("ERROR: cannot init Jaeger: %v\n", err)
	}

	return tracer
}

func initializeClient() {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
	httpc.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c, err := client.NewClient(
		urlToUse,
		defaultTestInstanceClientID,
		defaultTestInstanceClientSecret,
		logrus.New(),
		httpc,
		provideJaeger(),
		true,
	)
	if err != nil {
		panic(err)
	}
	todoClient = c
}

func ensureServerIsUp() {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !todoClient.IsUp() {
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

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}

	initializeClient()
	ensureServerIsUp()
	//testOAuth()
}

package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/db_bootstrap"

	_ "github.com/lib/pq" // importing for database initialization
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	debug = true

	nonexistentID          = 999999999
	localTestInstanceURL   = "https://localhost"
	defaultTestInstanceURL = "https://demo-server"

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

func initializeClient() {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
	httpc.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	tracer, err := tracing.ProvideTracer("integration-tests-client")
	if err != nil {
		log.Fatal(err)
	}
	opentracing.SetGlobalTracer(tracer)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	c, err := client.NewClient(
		urlToUse,
		defaultTestInstanceClientID,
		defaultTestInstanceClientSecret,
		logger,
		nil, // REPLACEME with actual logger
		httpc,
		tracer,
		debug,
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

	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		switch strings.ToLower(os.Getenv("DATABASE_TO_USE")) {
		case "postgres":
			db, err := postgres.ProvidePostgres(
				true,
				logrus.New(),
				zerolog.ProvideLogger(zerolog.ProvideZerologger()),
				opentracing.GlobalTracer(),
				dockerPostgresAddress,
			)
			if err != nil {
				log.Fatal(err)
			}

			if err := bootstrap.PreloadDatabase(db, ""); err != nil {
				log.Fatal(err)
			}
		}
	}

	// time.Sleep(10 * time.Minute)
	fmt.Println("Running tests")
}

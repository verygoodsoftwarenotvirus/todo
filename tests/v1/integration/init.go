package integration

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/zerolog"

	"github.com/icrowley/fake"
)

func init() {
	urlToUse = testutil.DetermineServiceURL()
	logger := zerolog.NewZeroLogger()

	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(urlToUse)

	fake.Seed(time.Now().UnixNano())

	u, err := testutil.CreateObligatoryUser(urlToUse, debug)
	if err != nil {
		logger.Fatal(err)
	}

	clientID, clientSecret, err := testutil.CreateObligatoryClient(urlToUse, *u)
	if err != nil {
		logger.Fatal(err)
	}

	todoClient = initializeClient(clientID, clientSecret)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func initializeClient(clientID, clientSecret string) *client.V1Client {
	uri, err := url.Parse(urlToUse)
	if err != nil {
		panic(err)
	}

	c, err := client.NewClient(
		context.Background(),
		clientID,
		clientSecret,
		uri,
		zerolog.NewZeroLogger(),
		buildHTTPClient(),
		debug,
	)
	if err != nil {
		panic(err)
	}
	return c
}

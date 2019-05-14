package load

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/http_client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/zerolog"

	"github.com/icrowley/fake"
)

const (
	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

var (
	debug                                   bool
	urlToUse, hcURL, clientID, clientSecret string
)

func buildHTTPClient() *http.Client {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	return httpc
}

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}
	logger := zerolog.NewZeroLogger()

	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(urlToUse)

	fake.Seed(time.Now().UnixNano())

	u, err := testutil.CreateObligatoryUser(urlToUse, debug)
	if err != nil {
		logger.Fatal(err)
	}

	clientID, clientSecret, err = testutil.CreateObligatoryClient(urlToUse, *u)
	if err != nil {
		logger.Fatal(err)
	}

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func initializeClient(clientID, clientSecret string) *client.V1Client {
	uri, _ := url.Parse(urlToUse)
	c, err := client.NewClient(
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

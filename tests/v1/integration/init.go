package integration

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/zerolog"

	"github.com/icrowley/fake"
)

const (
	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		ta := os.Getenv("TARGET_ADDRESS")
		if ta == "" {
			urlToUse = defaultTestInstanceURL
		} else {
			u, err := url.Parse(ta)
			if err != nil {
				panic(err)
			}
			urlToUse = u.String()
		}

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

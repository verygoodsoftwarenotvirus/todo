package httpclient

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	"github.com/icrowley/fake"
)

const (
	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

func buildURL(parts ...string) string {
	tu, _ := url.Parse(urlToUse)
	u, _ := url.Parse(strings.Join(parts, "/"))
	return tu.ResolveReference(u).String()
}

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}
	logger := zerolog.ProvideLogger()

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

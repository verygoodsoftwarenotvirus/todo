package frontend

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	fake "github.com/brianvoe/gofakeit"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/zerolog"
)

var urlToUse string

const (
	seleniumHubAddr = "http://selenium-hub:4444/wd/hub"
)

func init() {
	fake.Seed(time.Now().UnixNano())

	urlToUse = testutil.DetermineServiceURL()

	logger := zerolog.NewZeroLogger()
	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(urlToUse)

	fake.Seed(time.Now().UnixNano())

	// NOTE: this is sad, but also the only thing that consistently works
	// see above for my vain attempts at a real solution to this problem
	time.Sleep(10 * time.Second)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}
